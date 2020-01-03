package main

import (
	"fmt"
	"html/template"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/buger/goterm"
	"github.com/docopt/docopt.go"
	"github.com/rcrowley/goagain"
)

/*
TODO: Set configuration file (Command line)
TODO: Make SessionLimiter an interface so we can have different limiter types (e.g. queued requests?)
*/

var log = logrus.New()
var authManager = AuthorisationManager{}
var sessionLimiter = SessionLimiter{}
var config = Config{}
var templates = &template.Template{}
var systemError string = "{\"status\": \"system error, please contact administrator\"}"
var analytics = RedisAnalyticsHandler{}
var prof_file = &os.File{}
var doMemoryProfile bool

func displayConfig() {
	configTable := goterm.NewTable(0, 10, 5, ' ', 0)
	fmt.Fprintf(configTable, "Listening on port:\t%d\n", config.ListenPort)

	fmt.Println(configTable)
	fmt.Println("")
}

func setupGlobals() {
	if config.Storage.Type == "memory" {
		log.Warning("Using in-memory storage. Warning: this is not scalable.")
		authManager = AuthorisationManager{
			&InMemoryStorageManager{
				map[string]string{}}}
	} else if config.Storage.Type == "redis" {
		log.Info("Using redis storage manager.")
		authManager = AuthorisationManager{
			&RedisStorageManager{KeyPrefix: "apikey-"}}
		authManager.Store.Connect()
	}

	if config.EnableAnalytics && config.Storage.Type != "redis" {
		log.Panic("Analytics requires Redis Storage backend, please enable Redis in the raspberry.conf file.")
	}

	if config.EnableAnalytics {
		AnalyticsStore := RedisStorageManager{KeyPrefix: "analytics-"}
		log.Info("Setting up analytics DB connection")

		if config.AnalyticsConfig.Type == "csv" {
			log.Info("using CSV cache purge")
			analytics = RedisAnalyticsHandler{
				Store: &AnalyticsStore,
				Clean: &CSVPurger{&AnalyticsStore}}
		} else if config.AnalyticsConfig.Type == "mongo" {
			log.Info("Using MongoDB cache purge")
			analytics = RedisAnalyticsHandler{
				Store: &AnalyticsStore,
				Clean: &MongoPurger{&AnalyticsStore, nil}}
		}

		analytics.Store.Connect()
		go analytics.Clean.StartPurgeLoop(config.AnalyticsConfig.PurgeDelay)
	}

	templateFile := fmt.Sprintf("%s/error.json", config.TemplatePath)
	templates = template.Must(template.ParseFiles(templateFile))
}

func init() {
	usage := `Raspberry API Gateway.
	
	Usage:
		raspberry [options]

	Options:
		-h --help		Show this screen
		--conf=FILE		Load a named configuration file
		--port=PORT	 	Listen on PORT (overrides config file)
		--memprofile 	Generate a memory profile

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

	port, _ := arguments["--port"]
	if port != nil {
		portNum, err := strconv.Atoi(port.(string))
		if err != nil {
			log.Error("Port specified in flags must be a number!")
			log.Error(err)
		} else {
			config.ListenPort = portNum
		}
	}

	doMemoryProfile, _ = arguments["--memprofile"].(bool)

}

func intro() {
	fmt.Print("\n\n")
	fmt.Println(goterm.Bold(goterm.Color("Raspberry.io Gateway API v0.1", goterm.GREEN)))
	fmt.Println(goterm.Bold(goterm.Color("=============================", goterm.GREEN)))
	fmt.Print("Copyright Lance. @2019")
}

func loadApiEndpoints() {
	// set up main API handlers
	http.HandleFunc("/raspberry/keys/create", securityHandler(createKeyHandler))
	http.HandleFunc("/raspberry/keys/", securityHandler(keyHandler))
	http.HandleFunc("/raspberry/reload", securityHandler(resetHandler))
}

func loadApps() {
	// load the API defs
	log.Info("Loading API configurations.")
	thisApiLoader := ApiDefinitionLoader{}

	var ApiSpecs []ApiSpec

	if config.UseDBAppConfigs {
		log.Info("Using App Configuration from Mongo DB")
		ApiSpecs = thisApiLoader.LoadDefinitionsFromMongo()
	} else {
		ApiSpecs = thisApiLoader.LoadDefinitions("./apps/")
	}

	for _, spec := range ApiSpecs {
		// Create a new handler for each API spec
		remote, err := url.Parse(spec.ApiDefinition.Proxy.TargetUrl)
		if err != nil {
			log.Error("Could not parse target URL")
			log.Error(err)
		}
		log.Info(remote)
		proxy := httputil.NewSingleHostReverseProxy(remote)
		http.HandleFunc(spec.Proxy.ListenPath, handler(proxy, spec))
	}
}

func main() {
	intro()
	displayConfig()

	if doMemoryProfile {
		log.Info("Memory profiling active")
		prof_file, _ = os.Create("raspberry.mprof")
		defer prof_file.Close()
	}

	targetPort := fmt.Sprintf(":%d", config.ListenPort)
	loadApiEndpoints()

	// Handle reload when SIGUSR2 is received
	l, err := goagain.Listener()
	if nil != err {
		// Listen on a TCP or a UNIX domain socket (TCP here).
		l, err = net.Listen("tcp", targetPort)
		if nil != err {
			log.Fatalln("")
		}
		log.Println("Listening on ", l.Addr())

		// Accept connections in a new goroutine
		loadApps()
		go http.Serve(l, nil)
	} else {
		// Resume accepting connextions in a new goroutine.
		log.Panicln("Resuming listening on", l.Addr())
		loadApps()
		go http.Serve(l, nil)

		// Kill the parent, now that the child has started successfully.
		if err := goagain.Kill(); nil != err {
			log.Fatalln(err)
		}
	}

	// Block the main goroutine awaiting signals.
	if _, err := goagain.Wait(l); nil != err {
		log.Fatalln(err)
	}

	// Do whatever's necessary to ensure a graceful exit like waiting for
	// goroutines to terminate or a channel to become closed.
	//
	// In this case, we'll simply stop listening and wait one second.
	if err := l.Close(); nil != err {
		log.Fatalln(err)
	}
	time.Sleep(1e9)
}
