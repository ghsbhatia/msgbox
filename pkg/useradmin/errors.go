package useradmin

import (
	"errors"
)

var ErrBadRequest = errors.New("invalid request")
var ErrUserExists = errors.New("user with the same username already registered")
var ErrUserNotFound = errors.New("user not found")
var ErrGroupExists = errors.New("group with the same groupname already registered")
var ErrGroupNotFound = errors.New("group not found")
var ErrGroupEmpty = errors.New("group has no users")
