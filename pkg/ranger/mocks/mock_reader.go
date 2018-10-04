// Code generated by MockGen. DO NOT EDIT.
// Source: storj.io/storj/pkg/ranger (interfaces: Ranger)

// Package mock_ranger is a generated GoMock package.
package mock_ranger

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	io "io"
	reflect "reflect"
)

// MockRanger is a mock of Ranger interface
type MockRanger struct {
	ctrl     *gomock.Controller
	recorder *MockRangerMockRecorder
}

// MockRangerMockRecorder is the mock recorder for MockRanger
type MockRangerMockRecorder struct {
	mock *MockRanger
}

// NewMockRanger creates a new mock instance
func NewMockRanger(ctrl *gomock.Controller) *MockRanger {
	mock := &MockRanger{ctrl: ctrl}
	mock.recorder = &MockRangerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockRanger) EXPECT() *MockRangerMockRecorder {
	return m.recorder
}

// Range mocks base method
func (m *MockRanger) Range(arg0 context.Context, arg1, arg2 int64) (io.ReadCloser, error) {
	ret := m.ctrl.Call(m, "Range", arg0, arg1, arg2)
	ret0, _ := ret[0].(io.ReadCloser)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Range indicates an expected call of Range
func (mr *MockRangerMockRecorder) Range(arg0, arg1, arg2 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Range", reflect.TypeOf((*MockRanger)(nil).Range), arg0, arg1, arg2)
}

// Size mocks base method
func (m *MockRanger) Size() int64 {
	ret := m.ctrl.Call(m, "Size")
	ret0, _ := ret[0].(int64)
	return ret0
}

// Size indicates an expected call of Size
func (mr *MockRangerMockRecorder) Size() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Size", reflect.TypeOf((*MockRanger)(nil).Size))
}