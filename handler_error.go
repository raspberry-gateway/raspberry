package main

import (
	"fmt"
	"github.com/gorilla/context"
	"net/http"
	"runtime/pprof"
	"strings"
	"time"
)

type ErrorHandler struct {
	RaspberryMiddleware
}

func (e ErrorHandler) HandleError(w http.ResponseWriter, r *http.Request, err string, err_code int) {
	if config.EnableAnalytics {
		t := time.Now()
		keyName := r.Header.Get(e.Spec.ApiDefinition.Auth.AuthHeaderName)
		version := e.Spec.getVersionFromRequest(r)
		if version == "" {
			version = "Non Versioned"
		}

		if e.RaspberryMiddleware.Spec.ApiDefinition.Proxy.StripListenPath {
			r.URL.Path = strings.Replace(r.URL.Path, e.Spec.Proxy.ListenPath, "", 1)
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
			err_code,
			keyName,
			t,
			version,
			e.Spec.ApiDefinition.Name,
			e.Spec.ApiDefinition.ApiId,
			e.Spec.OrgId}
		analytics.RecordHit(thisRecord)
	}

	w.WriteHeader(err_code)
	w.Header().Add("Content-Type", "application/json")
	w.Header().Add("x-Generator", "raspberry.io")
	thisError := ApiError{fmt.Sprintf("%s", err)}
	templates.ExecuteTemplate(w, "error.json", &thisError)
	if doMemoryProfile {
		pprof.WriteHeapProfile(prof_file)
	}

	// Clean up
	context.Clear(r)
}
