// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/gardener/gardener-extensions/controllers/provider-alicloud/pkg/alicloud/client (interfaces: VPC,Factory)

// Package client is a generated GoMock package.
package client

import (
	vpc "github.com/aliyun/alibaba-cloud-sdk-go/services/vpc"
	client "github.com/gardener/gardener-extensions/controllers/provider-alicloud/pkg/alicloud/client"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockVPC is a mock of VPC interface
type MockVPC struct {
	ctrl     *gomock.Controller
	recorder *MockVPCMockRecorder
}

// MockVPCMockRecorder is the mock recorder for MockVPC
type MockVPCMockRecorder struct {
	mock *MockVPC
}

// NewMockVPC creates a new mock instance
func NewMockVPC(ctrl *gomock.Controller) *MockVPC {
	mock := &MockVPC{ctrl: ctrl}
	mock.recorder = &MockVPCMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockVPC) EXPECT() *MockVPCMockRecorder {
	return m.recorder
}

// DescribeNatGateways mocks base method
func (m *MockVPC) DescribeNatGateways(arg0 *vpc.DescribeNatGatewaysRequest) (*vpc.DescribeNatGatewaysResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DescribeNatGateways", arg0)
	ret0, _ := ret[0].(*vpc.DescribeNatGatewaysResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DescribeNatGateways indicates an expected call of DescribeNatGateways
func (mr *MockVPCMockRecorder) DescribeNatGateways(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DescribeNatGateways", reflect.TypeOf((*MockVPC)(nil).DescribeNatGateways), arg0)
}

// DescribeVpcs mocks base method
func (m *MockVPC) DescribeVpcs(arg0 *vpc.DescribeVpcsRequest) (*vpc.DescribeVpcsResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DescribeVpcs", arg0)
	ret0, _ := ret[0].(*vpc.DescribeVpcsResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DescribeVpcs indicates an expected call of DescribeVpcs
func (mr *MockVPCMockRecorder) DescribeVpcs(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DescribeVpcs", reflect.TypeOf((*MockVPC)(nil).DescribeVpcs), arg0)
}

// MockFactory is a mock of Factory interface
type MockFactory struct {
	ctrl     *gomock.Controller
	recorder *MockFactoryMockRecorder
}

// MockFactoryMockRecorder is the mock recorder for MockFactory
type MockFactoryMockRecorder struct {
	mock *MockFactory
}

// NewMockFactory creates a new mock instance
func NewMockFactory(ctrl *gomock.Controller) *MockFactory {
	mock := &MockFactory{ctrl: ctrl}
	mock.recorder = &MockFactoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockFactory) EXPECT() *MockFactoryMockRecorder {
	return m.recorder
}

// NewVPC mocks base method
func (m *MockFactory) NewVPC(arg0, arg1, arg2 string) (client.VPC, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewVPC", arg0, arg1, arg2)
	ret0, _ := ret[0].(client.VPC)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// NewVPC indicates an expected call of NewVPC
func (mr *MockFactoryMockRecorder) NewVPC(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewVPC", reflect.TypeOf((*MockFactory)(nil).NewVPC), arg0, arg1, arg2)
}
