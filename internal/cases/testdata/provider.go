// Code generated by MockGen. DO NOT EDIT.
// Source: provider.go

// Package testdata is a generated GoMock package.
package testdata

import (
	entities "Cryptoproject/internal/entities"
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockCryptoProvider is a mock of CryptoProvider interface.
type MockCryptoProvider struct {
	ctrl     *gomock.Controller
	recorder *MockCryptoProviderMockRecorder
}

// MockCryptoProviderMockRecorder is the mock recorder for MockCryptoProvider.
type MockCryptoProviderMockRecorder struct {
	mock *MockCryptoProvider
}

// NewMockCryptoProvider creates a new mock instance.
func NewMockCryptoProvider(ctrl *gomock.Controller) *MockCryptoProvider {
	mock := &MockCryptoProvider{ctrl: ctrl}
	mock.recorder = &MockCryptoProviderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCryptoProvider) EXPECT() *MockCryptoProviderMockRecorder {
	return m.recorder
}

// GetActualRates mocks base method.
func (m *MockCryptoProvider) GetActualRates(ctx context.Context, titles []string) ([]entities.Coin, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetActualRates", ctx, titles)
	ret0, _ := ret[0].([]entities.Coin)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetActualRates indicates an expected call of GetActualRates.
func (mr *MockCryptoProviderMockRecorder) GetActualRates(ctx, titles interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetActualRates", reflect.TypeOf((*MockCryptoProvider)(nil).GetActualRates), ctx, titles)
}
