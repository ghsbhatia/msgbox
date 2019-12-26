package useradmin

import (
	"context"
	"time"

	"github.com/ghsbhatia/msgbox/pkg/ctxlog"
	"github.com/go-kit/kit/log"
)

// Service interface for user admin functions
type Service interface {
	// register a new user
	RegisterUser(context.Context, string) error
	// get user for a given name
	GetUser(context.Context, string) (string, error)
	// register a new group
	RegisterGroup(context.Context, string, []string) error
	// get users for a groups
	GetGroupUsers(context.Context, string) ([]string, error)
}

// Create a new service instance with a given user repository
func NewService(repository UserRepository) Service {
	return &service{repository}
}

type service struct {
	repository UserRepository
}

func (s *service) RegisterUser(ctx context.Context, username string) (err error) {
	defer func(begin time.Time) {
		svcLogger := log.With(ctxlog.Logger(ctx), "component", "service")
		svcLogger.Log(
			"method", "register user",
			"username", username,
			"took", time.Since(begin),
			"err", err,
		)
	}(time.Now())
	{
		exists, err := s.repository.FindUser(ctx, username)
		if exists {
			return ErrUserExists
		}
		if err != nil {
			return err
		}
	}
	err = s.repository.StoreUser(ctx, username)
	return err
}

func (s *service) GetUser(ctx context.Context, username string) (user string, err error) {
	defer func(begin time.Time) {
		svcLogger := log.With(ctxlog.Logger(ctx), "component", "service")
		svcLogger.Log(
			"method", "get user",
			"username", username,
			"took", time.Since(begin),
			"err", err,
		)
	}(time.Now())
	{
		var exists bool
		exists, err = s.repository.FindUser(ctx, username)
		if exists {
			user = username
			return user, nil
		}
		return "", ErrUserNotFound
	}
}

func (s *service) RegisterGroup(ctx context.Context, groupname string, usernames []string) (err error) {
	defer func(begin time.Time) {
		svcLogger := log.With(ctxlog.Logger(ctx), "component", "service")
		svcLogger.Log(
			"method", "register group",
			"groupname", groupname,
			"took", time.Since(begin),
			"err", err,
		)
	}(time.Now())
	{
		if usernamesEmpty(usernames) {
			return ErrGroupEmpty
		}
		exists, err := s.repository.FindGroup(ctx, groupname)
		if exists {
			return ErrGroupExists
		}
		if err != nil {
			return err
		}
	}
	err = s.repository.StoreGroup(ctx, groupname, usernames)
	return err
}

func (s *service) GetGroupUsers(ctx context.Context, groupname string) (users []string, err error) {
	defer func(begin time.Time) {
		svcLogger := log.With(ctxlog.Logger(ctx), "component", "service")
		svcLogger.Log(
			"method", "get group users",
			"groupname", groupname,
			"took", time.Since(begin),
			"err", err,
		)
	}(time.Now())
	var exists bool
	exists, err = s.repository.FindGroup(ctx, groupname)
	if !exists {
		return nil, ErrGroupNotFound
	}
	if err != nil {
		return nil, err
	}
	users, err = s.repository.FetchGroupUsers(ctx, groupname)
	return users, err
}

func usernamesEmpty(names []string) bool {
	return len(names) == 0
}
