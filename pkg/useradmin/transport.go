package useradmin

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/go-kit/kit/endpoint"
	kitlog "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/transport"
	kithttp "github.com/go-kit/kit/transport/http"

	"github.com/ghsbhatia/msgbox/pkg/middleware"
)

// Create http.Handler instance for servicing useradmin requests
func MakeHandler(service Service, logger kitlog.Logger) http.Handler {

	r := mux.NewRouter()

	opts := []kithttp.ServerOption{
		kithttp.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
		kithttp.ServerErrorEncoder(encodeError),
	}

	userRegistrationHandler := kithttp.NewServer(
		makeUserRegistrationEndpoint(service),
		decodeUserRegistrationRequest,
		encodeResponse,
		opts...,
	)

	r.Handle("/users", userRegistrationHandler).Methods("POST")

	userQueryHandler := kithttp.NewServer(
		makeUserQueryEndpoint(service),
		decodeUserQueryRequest,
		encodeResponse,
		opts...,
	)

	r.Handle("/users/{userid}", userQueryHandler).Methods("GET")

	groupRegistrationHandler := kithttp.NewServer(
		makeGroupRegistrationEndpoint(service),
		decodeGroupRegistrationRequest,
		encodeResponse,
		opts...,
	)

	r.Handle("/groups", groupRegistrationHandler).Methods("POST")

	groupQueryHandler := kithttp.NewServer(
		makeGroupQueryEndpoint(service),
		decodeGroupQueryRequest,
		encodeResponse,
		opts...,
	)

	r.Handle("/groups/{groupid}", groupQueryHandler).Methods("GET")

	return middleware.NewHTTPInterceptor(r, logger)
}

type userRegistrationRequest struct {
	Username string `json:"username"`
}

type userRegistrationResponse struct {
	Id string `json:"id"`
}

func (m *userRegistrationResponse) StatusCode() int {
	return http.StatusCreated
}

type userQueryRequest struct {
	Username string
}

type userQueryResponse struct {
	Id string `json:"id"`
}

func (m *userQueryResponse) StatusCode() int {
	return http.StatusOK
}

type groupRegistrationRequest struct {
	Groupname string   `json:"groupname"`
	Usernames []string `json:"usernames"`
}

type groupRegistrationResponse struct {
	Id string `json:"id"`
}

func (m *groupRegistrationResponse) StatusCode() int {
	return http.StatusCreated
}

type groupQueryRequest struct {
	Groupname string
}

type groupQueryResponse struct {
	Groupname string   `json:"groupname"`
	Usernames []string `json:"usernames"`
}

func (m *groupQueryResponse) StatusCode() int {
	return http.StatusOK
}

func makeUserRegistrationEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(userRegistrationRequest)
		id, err := s.RegisterUser(ctx, req.Username)
		if err != nil {
			return nil, err
		}
		return &userRegistrationResponse{id}, err
	}
}

func makeUserQueryEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(userQueryRequest)
		user, err := s.GetUser(ctx, req.Username)
		if err != nil {
			return nil, err
		}
		return &userQueryResponse{user}, err
	}
}

func makeGroupRegistrationEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(groupRegistrationRequest)
		id, err := s.RegisterGroup(ctx, req.Groupname, req.Usernames)
		if err != nil {
			return nil, err
		}
		return &groupRegistrationResponse{id}, err
	}
}

func makeGroupQueryEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(groupQueryRequest)
		groupname := req.Groupname
		users, err := s.GetGroupUsers(ctx, groupname)
		return &groupQueryResponse{Groupname: groupname, Usernames: users}, err
	}
}

func decodeUserRegistrationRequest(_ context.Context, r *http.Request) (interface{}, error) {

	var body userRegistrationRequest

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	}

	if body.Username == "" {
		return nil, ErrBadRequest
	}

	urRequest := userRegistrationRequest{body.Username}

	return urRequest, nil

}

func decodeUserQueryRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	username := vars["userid"]
	uqRequest := userQueryRequest{username}
	return uqRequest, nil
}

func decodeGroupRegistrationRequest(_ context.Context, r *http.Request) (interface{}, error) {

	var body groupRegistrationRequest

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	}

	if body.Groupname == "" || len(body.Usernames) == 0 {
		return nil, ErrBadRequest
	}

	grRequest := groupRegistrationRequest{body.Groupname, body.Usernames}

	return grRequest, nil

}

func decodeGroupQueryRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	groupname := vars["groupid"]
	gqRequest := groupQueryRequest{groupname}
	return gqRequest, nil
}

// encode response
func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		encodeError(ctx, e.error(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	code := http.StatusOK
	if sc, ok := response.(kithttp.StatusCoder); ok {
		code = sc.StatusCode()
	}
	w.WriteHeader(code)
	if code == http.StatusNoContent {
		return nil
	}
	return json.NewEncoder(w).Encode(response)
}

type errorer interface {
	error() error
}

// encode request execution error
func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	switch err {
	case ErrUserExists:
		w.WriteHeader(http.StatusConflict)
	case ErrGroupExists:
		w.WriteHeader(http.StatusConflict)
	case ErrUserNotFound:
		w.WriteHeader(http.StatusNotFound)
	case ErrGroupNotFound:
		w.WriteHeader(http.StatusNotFound)
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}
