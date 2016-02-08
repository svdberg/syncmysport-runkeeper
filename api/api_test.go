package api

import (
	"net/http"
	"net/url"
	"testing"
)

type mockWriter struct{}

var responseWriter http.ResponseWriter = &mockWriter{}

func TestTokenDisassociate(t *testing.T) {
	var request http.Request = http.Request{}
	request.URL, _ = url.ParseRequestURI("http://www.syncmysport.com/token/12345")

	TokenDisassociate(responseWriter, &request)

}

//mocks
func (m mockWriter) Header() http.Header {
	return nil
}

func (m mockWriter) Write([]byte) (int, error) { return 0, nil }

func (m mockWriter) WriteHeader(int) {}
