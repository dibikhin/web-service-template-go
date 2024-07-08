// Code generated by mockery v2.43.2. DO NOT EDIT.

package mocks

import (
	context "context"
	domain "ws-dummy-go/internal/dummy/domain"

	mock "github.com/stretchr/testify/mock"
)

// UsersSQLRepo is an autogenerated mock type for the UsersSQLRepo type
type UsersSQLRepo struct {
	mock.Mock
}

type UsersSQLRepo_Expecter struct {
	mock *mock.Mock
}

func (_m *UsersSQLRepo) EXPECT() *UsersSQLRepo_Expecter {
	return &UsersSQLRepo_Expecter{mock: &_m.Mock}
}

// Insert provides a mock function with given fields: ctx, name
func (_m *UsersSQLRepo) Insert(ctx context.Context, name string) (domain.UserID, error) {
	ret := _m.Called(ctx, name)

	if len(ret) == 0 {
		panic("no return value specified for Insert")
	}

	var r0 domain.UserID
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (domain.UserID, error)); ok {
		return rf(ctx, name)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) domain.UserID); ok {
		r0 = rf(ctx, name)
	} else {
		r0 = ret.Get(0).(domain.UserID)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UsersSQLRepo_Insert_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Insert'
type UsersSQLRepo_Insert_Call struct {
	*mock.Call
}

// Insert is a helper method to define mock.On call
//   - ctx context.Context
//   - name string
func (_e *UsersSQLRepo_Expecter) Insert(ctx interface{}, name interface{}) *UsersSQLRepo_Insert_Call {
	return &UsersSQLRepo_Insert_Call{Call: _e.mock.On("Insert", ctx, name)}
}

func (_c *UsersSQLRepo_Insert_Call) Run(run func(ctx context.Context, name string)) *UsersSQLRepo_Insert_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *UsersSQLRepo_Insert_Call) Return(_a0 domain.UserID, _a1 error) *UsersSQLRepo_Insert_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *UsersSQLRepo_Insert_Call) RunAndReturn(run func(context.Context, string) (domain.UserID, error)) *UsersSQLRepo_Insert_Call {
	_c.Call.Return(run)
	return _c
}

// NewUsersSQLRepo creates a new instance of UsersSQLRepo. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewUsersSQLRepo(t interface {
	mock.TestingT
	Cleanup(func())
}) *UsersSQLRepo {
	mock := &UsersSQLRepo{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
