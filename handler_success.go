package main

import (
	"net/http"
	"net/http/httputil"
	"runtime/pprof"
	"strings"
	"time"

	"github.com/gorilla/context"
)

// ContextKey is a key type to avoid collisions
type ContextKey int

// Enums for keys to be stored in a session context - this is how gorilla expects
// these to be implemented and is lifted pretty much from docs
const (
	SessionData     = 0
	AuthHeaderValue = 1
)

// RaspberryMiddleware wraps up the APISpec and Proxy objects to be included in a
// middleware handler, this can probably be handled better.
type RaspberryMiddleware struct {
	Spec  APISpec
	Proxy *httputil.ReverseProxy
}

// SuccessHandler represents the final ServeHTTP() request for a proxied API request
type SuccessHandler struct {
	RaspberryMiddleware
}

// ServeHTTP will store the request details in the analytics store if necessary and proxy the request to it's
// final destination, this is invoked by the ProxyHandler or right at the start of a request chain if the URL
// Spec states the path is Ignored
func (s SuccessHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if config.EnableAnalytics {
		t := time.Now()
		keyName := r.Header.Get(s.Spec.APIDefinition.Auth.AuthHeaderName)
		version := s.Spec.getVersionFromRequest(r)
		if version == "" {
			version = "Non Versioned"
		}

		if s.Spec.APIDefinition.Proxy.StripListenPath {
			r.URL.Path = strings.Replace(r.URL.Path, s.Spec.Proxy.ListenPath, "", 1)
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
			s.Spec.APIDefinition.Name,
			s.Spec.APIDefinition.APIID,
			s.Spec.APIDefinition.OrgID}
		analytics.RecordHit(thisRecord)
	}

	s.Proxy.ServeHTTP(w, r)

	if doMemoryProfile {
		pprof.WriteHeapProfile(profileFile)
	}

	context.Clear(r)
}
