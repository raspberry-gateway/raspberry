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
TODO: Make SessionLimiter an interface so we can have different limiter types (e.g. queued requests?)
*/

var log = logrus.New()
var authManager = AuthorisationManager{}
var sessionLimiter = SessionLimiter{}
var config = Config{}
var templates = &template.Template{}
var systemError string = "{\"status\": \"system error, please contact administrator\"}"

func setupGlobals() {
	if config.Storage.Type == "memory" {
		authManager = AuthorisationManager{
			InMemoryStorageManager{
				map[string]string{}}}
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
	http.HandleFunc("/raspberry/keys/create", securityHandler(createKeyHandler))
	http.HandleFunc("/raspberry/keys/", securityHandler(keyHandler))
	http.HandleFunc(config.ListenPath, handler(proxy))
	targetPort := fmt.Sprintf(":%d", config.ListenPort)
	err = http.ListenAndServe(targetPort, nil)
	if err != nil {
		log.Error(err)
	}
}
