package middleware

import (
	"context"
	"net/http"
	"net/http/httputil"
	"time"

	httptransport "github.com/go-kit/kit/transport/http"
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

func (mw logmw) CreateUser(ctx context.Context, name string) (id domain.UserID, err error) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"method", "CreateUser",
			"name", name,
			"id", id,
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	id, err = mw.UserService.CreateUser(ctx, name)
	return
}

func (mw logmw) UpdateUser(ctx context.Context, id domain.UserID, name string) (err error) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"method", "UpdateUser",
			"id", id,
			"name", name,
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	err = mw.UserService.UpdateUser(ctx, id, name)
	return
}

func RequestLogging(logger log.Logger, mode string) httptransport.RequestFunc {
	return func(ctx context.Context, req *http.Request) context.Context {
		rawRequest := []byte("hidden")

		if mode == "debug" {
			var err error
			rawRequest, err = httputil.DumpRequest(req, true)
			if err != nil {
				logger.Log("msg", "dumping request", "err", err)
				return ctx
			}
		}
		reqID := ctx.Value(requestIDHeader).(string)
		logger.Log(
			"msg", "request", "method", req.Method, "url", req.URL, "len", req.ContentLength,
			"reqID", reqID, "rawRequest", rawRequest,
		)
		return ctx
	}
}
