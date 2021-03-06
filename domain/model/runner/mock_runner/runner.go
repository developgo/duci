// Code generated by MockGen. DO NOT EDIT.
// Source: domain/model/runner/runner.go

// Package mock_runner is a generated GoMock package.
package mock_runner

import (
	context "context"
	docker "github.com/duck8823/duci/domain/model/docker"
	job "github.com/duck8823/duci/domain/model/job"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockDockerRunner is a mock of DockerRunner interface
type MockDockerRunner struct {
	ctrl     *gomock.Controller
	recorder *MockDockerRunnerMockRecorder
}

// MockDockerRunnerMockRecorder is the mock recorder for MockDockerRunner
type MockDockerRunnerMockRecorder struct {
	mock *MockDockerRunner
}

// NewMockDockerRunner creates a new mock instance
func NewMockDockerRunner(ctrl *gomock.Controller) *MockDockerRunner {
	mock := &MockDockerRunner{ctrl: ctrl}
	mock.recorder = &MockDockerRunnerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockDockerRunner) EXPECT() *MockDockerRunnerMockRecorder {
	return m.recorder
}

// Run mocks base method
func (m *MockDockerRunner) Run(ctx context.Context, dir job.WorkDir, tag docker.Tag, cmd docker.Command) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Run", ctx, dir, tag, cmd)
	ret0, _ := ret[0].(error)
	return ret0
}

// Run indicates an expected call of Run
func (mr *MockDockerRunnerMockRecorder) Run(ctx, dir, tag, cmd interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Run", reflect.TypeOf((*MockDockerRunner)(nil).Run), ctx, dir, tag, cmd)
}
