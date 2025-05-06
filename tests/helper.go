package tests

import (
	"net/http"
	"net/http/httptest"
)

func mockHTTPServer(returnCode int) *httptest.Server {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(returnCode)
	}))
	return server
}
