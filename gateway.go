package main

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"net/http"
	"net/http/httputil"
	"runtime/pprof"
	"strings"
	"time"
)

type ApiError struct {
	Message string
}

func handler(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		for _, sub := range config.ExcludePaths {
			if strings.Contains(r.URL.Path, sub) {
				successHandler(w, r, p)
				return
			}
		}

		// Check for API key existence
		authHeaderValue := r.Header.Get("authorisation")
		if authHeaderValue != "" {
			// Check if API key valid
			keyAuthorised, thisSessionState := authManager.IsKeyAuthorised(authHeaderValue)
			keyExpired := authManager.IsKeyExpired(&thisSessionState)
			if keyAuthorised {
				if !keyExpired {
					// If valid, check if within rate limit
					forwardMessage, reason := sessionLimiter.ForwardMessage(&thisSessionState)
					if forwardMessage {
						successHandler(w, r, p)
					} else {
						if reason == 1 {
							log.WithFields(logrus.Fields{
								"path":   r.URL.Path,
								"origin": r.RemoteAddr,
								"key":    authHeaderValue,
							}).Info("rate limit exceeded.")
							handleError(w, r, "Rate limit exceeded", 409)
						} else if reason == 2 {
							log.WithFields(logrus.Fields{
								"path":   r.URL.Path,
								"origin": r.RemoteAddr,
								"key":    authHeaderValue,
							}).Info("Key quota limit exceeded.")
							handleError(w, r, "quota exceeded", 409)
						}
					}
					authManager.UpdateSession(authHeaderValue, thisSessionState)
				} else {
					log.WithFields(logrus.Fields{
						"path":   r.URL.Path,
						"origin": r.RemoteAddr,
						"key":    authHeaderValue,
					}).Info("Attempted access from expired key.")
					handleError(w, r, "Key has expired, please renew", 403)
				}
			} else {
				log.WithFields(logrus.Fields{
					"path":   r.URL.Path,
					"origin": r.RemoteAddr,
					"key":    authHeaderValue,
				}).Info("Attempted access with non-existend key.")
				handleError(w, r, "Key not authorised", 403)
			}
		} else {
			log.WithFields(logrus.Fields{
				"path":   r.URL.Path,
				"origin": r.RemoteAddr,
				"key":    authHeaderValue,
			}).Info("Attempted access with malformed header, no auth header found.")
			handleError(w, r, "Authorisation field missing", 400)
		}
	}
}

func successHandler(w http.ResponseWriter, r *http.Request, p *httputil.ReverseProxy) {
	if config.EnableAnalytics {
		t := time.Now()
		keyName := r.Header.Get(config.AuthHeaderName)
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
			t}
		analytics.RecordHit(thisRecord)
	}

	p.ServeHTTP(w, r)
	if doMemoryProfile {
		pprof.WriteHeapProfile(prof_file)
	}
}

func handleError(w http.ResponseWriter, r *http.Request, err string, errCode int) {
	if config.EnableAnalytics {
		t := time.Now()
		keyName := r.Header.Get(config.AuthHeaderName)
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
			t}
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
