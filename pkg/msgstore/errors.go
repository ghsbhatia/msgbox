package msgstore

import (
	"errors"
)

var ErrBadRequest = errors.New("invalid request")
var ErrMsgNotFound = errors.New("message not found")
var ErrUserNotFound = errors.New("user not found")
var ErrGroupNotFound = errors.New("group not found")
var ErrSystemError = errors.New("system error")
