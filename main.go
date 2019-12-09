package main

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/docopt/docopt.go"
	"net/http"
	"net/http/httputil"
	"net/url"
)

/*
TODO: Set configuration file (Command line)
TODO: ConfigurationL: set redis DB details
TODO: Redis storage manager
TODO: API endpoints for management functions: AddKey, RevokeKey, UpdateKey, GetKeyDetails, RequestKey (creates a key for user instead of self supplied)
TODO: Secure API endpoints (perhaps with just a shared secret, should be internally used anyway)
TODO: Configuration: Set shared secret
TODO: Configuration: Error template file path
TODO: Make SessionLimiter an interface so we can have different limiter types (e.g. queued requests?)
TODO: Add QuotaLimiter so time-based quotas can be added
*/

var log = logrus.New()
var authManager = AuthorisationManager{}
var sessionLimiter = SessionLimiter{}
var config = Config{}

func setupGlobals() {
	if config.Storage.Type == "memory" {
		authManager = AuthorisationManager{InMemoryStorageManager{map[string]string{}}}
	}
}

func init() {
	usage := `Raspberry API Gateway.
	
	Usage:
		raspberry [options]

	Options:
		-h --help	Show this screen
		--conf=FILE	Load a named configuration file
		--test		Create a test key

	`

	arguments, err := docopt.Parse(usage, nil, true, "Raspberry v1.0", false)
	if err != nil {
		log.Println("Error while parsing auguments.")
		log.Fatal(err)
	}

	filename := "raspberry.conf"

	value, _ := arguments["--conf"]
	if value != nil {
		log.Info(fmt.Sprintf("Using %s for configuration", value.(string)))
		filename = arguments["--conf"].(string)
	} else {
		log.Info("No configuration file defined, will try to use default (./raspberry.conf)")
	}

	loadConfig(filename, &config)
	setupGlobals()

	testValue, _ := arguments["--test"].(bool)
	if testValue {
		log.Info("Adding test key: '1234' to storage map")
		authManager.Store.SetKey("1234", "{\"LastCheck\":1399469149,\"Allowance\":5.0,\"Rate\":1.0,\"Per\":1.0}")
	}
}

func main() {
	loadConfig("respberry.conf", &config)
	remote, err := url.Parse(config.TargetUrl)
	if err != nil {
		log.Error(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)
	http.HandleFunc(config.ListenPath, handler(proxy))
	targetPort := fmt.Sprintf(":%d", config.ListenPort)
	err = http.ListenAndServe(targetPort, nil)
	if err != nil {
		log.Error(err)
	}
}

func handler(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check for API key existence
		authorisation := r.Header.Get("authorisation")
		if authorisation != "" {
			// Check if API key valid
			keyAuthorised, thisSessionState := authManager.IsKeyAuthorised(authorisation)
			if keyAuthorised {
				// If valid, check if within rate limit
				forwardMessage := sessionLimiter.ForwardMessage(&thisSessionState)
				if forwardMessage {
					successHandler(w, r, p)
				} else {
					handleError(w, r, "Rate limit exceeded", 429)
				}
			} else {
				handleError(w, r, "Key not authorised", 403)
			}
		} else {
			handleError(w, r, "Authorisation field missing", 400)
		}
	}
}

func successHandler(w http.ResponseWriter, r *http.Request, p *httputil.ReverseProxy) {
	p.ServeHTTP(w, r)
}

func handleError(w http.ResponseWriter, r *http.Request, err string, errCode int) {
	w.WriteHeader(errCode)
	// TODO: This should be a template
	fmt.Fprintf(w, "NOT AUTHORISED: %s", err)
}
