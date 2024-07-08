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
		id, err := svc.Create(req.Name)
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

func MakeLoggingMiddleware(logger log.Logger, mode string) DecodingMiddleware {
	return func(next httptransport.DecodeRequestFunc) httptransport.DecodeRequestFunc {
		return func(ctx context.Context, r *http.Request) (interface{}, error) {
			body := []byte("hidden")

			if mode == "debug" {
				var err error
				body, err = httputil.DumpRequest(r, true)
				if err != nil {
					return nil, fmt.Errorf("dumping request: %w", err)
				}
			}
			logger.Log(
				"msg", "got request", "method", r.Method, "URL", r.URL, "len", r.ContentLength, "body", body,
			)
			return next(ctx, r)
		}
	}
}

func EncodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	return json.NewEncoder(w).Encode(response)
}

type createUserRequest struct {
	Name string `json:"name"`
}

type createUserResponse struct {
	ID  string `json:"id"`
	Err string `json:"err,omitempty"`
}
