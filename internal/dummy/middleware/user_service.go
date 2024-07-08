package middleware

import (
	"ws-dummy-go/internal/dummy"
)

// UserServiceMiddleware is a chainable behavior modifier for UserService.
type UserServiceMiddleware func(dummy.UserService) dummy.UserService
