// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/alicansa/go-linkcrawler/crawler (interfaces: CrawlPolicyExecuter,WebCrawler)

// Package mocks is a generated GoMock package.
package mocks

import (
	io "io"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockCrawlPolicyExecuter is a mock of CrawlPolicyExecuter interface.
type MockCrawlPolicyExecuter struct {
	ctrl     *gomock.Controller
	recorder *MockCrawlPolicyExecuterMockRecorder
}

// MockCrawlPolicyExecuterMockRecorder is the mock recorder for MockCrawlPolicyExecuter.
type MockCrawlPolicyExecuterMockRecorder struct {
	mock *MockCrawlPolicyExecuter
}

// NewMockCrawlPolicyExecuter creates a new mock instance.
func NewMockCrawlPolicyExecuter(ctrl *gomock.Controller) *MockCrawlPolicyExecuter {
	mock := &MockCrawlPolicyExecuter{ctrl: ctrl}
	mock.recorder = &MockCrawlPolicyExecuterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCrawlPolicyExecuter) EXPECT() *MockCrawlPolicyExecuterMockRecorder {
	return m.recorder
}

// Execute mocks base method.
func (m *MockCrawlPolicyExecuter) Execute(arg0 io.ReadCloser) ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Execute", arg0)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Execute indicates an expected call of Execute.
func (mr *MockCrawlPolicyExecuterMockRecorder) Execute(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Execute", reflect.TypeOf((*MockCrawlPolicyExecuter)(nil).Execute), arg0)
}

// MockWebCrawler is a mock of WebCrawler interface.
type MockWebCrawler struct {
	ctrl     *gomock.Controller
	recorder *MockWebCrawlerMockRecorder
}

// MockWebCrawlerMockRecorder is the mock recorder for MockWebCrawler.
type MockWebCrawlerMockRecorder struct {
	mock *MockWebCrawler
}

// NewMockWebCrawler creates a new mock instance.
func NewMockWebCrawler(ctrl *gomock.Controller) *MockWebCrawler {
	mock := &MockWebCrawler{ctrl: ctrl}
	mock.recorder = &MockWebCrawlerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockWebCrawler) EXPECT() *MockWebCrawlerMockRecorder {
	return m.recorder
}

// Crawl mocks base method.
func (m *MockWebCrawler) Crawl(arg0 string, arg1 func([]string) error) (map[string]struct{}, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Crawl", arg0, arg1)
	ret0, _ := ret[0].(map[string]struct{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Crawl indicates an expected call of Crawl.
func (mr *MockWebCrawlerMockRecorder) Crawl(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Crawl", reflect.TypeOf((*MockWebCrawler)(nil).Crawl), arg0, arg1)
}