package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"runtime/debug"

	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/go-kit/log"
	"github.com/go-playground/validator/v10"

	"ws-dummy-go/internal/dummy"
)

var (
	validate = validator.New(validator.WithRequiredStructEnabled())
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
	if err := validate.Struct(request); err != nil {
		return nil, NewValidationError(err.Error())
	}
	return request, nil
}

func MakeCreateUserEndpoint(svc dummy.UserService) endpoint.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		request, ok := req.(createUserRequest)
		if !ok {
			return createUserResponse{}, nil // TODO: 500
		}
		id, err := svc.CreateUser(ctx, request.Name)
		if err != nil {
			var e *dummy.NotFoundError
			if errors.As(err, &e) {
				return nil, NewNotFoundError(e.Error())
			}
			return nil, NewInternalServerError()
		}
		return createUserResponse{UserID: string(id)}, nil
	}
}

func ErrorEncoder(
	srf httptransport.ServerResponseFunc, ee httptransport.ErrorEncoder,
) httptransport.ErrorEncoder {
	return func(ctx context.Context, err error, w http.ResponseWriter) {
		srf(ctx, w)
		ee(ctx, err, w)
	}
}
