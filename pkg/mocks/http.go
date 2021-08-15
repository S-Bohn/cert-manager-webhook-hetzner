package mocks

import "net/http"

type HTTPMockClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (s HTTPMockClient) Do(req *http.Request) (*http.Response, error) {
	return s.DoFunc(req)
}
