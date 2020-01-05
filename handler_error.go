package main

import (
	"fmt"
	"github.com/gorilla/context"
	"net/http"
	"runtime/pprof"
	"strings"
	"time"
)

// APIError is generic error object returned if there is something wrong with the request
type APIError struct {
	Message string
}

// ErrorHandler is invoked whenever there is an issue with a proxied request, most middleware will invoke
// the ErrorHandler if something is wrong with the request and halt the request processing through the chain
type ErrorHandler struct {
	RaspberryMiddleware
}

// HandleError is the actual error handler and will store the error details in analytics if analytics processing is enabled.
func (e ErrorHandler) HandleError(w http.ResponseWriter, r *http.Request, err string, errCode int) {
	if config.EnableAnalytics {
		t := time.Now()
		keyName := r.Header.Get(e.Spec.APIDefinition.Auth.AuthHeaderName)
		version := e.Spec.getVersionFromRequest(r)
		if version == "" {
			version = "Non Versioned"
		}

		if e.RaspberryMiddleware.Spec.APIDefinition.Proxy.StripListenPath {
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
			errCode,
			keyName,
			t,
			version,
			e.Spec.APIDefinition.Name,
			e.Spec.APIDefinition.APIID,
			e.Spec.APIDefinition.OrgID}
		analytics.RecordHit(thisRecord)
	}

	w.WriteHeader(errCode)
	w.Header().Add("Content-Type", "application/json")
	w.Header().Add("x-Generator", "raspberry.io")
	thisError := APIError{fmt.Sprintf("%s", err)}
	templates.ExecuteTemplate(w, "error.json", &thisError)
	if doMemoryProfile {
		pprof.WriteHeapProfile(prof_file)
	}

	// Clean up
	context.Clear(r)
}
