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

	"github.com/RangelReale/osin"
	"github.com/Sirupsen/logrus"
	"github.com/buger/goterm"
	"github.com/docopt/docopt.go"
	"github.com/justinas/alice"
	"github.com/rcrowley/goagain"
)

/*
TODO: Set configuration file (Command line)
TODO: Make SessionLimiter an interface so we can have different limiter types (e.g. queued requests?)
*/

var log = logrus.New()
var authManager = AuthorisationManager{}
var config = Config{}
var templates = &template.Template{}
var analytics = RedisAnalyticsHandler{}
var profileFile = &os.File{}
var doMemoryProfile bool
var genericOsinStorage *RedisOsinStorageInterface

// Generic system error
const (
	E_SYSTEM_ERROR          string = "{\"status\": \"system error, please contact administrator\"}"
	OAUTH_AUTH_CODE_TIMEOUT int    = 60 * 60
	OAUTH_PREFIX            string = "oauth-data."
)

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

	genericOsinStorage = MakeNewOsinServer()

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
	fmt.Print("\nhttp://www.respberry.io\n\n")
}

func loadAPIEndpoints(Muxer *http.ServeMux) {
	// set up main API handlers
	Muxer.HandleFunc("/raspberry/keys/create", CheckIsAPIOwner(createKeyHandler))
	Muxer.HandleFunc("/raspberry/keys/", CheckIsAPIOwner(keyHandler))
	Muxer.HandleFunc("/raspberry/reload", CheckIsAPIOwner(resetHandler))
	Muxer.HandleFunc("/raspberry/oauth/clients/create", CheckIsAPIOwner(createOauthClient))
}

func getAPISpecs() []APISpec {
	var APISpecs []APISpec
	thisAPILoader := APIDefinitionLoader{}

	if config.UseDBAppConfigs {
		log.Info("Using App Configuration from Mongo DB")
		APISpecs = thisAPILoader.LoadDefinitionsFromMongo()
	} else {
		APISpecs = thisAPILoader.LoadDefinitions("./apps/")
	}

	return APISpecs
}

func addOAuthHandlers(spec APISpec, Muxer *http.ServeMux) {
	apiAuthorizePath := spec.Proxy.ListenPath + "/raspberry/oauth/authorize-client/"
	clientAuthHandlePath := spec.Proxy.ListenPath + "/raspberry/oauth/clients"
	clientAuthPath := spec.Proxy.ListenPath + "oauth/authorize/"
	clientAccessPath := spec.Proxy.ListenPath + "oauth/token/"

	serverConfig := osin.NewServerConfig()
	serverConfig.ErrorStatusCode = 403
	serverConfig.AllowedAccessTypes = spec.Oauth2Meta.AllowedAccessTypes       // osin.AllowedAccessType{osin.AUTHORIZATION_CODE, osin.REFRESH_TOKEN}
	serverConfig.AllowedAuthorizeTypes = spec.Oauth2Meta.AllowedAuthorizeTypes // osin.AllowedAuthorizeType{osin.CODE, osin.TOKEN}

	OAuthPrefix := OAUTH_PREFIX + spec.APIID + "."
	storageManager := RedisStorageManager{KeyPrefix: OAuthPrefix}
	storageManager.Connect()
	osinStorage := RedisOsinStorageInterface{&storageManager}

	// TODO: Remove this
	log.Warning("Adding test client")
	testClient := &osin.Client{
		Id:          "1234",
		Secret:      "aabbccdd",
		RedirectUri: "http://client.oauth.com",
	}
	osinStorage.SetClient(testClient.Id, testClient, false)
	log.Warning("Test client added")

	osinServer := osin.NewServer(serverConfig, osinStorage)
	osinServer.AccessTokenGen = &AccessTokenGenRaspberry{}
	oauthManager := OAuthManager{spec, osinServer}
	oauthHandlers := OAuthHandlers{oauthManager}

	log.Warning("Configuration", spec.NotificationsDetails)

	Muxer.HandleFunc(apiAuthorizePath, CheckIsAPIOwner(oauthHandlers.HandleGenerateAuthCodeData))
	Muxer.HandleFunc(clientAuthPath, oauthHandlers.HandleAuthorizePassthrough)
	Muxer.HandleFunc(clientAccessPath, oauthHandlers.HandleAccessRequest)
	Muxer.HandleFunc(clientAuthHandlePath, CheckIsAPIOwner(oAuthClientHandler))
}

func loadApps(APISpecs []APISpec, Muxer *http.ServeMux) {
	// load the API defs
	log.Info("Loading API configurations.")

	for _, spec := range APISpecs {
		// Create a new handler for each API spec
		remote, err := url.Parse(spec.APIDefinition.Proxy.TargetURL)
		if err != nil {
			log.Error("Could not parse target URL")
			log.Error(err)
		}

		// TODO: Remove this, testing only
		if spec.UseOauth2 {
			addOAuthHandlers(spec, Muxer)
		}

		proxy := httputil.NewSingleHostReverseProxy(remote)

		proxyHandler := http.HandlerFunc(ProxyHandler(proxy, spec))
		raspberryMiddleware := RaspberryMiddleware{spec, proxy}

		chain := alice.New(
			VersionCheck{raspberryMiddleware}.New(),
			KeyExists{raspberryMiddleware}.New(),
			Oauth2KeyExists{raspberryMiddleware}.New(),
			KeyExpired{raspberryMiddleware}.New(),
			AccessRightsCheck{raspberryMiddleware}.New(),
			RateLimitAndQuotaCheck{raspberryMiddleware}.New()).Then(proxyHandler)
		Muxer.Handle(spec.Proxy.ListenPath, chain)
	}
}

// ReloadURLStructure will create a new muxer, reload all the app configs for an
// instance and then replace the DefaultServeMux with the new one, this enables a
// reconfiguration to take place without stopping and requests from being handled.
func ReloadURLStructure() {
	newMuxes := http.NewServeMux()
	loadAPIEndpoints(newMuxes)
	specs := getAPISpecs()
	loadApps(specs, newMuxes)

	http.DefaultServeMux = newMuxes
	log.Info("Reload complete")
}

func main() {
	intro()
	displayConfig()

	if doMemoryProfile {
		log.Info("Memory profiling active")
		profileFile, _ = os.Create("raspberry.mprof")
		defer profileFile.Close()
	}

	targetPort := fmt.Sprintf(":%d", config.ListenPort)
	loadAPIEndpoints(http.DefaultServeMux)

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
		specs := getAPISpecs()
		loadApps(specs, http.DefaultServeMux)
		go http.Serve(l, nil)
	} else {
		// Resume accepting connextions in a new goroutine.
		log.Panicln("Resuming listening on", l.Addr())
		specs := getAPISpecs()
		loadApps(specs, http.DefaultServeMux)
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
