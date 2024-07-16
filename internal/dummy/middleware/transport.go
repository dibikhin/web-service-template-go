package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"runtime/debug"

	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/go-kit/log"
	"github.com/rs/xid"

	"ws-dummy-go/internal/dummy"
)

const (
	RequestIDHeader RequestIDType = "X-Request-ID"
)

type (
	RequestIDType      string
	DecodingMiddleware func(httptransport.DecodeRequestFunc) httptransport.DecodeRequestFunc
)

type createUserRequest struct {
	Name string `json:"name"`
}

type createUserResponse struct {
	UserID string `json:"userId"`
	Err    string `json:"err,omitempty"`
}

func Recovery(logger log.Logger) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, req interface{}) (v interface{}, e error) {
			defer func() {
				if err := recover(); err != nil {
					logger.Log("msg", "panic recovered", "err", err, "stack", string(debug.Stack()))
					v = nil
					e = NewInternalServerError()
				}
			}()
			return next(ctx, req)
		}
	}
}

func MakeCreateUserEndpoint(svc dummy.UserService) endpoint.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		request, ok := req.(createUserRequest)
		if !ok {
			return createUserResponse{"", "invalid request"}, nil
		}
		id, err := svc.CreateUser(ctx, request.Name)
		if err != nil {
			return createUserResponse{"", err.Error()}, nil
		}
		return createUserResponse{string(id), ""}, nil
	}
}

func DecodeCreateUserRequest(_ context.Context, req *http.Request) (interface{}, error) {
	var request createUserRequest

	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		return nil, fmt.Errorf("decoding request: %w", err)
	}
	return request, nil
}

func DecodingRecovery(logger log.Logger) DecodingMiddleware {
	return func(next httptransport.DecodeRequestFunc) httptransport.DecodeRequestFunc {
		return func(ctx context.Context, req *http.Request) (v interface{}, e error) {
			defer func() {
				if err := recover(); err != nil {
					logger.Log("msg", "panic recovered", "err", err, "stack", string(debug.Stack()))
					v = nil
					e = NewValidationError("invalid request")
				}
			}()
			return next(ctx, req)
		}
	}
}

func RequestID() DecodingMiddleware {
	return func(next httptransport.DecodeRequestFunc) httptransport.DecodeRequestFunc {
		return func(ctx context.Context, req *http.Request) (interface{}, error) {

			reqID := req.Header.Get(string(RequestIDHeader))
			if reqID == "" {
				reqID = xid.New().String()
			}
			ctx = context.WithValue(ctx, RequestIDHeader, reqID)
			return next(ctx, req)
		}
	}
}

func Logging(logger log.Logger, mode string) DecodingMiddleware {
	return func(next httptransport.DecodeRequestFunc) httptransport.DecodeRequestFunc {
		return func(ctx context.Context, req *http.Request) (interface{}, error) {
			body := []byte("hidden")

			reqID := ctx.Value(RequestIDHeader).(string)

			if mode == "debug" {
				var err error
				body, err = httputil.DumpRequest(req, true)
				if err != nil {
					return nil, fmt.Errorf("dumping request: %w", err)
				}
			}
			logger.Log(
				"msg", "got request", "method", req.Method, "URL", req.URL, "len", req.ContentLength,
				"reqID", reqID, "body", body,
			)
			return next(ctx, req)
		}
	}
}
