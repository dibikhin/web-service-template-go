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

type DecodingMiddleware func(httptransport.DecodeRequestFunc) httptransport.DecodeRequestFunc

func RecoveryMiddleware(logger log.Logger) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			defer func() {
				if err := recover(); err != nil {
					logger.Log("msg", "panic", "err", err, "stack", string(debug.Stack()))
				}
			}()
			return next(ctx, request)
		}
	}
}

func MakeCreateUserEndpoint(svc dummy.UserService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(createUserRequest)
		if !ok {
			return createUserResponse{"", "invalid request"}, nil
		}
		id, err := svc.CreateUser(ctx, req.Name)
		if err != nil {
			return createUserResponse{"", err.Error()}, nil
		}
		return createUserResponse{string(id), ""}, nil
	}
}

func DecodeCreateUserRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var request createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, fmt.Errorf("decoding request: %w", err)
	}
	return request, nil
}

type RequestIDType string

const RequestID RequestIDType = "RequestID"

func RequestIDMiddleware() DecodingMiddleware {
	return func(next httptransport.DecodeRequestFunc) httptransport.DecodeRequestFunc {
		return func(ctx context.Context, r *http.Request) (interface{}, error) {

			reqID := r.Header.Get("X-Request-ID")
			if reqID == "" {
				reqID = xid.New().String()
			}
			ctx = context.WithValue(ctx, RequestID, reqID)
			return next(ctx, r)
		}
	}
}

func MakeLoggingMiddleware(logger log.Logger, mode string) DecodingMiddleware {
	return func(next httptransport.DecodeRequestFunc) httptransport.DecodeRequestFunc {
		return func(ctx context.Context, req *http.Request) (interface{}, error) {
			body := []byte("hidden")

			reqID := ctx.Value(RequestID).(string)

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

type createUserRequest struct {
	Name string `json:"name"`
}

type createUserResponse struct {
	ID  string `json:"id"`
	Err string `json:"err,omitempty"`
}
