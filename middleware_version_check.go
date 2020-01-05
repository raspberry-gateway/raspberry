package main

import (
	"net/http"
)

// VersionCheck will check whether the version of the requested API the request is accessing has any restrictions on URL endpoints
type VersionCheck struct {
	RaspberryMiddleware
}

// New creates a new HttpHandler for the alice middleware package
func (s VersionCheck) New() func(http.Handler) http.Handler {
	aliceHandler := func(h http.Handler) http.Handler {
		thisHandler := func(w http.ResponseWriter, r *http.Request) {

			// Check versioning, blacklist, whitelist and ignored status
			requestValid, stat := s.RaspberryMiddleware.Spec.IsRequestValid(r)
			if requestValid == false {
				handler := ErrorHandler{s.RaspberryMiddleware}
				// Stop execution
				handler.HandleError(w, r, string(stat), 409)
				return
			}

			if stat == StatusOkAndIgnore {
				handler := SuccessHandler{s.RaspberryMiddleware}
				// Skip all other execution
				handler.ServeHTTP(w, r)
				return
			}

			// Request is valid, carry on
			h.ServeHTTP(w, r)
		}

		return http.HandlerFunc(thisHandler)
	}

	return aliceHandler
}
