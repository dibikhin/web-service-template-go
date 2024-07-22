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

func (mw logmw) CreateUser(ctx context.Context, name string) (output domain.UserID, err error) {
	defer func(begin time.Time) {
		mw.logger.Log(
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
