package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
)

type ApiError struct {
	Message string
}

func handler(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		// Check for API key existence
		authorisation := r.Header.Get("authorisation")
		if authorisation != "" {
			// Check if API key valid
			keyAuthorised, thisSessionState := authManager.IsKeyAuthorised(authorisation)
			keyExpired := authManager.IsKeyExpired(&thisSessionState)
			if keyAuthorised {
				if !keyExpired {
					// If valid, check if within rate limit
					forwardMessage, reason := sessionLimiter.ForwardMessage(&thisSessionState)
					if forwardMessage {
						successHandler(w, r, p)
					} else {
						if reason == 1 {
							handleError(w, r, "Rate limit exceeded", 429)
						}
						handleError(w, r, "quota exceeded", 429)
					}
				} else {
					handleError(w, r, "Key has expired, please renew", 403)
				}
			} else {
				handleError(w, r, "Key not authorised", 403)
			}
		} else {
			handleError(w, r, "Authorisation field missing", 400)
		}
	}
}

func successHandler(w http.ResponseWriter, r *http.Request, p *httputil.ReverseProxy) {
	p.ServeHTTP(w, r)
}

func handleError(w http.ResponseWriter, r *http.Request, err string, errCode int) {
	w.WriteHeader(errCode)
	thisError := ApiError{fmt.Sprintf("%s", err)}
	templates.ExecuteTemplate(w, "error.json", &thisError)
}
