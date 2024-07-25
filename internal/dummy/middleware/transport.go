package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"runtime/debug"

	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/go-kit/log"
)

type (
	DecodingMiddleware func(httptransport.DecodeRequestFunc) httptransport.DecodeRequestFunc
)

type createUserRequest struct {
	Name string `json:"name" validate:"required"`
}

type createUserResponse struct {
	UserID string `json:"userId"`
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

func DecodingRecovery(logger log.Logger) DecodingMiddleware {
	return func(next httptransport.DecodeRequestFunc) httptransport.DecodeRequestFunc {
		return func(ctx context.Context, req *http.Request) (v interface{}, e error) {
			defer func() {
				if err := recover(); err != nil {
					logger.Log("msg", "panic recovered", "err", err, "stack", string(debug.Stack()))
					v = nil
					e = NewValidationError("request validation failed")
				}
			}()
			return next(ctx, req)
		}
	}
}

func DecodeCreateUserRequest(_ context.Context, req *http.Request) (interface{}, error) {
	if req.ContentLength == 0 {
		return nil, NewValidationError("empty request")
	}
	var request createUserRequest
	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		// TODO: log error
		return nil, NewValidationError("cannot decode request")
	}
	return request, nil
}

func ErrorEncoder() httptransport.ErrorEncoder {
	return func(ctx context.Context, err error, w http.ResponseWriter) {
		SetRequestID(ctx, w)
		httptransport.DefaultErrorEncoder(ctx, err, w)
	}
}
