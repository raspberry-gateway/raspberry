package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/gorilla/context"
	"net/http"
)

type AccessRightsCheck struct {
	RaspberryMiddleware
}

func (a AccessRightsCheck) New() func(http.Handler) http.Handler {
	aliceHandler := func(h http.Handler) http.Handler {
		thisHandler := func(w http.ResponseWriter, r *http.Request) {

			accessingVersion := a.Spec.getVersionFromRequest(r)
			thisSessionState := context.Get(r, SessionData).(SessionState)
			authHeaderValue := context.Get(r, AuthHeaderValue).(string)

			// If there's nothing in our profile, we let them through to the next phase
			if len(thisSessionState.AccessRights) > 0 {
				// Otherwise, run auth checks
				versionList, apiExists := thisSessionState.AccessRights[a.Spec.ApiId]
				if !apiExists {
					log.WithFields(logrus.Fields{
						"path":   r.URL.Path,
						"origin": r.RemoteAddr,
						"key":    authHeaderValue,
					}).Info("Attempted access to unauthorised API.")
					handler := ErrorHandler{a.RaspberryMiddleware}
					handler.HandleError(w, r, "Access to this API has been disallowed", 403)
					return
				}

				// Find the version in their key access details
				found := false
				for _, vInfo := range versionList.Versions {
					if vInfo == accessingVersion {
						found = true
						break
					}
				}
				if !found {
					// Not found? Bounce
					log.WithFields(logrus.Fields{
						"path":   r.URL.Path,
						"origin": r.RemoteAddr,
						"key":    authHeaderValue,
					}).Info("Attempted access to unauthorised API version.")
					handler := ErrorHandler{a.RaspberryMiddleware}
					handler.HandleError(w, r, "Access to this API has been disallowed", 403)
					return
				}
			}

			// No fates failed, request is valid, carry on
			h.ServeHTTP(w, r)
		}

		return http.HandlerFunc(thisHandler)
	}

	return aliceHandler
}
