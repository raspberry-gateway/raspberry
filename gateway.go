package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"runtime/pprof"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
)

type ApiError struct {
	Message string
}

func handler(p *httputil.ReverseProxy, apiSpec ApiSpec) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		// Check versioning, blacklist, whitelist and ignored status
		requestValid, stat := apiSpec.IsRequestValid(r)
		if requestValid == false {
			handleError(w, r, string(stat), 409, apiSpec)
		}

		if stat == StatusOkAndIgnore {
			successHandler(w, r, p, apiSpec)
			return
		}

		// All is ok with the request itself, now auth and validate the rest
		// Check for API key existence
		authHeaderValue := r.Header.Get(apiSpec.ApiDefinition.Auth.AuthHeaderName)
		if authHeaderValue != "" {
			// Check if API key valid
			keyAuthorised, thisSessionState := authManager.IsKeyAuthorised(authHeaderValue)

			// Check if this version is allowable!
			accessingVersion := apiSpec.getVersionFromRequest(r)
			apiId := apiSpec.ApiId

			// If there's nothing in our profile, we let them through to the next phase
			if len(thisSessionState.AccessRights) > 0 {
				// Run auth checks
				versionList, apiExists := thisSessionState.AccessRights[apiId]
				if !apiExists {
					log.WithFields(logrus.Fields{
						"path":   r.URL.Path,
						"origin": r.RemoteAddr,
						"key":    authHeaderValue,
					}).Info("Attempted access to unauthorised API.")
					handleError(w, r, "Access to this API has been disallowed", 403, apiSpec)
					return
				} else {
					found := false
					for _, vInfo := range versionList.Versions {
						if vInfo == accessingVersion {
							found = true
							break
						}
					}
					if !found {
						log.WithFields(logrus.Fields{
							"path":   r.URL.Path,
							"origin": r.RemoteAddr,
							"key":    authHeaderValue,
						}).Info("Attempted access to unauthorised API version.")
						handleError(w, r, "Access to this API version has been disallowed", 403, apiSpec)
						return
					}
				}
			}

			keyExpired := authManager.IsKeyExpired(&thisSessionState)
			if keyAuthorised {
				if !keyExpired {
					// If valid, check if within rate limit
					forwardMessage, reason := sessionLimiter.ForwardMessage(&thisSessionState)
					if forwardMessage {
						successHandler(w, r, p, apiSpec)
					} else {
						// TODO Use an Enum!
						if reason == 1 {
							log.WithFields(logrus.Fields{
								"path":   r.URL.Path,
								"origin": r.RemoteAddr,
								"key":    authHeaderValue,
							}).Info("rate limit exceeded.")
							handleError(w, r, "Rate limit exceeded", 409, apiSpec)
						} else if reason == 2 {
							log.WithFields(logrus.Fields{
								"path":   r.URL.Path,
								"origin": r.RemoteAddr,
								"key":    authHeaderValue,
							}).Info("Key quota limit exceeded.")
							handleError(w, r, "quota exceeded", 409, apiSpec)
						}
					}
					authManager.UpdateSession(authHeaderValue, thisSessionState)
				} else {
					log.WithFields(logrus.Fields{
						"path":   r.URL.Path,
						"origin": r.RemoteAddr,
						"key":    authHeaderValue,
					}).Info("Attempted access from expired key.")
					handleError(w, r, "Key has expired, please renew", 403, apiSpec)
				}
			} else {
				log.WithFields(logrus.Fields{
					"path":   r.URL.Path,
					"origin": r.RemoteAddr,
					"key":    authHeaderValue,
				}).Info("Attempted access with non-existend key.")
				handleError(w, r, "Key not authorised", 403, apiSpec)
			}
		} else {
			log.WithFields(logrus.Fields{
				"path":   r.URL.Path,
				"origin": r.RemoteAddr,
				"key":    authHeaderValue,
			}).Info("Attempted access with malformed header, no auth header found.")
			handleError(w, r, "Authorisation field missing", 400, apiSpec)
		}
	}
}

func successHandler(w http.ResponseWriter, r *http.Request, p *httputil.ReverseProxy, apiSpec ApiSpec) {
	if config.EnableAnalytics {
		t := time.Now()
		keyName := r.Header.Get(apiSpec.ApiDefinition.Auth.AuthHeaderName)
		version := apiSpec.getVersionFromRequest(r)
		if version == "" {
			version = "Non Versioned"
		}
		thisRecord := AnalyticsRecord{
			r.Method,
			r.URL.Path,
			r.ContentLength,
			r.Header.Get("User-agent"),
			t.Day(),
			t.Month(),
			t.Year(),
			t.Hour(),
			200,
			keyName,
			t,
			version,
			apiSpec.ApiDefinition.Name,
			apiSpec.ApiDefinition.ApiId,
			apiSpec.OrgId}
		analytics.RecordHit(thisRecord)
	}

	if apiSpec.ApiDefinition.Proxy.StripListenPath {
		r.URL.Path = strings.Replace(r.URL.Path, apiSpec.Proxy.ListenPath, "", 1)
	}

	p.ServeHTTP(w, r)

	if doMemoryProfile {
		pprof.WriteHeapProfile(prof_file)
	}
}

func handleError(w http.ResponseWriter, r *http.Request, err string, errCode int, apiSpec ApiSpec) {
	if config.EnableAnalytics {
		t := time.Now()
		keyName := r.Header.Get(apiSpec.ApiDefinition.Auth.AuthHeaderName)
		version := apiSpec.getVersionFromRequest(r)
		if version == "" {
			version = "Non Versioned"
		}
		thisRecord := AnalyticsRecord{
			r.Method,
			r.URL.Path,
			r.ContentLength,
			r.Header.Get("User-agent"),
			t.Day(),
			t.Month(),
			t.Year(),
			t.Hour(),
			errCode,
			keyName,
			t,
			version,
			apiSpec.ApiDefinition.Name,
			apiSpec.ApiDefinition.ApiId,
			apiSpec.OrgId}
		analytics.RecordHit(thisRecord)
	}

	w.WriteHeader(errCode)
	w.Header().Add("Content-Type", "application/json")
	w.Header().Add("x-Generator", "raspberry.io")
	thisError := ApiError{fmt.Sprintf("%s", err)}
	templates.ExecuteTemplate(w, "error.json", &thisError)
	if doMemoryProfile {
		pprof.WriteHeapProfile(prof_file)
	}
}
