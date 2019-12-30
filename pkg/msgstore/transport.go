package msgstore

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

// Create http.Handler instance for servicing message store requests
func MakeHandler(service Service, logger kitlog.Logger) http.Handler {

	r := mux.NewRouter()

	opts := []kithttp.ServerOption{
		kithttp.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
		kithttp.ServerErrorEncoder(encodeError),
	}

	createMessageHandler := kithttp.NewServer(
		makeCreateMessageEndpoint(service),
		decodeMessageCreateRequest,
		encodeResponse,
		opts...,
	)

	r.Handle("/messages", createMessageHandler).Methods("POST")

	createReplyHandler := kithttp.NewServer(
		makeCreateReplyEndpoint(service),
		decodeReplyCreateRequest,
		encodeResponse,
		opts...,
	)

	r.Handle("/messages/{msgid}/replies", createReplyHandler).Methods("POST")

	getMessageHandler := kithttp.NewServer(
		makeQueryMessageForIdEndpoint(service),
		decodeMessageForIdQueryRequest,
		encodeResponse,
		opts...,
	)

	r.Handle("/messages/{msgid}", getMessageHandler).Methods("GET")

	getUserMessagesHandler := kithttp.NewServer(
		makeQueryMessagesForUserEndpoint(service),
		decodeMessagesForUserQueryRequest,
		encodeResponse,
		opts...,
	)

	r.Handle("/users/{userid}/mailbox", getUserMessagesHandler).Methods("GET")

	return middleware.NewHTTPInterceptor(r, logger)
}

type receiver struct {
	Groupname string `json:"groupname,omitempty"`
	Username  string `json:"username,omitempty"`
}

type message struct {
	Id        string   `json:"id"`
	Re        string   `json:"re,omitempty"`
	Sender    string   `json:"sender"`
	Recipient receiver `json:"recipient"`
	Subject   string   `json:"subject"`
	Body      string   `json:"body"`
	Timestamp string   `json:"sentAt"`
}

type messageCreateRequest struct {
	Content message
}

type messageCreateResponse struct {
	Id string
}

func (m *messageCreateResponse) StatusCode() int {
	return http.StatusCreated
}

type replyCreateRequest struct {
	Content message
}

type replyCreateResponse struct {
	Id string
}

func (m *replyCreateResponse) StatusCode() int {
	return http.StatusCreated
}

type messageForIdQueryRequest struct {
	Id string
}

type messageForIdQueryResponse struct {
	Content message
}

func (m *messageForIdQueryResponse) StatusCode() int {
	return http.StatusOK
}

type messagesForUserQueryRequest struct {
	Username string
}

type messagesForUserQueryResponse struct {
	Content []message
}

func (m *messagesForUserQueryResponse) StatusCode() int {
	return http.StatusOK
}

type replyQueryRequest struct {
	Id string
}

type replyQueryResponse struct {
	Content []message
}

func (m *replyQueryResponse) StatusCode() int {
	return http.StatusOK
}

func makeCreateMessageEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(messageCreateRequest)
		msgid, err := s.StoreMessage(ctx, req.Content)
		return &messageCreateResponse{msgid}, err
	}
}

func makeCreateReplyEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(replyCreateRequest)
		msgid, err := s.StoreMessage(ctx, req.Content)
		return &replyCreateResponse{msgid}, err
	}
}

func makeQueryMessageForIdEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(messageForIdQueryRequest)
		msg, err := s.GetMessage(ctx, req.Id)
		return &messageForIdQueryResponse{msg}, err
	}
}

func makeQueryMessagesForUserEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(messagesForUserQueryRequest)
		msgs, err := s.GetMessages(ctx, req.Username)
		return &messagesForUserQueryResponse{msgs}, err
	}
}

func makeQueryReplyEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(replyQueryRequest)
		msgs, err := s.GetMessages(ctx, req.Id)
		return &replyQueryResponse{msgs}, err
	}
}

func decodeMessageCreateRequest(_ context.Context, r *http.Request) (interface{}, error) {

	mcRequest := messageCreateRequest{}

	if err := json.NewDecoder(r.Body).Decode(&mcRequest.Content); err != nil {
		return nil, err
	}

	msg := &mcRequest.Content

	if msg.Recipient.Groupname == "" && msg.Recipient.Username == "" {
		return nil, ErrBadRequest
	}

	if msg.Sender == "" || msg.Subject == "" || msg.Body == "" {
		return nil, ErrBadRequest
	}

	return mcRequest, nil

}

func decodeReplyCreateRequest(ctx context.Context, r *http.Request) (interface{}, error) {

	rcRequest := replyCreateRequest{}

	if err := json.NewDecoder(r.Body).Decode(&rcRequest.Content); err != nil {
		return nil, err
	}

	msg := &rcRequest.Content

	if msg.Sender == "" || msg.Subject == "" || msg.Body == "" {
		return nil, ErrBadRequest
	}

	if len(msg.Recipient.Groupname) > 0 || len(msg.Recipient.Username) > 0 {
		return nil, ErrBadRequest
	}

	msg.Re = mux.Vars(r)["msgid"]

	return rcRequest, nil

}

func decodeMessageForIdQueryRequest(_ context.Context, r *http.Request) (interface{}, error) {
	msgid := mux.Vars(r)["msgid"]
	mqRequest := messageForIdQueryRequest{msgid}
	return mqRequest, nil
}

func decodeMessagesForUserQueryRequest(_ context.Context, r *http.Request) (interface{}, error) {
	userid := mux.Vars(r)["userid"]
	mqRequest := messagesForUserQueryRequest{userid}
	return mqRequest, nil
}

func decodeReplyQueryRequest(_ context.Context, r *http.Request) (interface{}, error) {
	msgid := mux.Vars(r)["msgid"]
	rqRequest := replyQueryRequest{msgid}
	return rqRequest, nil
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
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}
