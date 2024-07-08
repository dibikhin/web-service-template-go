package middleware

import (
	"context"
	"time"

	"github.com/go-kit/log"

	"ws-dummy-go/internal/dummy"
	"ws-dummy-go/internal/dummy/domain"
)

func NewLoggingMiddleware(logger log.Logger) UserServiceMiddleware {
	return func(next dummy.UserService) dummy.UserService {
		return logmw{logger, next}
	}
}

type logmw struct {
	logger log.Logger
	dummy.UserService
}

func (mw logmw) CreateUser(ctx context.Context, name string) (output domain.UserID, err error) {
	defer func(begin time.Time) {
		_ = mw.logger.Log(
			"method", "CreateUser",
			"input", name,
			"output", output,
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	output, err = mw.UserService.CreateUser(ctx, name)
	return
}
