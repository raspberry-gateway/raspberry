package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/gorilla/context"
	"net/http"
)

type KeyExists struct {
	RaspberryMiddleware
}

func (k KeyExists) New() func(http.Handler) http.Handler {
	aliceHandler := func(h http.Handler) http.Handler {
		thisHandler := func(w http.ResponseWriter, r *http.Request) {

			authHeaderValue := r.Header.Get(k.Spec.APIDefinition.Auth.AuthHeaderName)
			if authHeaderValue == "" {
				// No header value, fail
				log.WithFields(logrus.Fields{
					"path":   r.URL.Path,
					"origin": r.RemoteAddr,
				}).Info("Attempted access with malformed header, no auth header found.")

				handler := ErrorHandler{k.RaspberryMiddleware}
				handler.HandleError(w, r, "Authorisation failed missing", 400)
				return
			}

			// Check if API key valid
			key_exists, thisSessionState := authManager.IsKeyAuthorised(authHeaderValue)
			if !key_exists {
				log.WithFields(logrus.Fields{
					"path":   r.URL.Path,
					"origin": r.RemoteAddr,
					"key":    authHeaderValue,
				}).Info("Attempted access with non-existent key.")

				handler := ErrorHandler{k.RaspberryMiddleware}
				handler.HandleError(w, r, "Key not authorised", 403)
				return
			}

			// Set session state on context, we will need it later
			context.Set(r, SessionData, thisSessionState)
			context.Set(r, AuthHeaderValue, authHeaderValue)

			// Request is valid, carry on
			h.ServeHTTP(w, r)
		}

		return http.HandlerFunc(thisHandler)
	}

	return aliceHandler
}
