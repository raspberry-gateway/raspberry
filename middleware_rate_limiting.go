package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/gorilla/context"
	"net/http"
)

// RateLimitAndQuotaCheck will check the incomming request and key whether it is within it's quota and
// within it's rate limit, it makes use of the SessionLimiter object to do this
type RateLimitAndQuotaCheck struct {
	RaspberryMiddleware
}

// New creates a new HttpHandler for the alice middleware package
func (k RateLimitAndQuotaCheck) New() func(http.Handler) http.Handler {
	aliceHandler := func(h http.Handler) http.Handler {
		thisHandler := func(w http.ResponseWriter, r *http.Request) {

			sessionLimiter := SessionLimiter{}
			thisSessionState := context.Get(r, SessionData).(SessionState)
			authHeaderValue := context.Get(r, AuthHeaderValue).(string)
			forwardMessage, reason := sessionLimiter.ForwardMessage(&thisSessionState)

			// Ensure quota and rate data for this session are recorded
			authManager.UpdateSession(authHeaderValue, thisSessionState)

			log.Info(thisSessionState)

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
