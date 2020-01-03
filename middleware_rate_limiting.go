package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/gorilla/context"
	"net/http"
)

type RateLimitAndQuotaCheck struct {
	RaspberryMiddleware
}

func (k RateLimitAndQuotaCheck) New() func(http.Handler) http.Handler {
	aliceHandler := func(h http.Handler) http.Handler {
		thisHandler := func(w http.ResponseWriter, r *http.Request) {

			thisSessionState := context.Get(r, SessionData).(SessionState)
			authHeaderValue := context.Get(r, AuthHeaderValue).(string)
			forwardMessage, reason := sessionLimiter.ForwardMessage(&thisSessionState)

			// Ensure all this gets recorded
			authManager.UpdateSession(authHeaderValue, thisSessionState)

			if !forwardMessage {
				// TODO Use an Enum!
				if reason == 1 {
					log.WithFields(logrus.Fields{
						"path":   r.URL.Path,
						"origin": r.RemoteAddr,
						"key":    authHeaderValue,
					}).Info("Key rate limit exceeded.")

					handler := ErrorHandler{k.RaspberryMiddleware}
					handler.HandleError(w, r, "Key rate limit exceeded", 403)
					return
				} else if reason == 2 {
					log.WithFields(logrus.Fields{
						"path":   r.URL.Path,
						"origin": r.RemoteAddr,
						"key":    authHeaderValue,
					}).Info("Key quota limit exceeded.")
					handler := ErrorHandler{k.RaspberryMiddleware}
					handler.HandleError(w, r, "Key quota limit exceeded.", 403)
					return
				}
				// Other reason? Still not allowed
				return
			}

			// Request is valid, carry on
			h.ServeHTTP(w, r)
		}

		return http.HandlerFunc(thisHandler)
	}

	return aliceHandler
}
