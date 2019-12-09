package main

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/docopt/docopt.go"
	"html/template"
	"net/http"
	"net/http/httputil"
	"net/url"
)

/*
TODO: Set configuration file (Command line)
TODO: ConfigurationL: set redis DB details
TODO: Redis storage manager
TODO: API endpoints for management functions: RequestKey (creates a key for user instead of self supplied)
TODO: Secure API endpoints (perhaps with just a shared secret, should be internally used anyway)
TODO: Configuration: Set shared secret
TODO: Make SessionLimiter an interface so we can have different limiter types (e.g. queued requests?)
TODO: Add QuotaLimiter so time-based quotas can be added
TODO: Keys should expire
*/

var log = logrus.New()
var authManager = AuthorisationManager{}
var sessionLimiter = SessionLimiter{}
var config = Config{}
var templates = &template.Template{}
var systemError string = "{\"status\": \"system error, please contact administrator\"}"

func setupGlobals() {
	if config.Storage.Type == "memory" {
		authManager = AuthorisationManager{InMemoryStorageManager{map[string]string{}}}
	}

	templateFile := fmt.Sprintf("%s/error.json", config.TemplatePath)
	templates = template.Must(template.ParseFiles(templateFile))
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
	createSampleSession()
	loadConfig("respberry.conf", &config)
	remote, err := url.Parse(config.TargetUrl)
	if err != nil {
		log.Error("Couldn't parse target URL")
		log.Error(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)
	http.HandleFunc("/raspberry/key/", keyHandler)
	http.HandleFunc(config.ListenPath, handler(proxy))
	targetPort := fmt.Sprintf(":%d", config.ListenPort)
	err = http.ListenAndServe(targetPort, nil)
	if err != nil {
		log.Error(err)
	}
}
