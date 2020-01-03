package main

import (
	"net/http"
	"net/http/httputil"
	"runtime/pprof"
	"strings"
	"time"

	"github.com/gorilla/context"
)

type ContextKey int

const (
	SessionData     = 0
	AuthHeaderValue = 1
)

type RaspberryMiddleware struct {
	Spec  ApiSpec
	Proxy *httputil.ReverseProxy
}

type SuccessHandler struct {
	RaspberryMiddleware
}

func (s SuccessHandler) ServeHttp(w http.ResponseWriter, r *http.Request) {
	if config.EnableAnalytics {
		t := time.Now()
		keyName := r.Header.Get(s.Spec.ApiDefinition.Auth.AuthHeaderName)
		version := s.Spec.getVersionFromRequest(r)
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
			s.Spec.ApiDefinition.Name,
			s.Spec.ApiDefinition.ApiId,
			s.Spec.OrgId}
		analytics.RecordHit(thisRecord)
	}

	if s.Spec.ApiDefinition.Proxy.StripListenPath {
		r.URL.Path = strings.Replace(r.URL.Path, s.Spec.Proxy.ListenPath, "", 1)
	}

	s.Proxy.ServeHTTP(w, r)

	if doMemoryProfile {
		pprof.WriteHeapProfile(prof_file)
	}

	context.Clear(r)
}
