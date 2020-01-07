package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/gorilla/context"
	"net/http"
	"strings"
)

// Oauth2KeyExists will check if the key being used to access the API is in the request data,
// and then if the key is in the storage engine
type Oauth2KeyExists struct {
	RaspberryMiddleware
}

// New creates a new HttpHandler for the alice middleware package
func (k Oauth2KeyExists) New() func(http.Handler) http.Handler {
	aliceHandler := func(h http.Handler) http.Handler {
		thisHandler := func(w http.ResponseWriter, r *http.Request) {

			if !k.Spec.UseOauth2 {
				// If we're not using OAuth2, skip
				h.ServeHTTP(w, r)
				return
			}

			// We're using OAuth2, start checking for access keys
			authHeaderValue := r.Header.Get("Authorization")
			parts := strings.Split(authHeaderValue, " ")
			if len(parts) < 2 {
				log.WithFields(logrus.Fields{
					"path":   r.URL.Path,
					"origin": r.RemoteAddr,
				}).Info("Attempted access with malformed header, no auth header found.")

				handler := ErrorHandler{k.RaspberryMiddleware}
				handler.HandleError(w, r, "Authorisation field missing", 400)
				return
			}

			if strings.ToLower(parts[0]) != "bearer" {
				log.WithFields(logrus.Fields{
					"path":   r.URL.Path,
					"origin": r.RemoteAddr,
				}).Info("Bearer token malformed")

				handler := ErrorHandler{k.RaspberryMiddleware}
				handler.HandleError(w, r, "Bearer token malformed", 400)
				return
			}

			accessToken := parts[1]
			keyExists, thisSessionState := authManager.IsKeyAuthorised(accessToken)

			if !keyExists {
				log.WithFields(logrus.Fields{
					"path":   r.URL.Path,
					"origin": r.RemoteAddr,
					"key":    accessToken,
				}).Info("Attempted access with non-existent key.")

				handler := ErrorHandler{k.RaspberryMiddleware}
				handler.HandleError(w, r, "Key not authorised", 403)
				return
			}

			// Set session state on context, we will need it later
			context.Set(r, SessionData, thisSessionState)
			context.Set(r, AuthHeaderValue, accessToken)

			// Request is valid, carry on
			h.ServeHTTP(w, r)

		}

		return http.HandlerFunc(thisHandler)
	}

	return aliceHandler
}
