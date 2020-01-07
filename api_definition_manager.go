package main

import (
	"encoding/json"
	"github.com/RangelReale/osin"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// APIDefinition represents the configuration for a single proxied API and it's versions.
type APIDefinition struct {
	ID         bson.ObjectId `bson:"_id,omitempty" json:"id"`
	Name       string        `bson:"name" json:"name"`
	APIID      string        `bson:"api_id" json:"api_id"`
	OrgID      string        `bson:"org_id" json:"org_id"`
	UseOauth2  bool          `bson:"use_oauth2" json:"use_oauth2"`
	Oauth2Meta struct {
		AllowedAccessTypes    []osin.AllowedAccessType    `bson:"allowed_access_types" json:"allowed_access_types"`
		AllowedAuthorizeTypes []osin.AllowedAuthorizeType `bson:"allowed_authorize_types" json:"allowed_authorize_types"`
	} `bson:"oauth2_meta" json:"oauth2_meta"`
	VersionDefinition struct {
		Location string `bson:"location" json:"location"`
		Key      string `bson:"key" json:"key"`
	} `bson:"definition" json:"definition"`
	VersionData struct {
		NotVersioned bool                   `bson:"not_versioned" json:"not_versioned"`
		Versions     map[string]VersionInfo `bson:"versions" json:"versions"`
	} `bson:"version_data" json:"version_data"`
	Proxy struct {
		ListenPath      string `bson:"listen_path" json:"listen_path"`
		TargetURL       string `bson:"target_url" json:"target_url"`
		StripListenPath bool   `bson:"strip_listen_path" json:"strip_listen_path"`
	} `bson:"proxy" json:"proxy"`
	Auth struct {
		AuthHeaderName string `bson:"auth_header_name" json:"auth_header_name"`
	} `bson:"auth" json:"auth"`
	Active bool `bson:"active" json:"active"`
}

// VersionInfo encapsulates all the data for a specific api_version, elements in the
// Paths array are checked as part of the proxy routing.
type VersionInfo struct {
	Name    string `bson:"name" json:"name"`
	Expires string `bson:"expires" json:"expires"`
	Paths   struct {
		Ignored   []string `bson:"ignored" json:"ignored"`
		WhiteList []string `bson:"white_list" json:"white_list"`
		BlackList []string `bson:"black_list" json:"black_list"`
	} `bson:"paths" json:"paths"`
}

// URLStatus is a custom enum type to avoid collisions
type URLStatus int

// Enums representing the various statuses for a VersionInfo Path match during a
// proxy requst
const (
	Ignored   URLStatus = 1
	WhiteList URLStatus = 2
	BlackList URLStatus = 3
)

// RequestStatus is a custom type to avoid collisions
type RequestStatus string

// Statuses of the request, all are false-y except StatusOK and StatusOkAndIgnore
const (
	VersionNotFound                RequestStatus = "Version information not found"
	VersionDoesNotExist            RequestStatus = "This API version doesn't seem to exist"
	VersionPathsNotFound           RequestStatus = "Path information could not be foound for version"
	VersionWhiteListStatusNotFound RequestStatus = "WhiteListStatus for path not found"
	VersionExpired                 RequestStatus = "Api Version has expired, please check documentation or contack administractor"
	EndPointNotAllowed             RequestStatus = "Requested endpoint is forbidden"
	GeneralFailure                 RequestStatus = "An error occurred that should have not been possible"
	StatusOkAndIgnore              RequestStatus = "Everything OK, passing and not filtering"
	StatusOK                       RequestStatus = "Everything OK, passing"
)

// URLSpec repersents a flattened specification for URLs, used to check if a proxy URL
// path is on any of the white, plack or ignored lists. This is generated as part of the
// configuration init
type URLSpec struct {
	Spec   *regexp.Regexp
	Status URLStatus
}

// APISpec represents a path sepcification for an API, to avoid enumerating multiple nested lists, a single
// flattened URL list is checked for matching paths and then it's status evaluated if found.
type APISpec struct {
	APIDefinition
	RxPaths          map[string][]URLSpec
	WhiteListEnabled map[string]bool
}

// APIDefinitionLoader will load an Api definition from a storage system. It has two methods LoadDefinitionsFromMongo()
// and LoadDefinitions, each will pull api specifications from different locations.
type APIDefinitionLoader struct {
	dbSession *mgo.Session
}

// Connect connects to the storage engine - can be null
func (a *APIDefinitionLoader) Connect() {
	var err error
	a.dbSession, err = mgo.Dial(config.AnalyticsConfig.MongoURL)
	if err != nil {
		log.Error("Mongo connection failed:")
		log.Panic(err)
	}
}

// MakeSpec will generate a flattened URLSpec from and APIDefinitions' VersionInfo data. paths are
// keyed to the Api version name, which is determined during routing to speed up lookups
func (a *APIDefinitionLoader) MakeSpec(thisAppConfig APIDefinition) APISpec {
	newAppSpec := APISpec{}
	newAppSpec.APIDefinition = thisAppConfig
	newAppSpec.RxPaths = make(map[string][]URLSpec)
	newAppSpec.WhiteListEnabled = make(map[string]bool)
	for _, v := range thisAppConfig.VersionData.Versions {
		pathSpecs, whiteListSpecs := a.getPathSpecs(v)
		newAppSpec.RxPaths[v.Name] = pathSpecs
		newAppSpec.WhiteListEnabled[v.Name] = whiteListSpecs
	}

	return newAppSpec
}

// LoadDefinitionsFromMongo will connect and download ApiDefinitions from a Mongo DB instance.
func (a *APIDefinitionLoader) LoadDefinitionsFromMongo() []APISpec {
	var APISpecs = []APISpec{}

	a.Connect()
	apiCollection := a.dbSession.DB("").C("raspberry_apis")

	search := bson.M{
		"active": true,
	}

	var APIDefinitions = []APIDefinition{}
	mongoErr := apiCollection.Find(search).All(&APIDefinitions)

	if mongoErr != nil {
		log.Error("Could not find any application configs!")
		return APISpecs
	}

	for _, thisAppConfig := range APIDefinitions {
		// Got the configuration, build the spec!
		newAppSpec := a.MakeSpec(thisAppConfig)
		APISpecs = append(APISpecs, newAppSpec)
	}
	return APISpecs
}

// LoadDefinitions will load APIDefintions from a directory on the filesystem. definitions need
// to be the JSON representation of APIDefinition object
func (a *APIDefinitionLoader) LoadDefinitions(dir string) []APISpec {
	var APISpecs = []APISpec{}
	// Grab json files from directory
	files, _ := ioutil.ReadDir(dir)
	for _, f := range files {
		if strings.Contains(f.Name(), ".json") {
			filePath := filepath.Join(dir, f.Name())
			log.Info("Loading API Specification from ", filePath)
			appConfig, err := ioutil.ReadFile(filePath)
			if err != nil {
				log.Error("Couldn't load app configuration file")
				log.Error(err)
			} else {
				thisAppConfig := APIDefinition{}
				err := json.Unmarshal(appConfig, &thisAppConfig)
				if err != nil {
					log.Error("Couldn't unmarshal api configuration")
					log.Error(err)
				} else {
					// Got the configuration, build the spec!
					newAppSpec := a.MakeSpec(thisAppConfig)
					APISpecs = append(APISpecs, newAppSpec)
				}
			}
		}
	}

	return APISpecs
}

func (a *APIDefinitionLoader) getPathSpecs(apiVersionDef VersionInfo) ([]URLSpec, bool) {
	ignoredPaths := a.compilePathSpec(apiVersionDef.Paths.Ignored, Ignored)
	blackListPaths := a.compilePathSpec(apiVersionDef.Paths.BlackList, BlackList)
	whiteListPaths := a.compilePathSpec(apiVersionDef.Paths.WhiteList, WhiteList)

	combinedPath := []URLSpec{}
	combinedPath = append(combinedPath, ignoredPaths...)
	combinedPath = append(combinedPath, blackListPaths...)
	combinedPath = append(combinedPath, whiteListPaths...)

	if len(whiteListPaths) > 0 {
		return combinedPath, true
	}
	return combinedPath, false
}

func (a *APIDefinitionLoader) compilePathSpec(paths []string, specType URLStatus) []URLSpec {
	// transform a configuration URL into an array of URLSpecs
	// This way we can interate the whole array once, on match we break with status
	apiLandIDsRegex, _ := regexp.Compile("{(.*?)}")
	thisURLSpec := []URLSpec{}

	for _, stringSpec := range paths {
		asRegexStr := apiLandIDsRegex.ReplaceAllString(stringSpec, "{(.*?)}")
		asRegex, _ := regexp.Compile(asRegexStr)

		newSpec := URLSpec{}
		newSpec.Spec = asRegex
		newSpec.Status = specType
		thisURLSpec = append(thisURLSpec, newSpec)
	}

	return thisURLSpec
}

// IsURLAllowedAndIgnored checks if a url allowed and ignored.
func (a *APISpec) IsURLAllowedAndIgnored(url string, RxPaths []URLSpec, WhiteListStatus bool) (bool, bool) {
	// Check if ignored
	for _, v := range RxPaths {
		match := v.Spec.Match([]byte(url))
		if match {
			if v.Status == Ignored {
				// Let it pass, and do not check auth
				return true, true
			} else if v.Status == BlackList {
				// Matched a blacklist Url, disallow access and filter (irrelevant here)
				return false, false
			} else if v.Status == WhiteList {
				// Matched a whitelist, allow request but filter
				return true, false
			}

			// Should not occur, something has gone wrong
			log.Error("URL Status was not one of Ignored, BlackList or WhiteList! Blocking.")
			return false, false
		}
	}

	// Nothing matched - should we still let it through?
	if WhiteListStatus {
		// We have a whitelist, nothing gets through unless specifically defined
		return false, false
	}

	// No whitelist, but also not in any of the other lists, let is through and filter
	return true, false
}

func (a *APISpec) getVersionFromRequest(r *http.Request) string {
	if a.APIDefinition.VersionDefinition.Location == "header" {
		versionHeaderVal := r.Header.Get(a.APIDefinition.VersionDefinition.Key)
		if versionHeaderVal != "" {
			return versionHeaderVal
		}

		return ""
	} else if a.APIDefinition.VersionDefinition.Location == "url-param" {
		formParam := r.FormValue(a.APIDefinition.VersionDefinition.Key)
		if formParam != "" {
			return formParam
		}

		return ""
	} else {
		return ""
	}

	return ""
}

// IsThisAPIVersionExpired checks if an API version (during a proxied request) is expired
func (a *APISpec) IsThisAPIVersionExpired(versionDef VersionInfo) bool {
	// Never expores
	if versionDef.Expires == "-1" {
		return false
	}

	if versionDef.Expires == "" {
		return false
	}

	// otherwise - calculate the time
	t, err := time.Parse("2006-01-02 15:04", versionDef.Expires)
	if err != nil {
		log.Error("Could not parse expiry date for API, dissallow")
		log.Error(err)
		return true
	}

	remaining := time.Since(t)
	if remaining < 0 {
		// It's in the future, keep going
		return false
	}

	// It's in the past, expire
	return true
}

// IsRequestValid will check if an incoming request has valid version data and return a RequestStatus that
// describes the status of the request
func (a *APISpec) IsRequestValid(r *http.Request) (bool, RequestStatus) {
	versionMetaData, versionPaths, whiteListStatus, stat := a.GetVersionData(r)

	// Screwed up version info - fail and pass through
	if stat != StatusOK {
		return false, stat
	}

	// Is the API version expired?
	if a.IsThisAPIVersionExpired(versionMetaData) {
		// Expired - fail
		return false, VersionExpired
	}

	// not expired, let's check path info
	allowURL, ignore := a.IsURLAllowedAndIgnored(r.URL.Path, versionPaths, whiteListStatus)
	if !allowURL {
		return false, EndPointNotAllowed
	}

	if ignore {
		return true, StatusOkAndIgnore
	}

	return true, StatusOK
}

// GetVersionData attempts to extract the version data from a request, depending on where it is stored in the
// request (currently only "header" is supported)
func (a *APISpec) GetVersionData(r *http.Request) (VersionInfo, []URLSpec, bool, RequestStatus) {
	var thisVersion = VersionInfo{}
	var versionKey string
	var versionRxPaths = []URLSpec{}
	var versionWLStatus bool

	// Are we versioned?
	if a.APIDefinition.VersionData.NotVersioned {
		// Get the first one in the list
		for k, v := range a.APIDefinition.VersionData.Versions {
			versionKey = k
			thisVersion = v
			break
		}
	} else {
		// Extract Version Info
		versionKey = a.getVersionFromRequest(r)
		if versionKey == "" {
			return thisVersion, versionRxPaths, versionWLStatus, VersionNotFound
		}
	}

	// Load Version Data - General
	var ok bool
	thisVersion, ok = a.APIDefinition.VersionData.Versions[versionKey]
	if !ok {
		return thisVersion, versionRxPaths, versionWLStatus, VersionDoesNotExist
	}

	// Load path data and whitelist data for version
	RxPaths, rxOk := a.RxPaths[versionKey]
	WhiteListStatus, wlOK := a.WhiteListEnabled[versionKey]

	if !rxOk {
		log.Error("no RX Paths found for version")
		log.Error(versionKey)
		return thisVersion, versionRxPaths, versionWLStatus, VersionDoesNotExist
	}

	if !wlOK {
		log.Error("No whitelist data found")
		return thisVersion, versionRxPaths, versionWLStatus, VersionWhiteListStatusNotFound
	}

	versionRxPaths = RxPaths
	versionWLStatus = WhiteListStatus

	return thisVersion, versionRxPaths, versionWLStatus, StatusOK
}
