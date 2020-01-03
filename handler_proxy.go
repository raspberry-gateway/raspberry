package main

import (
	"net/http"
	"net/http/httputil"
)

type ApiError struct {
	Message string
}

// proxies request onwards
func ProxyHandler(p *httputil.ReverseProxy, apiSpec ApiSpec) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		tm := RaspberryMiddleware{apiSpec, p}
		handler := SuccessHandler{tm}
		// Skip all other execution
		handler.ServeHttp(w, r)
		return
	}
}
