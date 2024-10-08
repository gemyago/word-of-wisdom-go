// Code generated by mockery. DO NOT EDIT.

//go:build !release

package server

import (
	context "context"
	services "word-of-wisdom-go/internal/services"

	mock "github.com/stretchr/testify/mock"
)

// mockCommandHandler is an autogenerated mock type for the commandHandler type
type mockCommandHandler struct {
	mock.Mock
}

type mockCommandHandler_Expecter struct {
	mock *mock.Mock
}

func (_m *mockCommandHandler) EXPECT() *mockCommandHandler_Expecter {
	return &mockCommandHandler_Expecter{mock: &_m.Mock}
}

// Handle provides a mock function with given fields: ctx, session
func (_m *mockCommandHandler) Handle(ctx context.Context, session *services.SessionIO) error {
	ret := _m.Called(ctx, session)

	if len(ret) == 0 {
		panic("no return value specified for Handle")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *services.SessionIO) error); ok {
		r0 = rf(ctx, session)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockCommandHandler_Handle_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Handle'
type mockCommandHandler_Handle_Call struct {
	*mock.Call
}

// Handle is a helper method to define mock.On call
//   - ctx context.Context
//   - session *services.SessionIO
func (_e *mockCommandHandler_Expecter) Handle(ctx interface{}, session interface{}) *mockCommandHandler_Handle_Call {
	return &mockCommandHandler_Handle_Call{Call: _e.mock.On("Handle", ctx, session)}
}

func (_c *mockCommandHandler_Handle_Call) Run(run func(ctx context.Context, session *services.SessionIO)) *mockCommandHandler_Handle_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*services.SessionIO))
	})
	return _c
}

func (_c *mockCommandHandler_Handle_Call) Return(_a0 error) *mockCommandHandler_Handle_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockCommandHandler_Handle_Call) RunAndReturn(run func(context.Context, *services.SessionIO) error) *mockCommandHandler_Handle_Call {
	_c.Call.Return(run)
	return _c
}

// newMockCommandHandler creates a new instance of mockCommandHandler. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newMockCommandHandler(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockCommandHandler {
	mock := &mockCommandHandler{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
