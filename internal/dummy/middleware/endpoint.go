package middleware

import (
	"context"
	"errors"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-playground/validator/v10"

	"ws-dummy-go/internal/dummy"
	"ws-dummy-go/internal/dummy/domain"
)

var (
	validate = validator.New(validator.WithRequiredStructEnabled())
)

func MakeCreateUserEndpoint(svc dummy.UserService) endpoint.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		request, ok := req.(createUserRequest)
		if !ok {
			return nil, NewNotImplementedError()
		}
		if err := validate.Struct(request); err != nil {
			return nil, NewValidationError(err.Error())
		}
		id, err := svc.CreateUser(ctx, request.Name)
		if err != nil {
			var e *domain.NotFoundError
			if errors.As(err, &e) {
				return nil, NewNotFoundError(e.Error())
			}
			return nil, NewInternalServerError()
		}
		return createUserResponse{UserID: string(id)}, nil
	}
}

func MakeUpdateUserEndpoint(svc dummy.UserService) endpoint.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		request, ok := req.(updateUserRequest)
		if !ok {
			return nil, NewNotImplementedError()
		}
		if err := validate.Struct(request); err != nil {
			return nil, NewValidationError(err.Error())
		}
		userID := domain.UserID(request.UserID)

		if err := svc.UpdateUser(ctx, userID, request.Name); err != nil {
			var e *domain.NotFoundError
			if errors.As(err, &e) {
				return nil, NewNotFoundError(e.Error())
			}
			return nil, NewInternalServerError()
		}
		return updateUserResponse{}, nil
	}
}
