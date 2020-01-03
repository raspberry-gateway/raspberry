package main

import (
	"encoding/json"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type ApiDefinition struct {
	Id                bson.ObjectId `bson:"_id,omitempty" json:"id"`
	Name              string        `bson:"name" json:"name"`
	ApiId             string        `bson:"api_id" json:"api_id"`
	OrgId             string        `bson:"org_id" json:"org_id"`
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
		TargetUrl       string `bson:"target_url" json:"target_url"`
		StripListenPath bool   `bson:"strip_listen_path" json:"strip_listen_path"`
	} `bson:"proxy" json:"proxy"`
	Auth struct {
		AuthHeaderName string `bson:"auth_header_name" json:"auth_header_name"`
	} `bson:"auth" json:"auth"`
	Active bool `bson:"active" json:"active"`
}

type VersionInfo struct {
	Name    string `bson:"name" json:"name"`
	Expires string `bson:"expires" json:"expires"`
	Paths   struct {
		Ignored   []string `bson:"ignored" json:"ignored"`
		WhiteList []string `bson:"white_list" json:"white_list"`
		BlackList []string `bson:"black_list" json:"black_list"`
	} `bson:"paths" json:"paths"`
}

type UrlStatus int

const (
	Ignored   UrlStatus = 1
	WhiteList UrlStatus = 2
	BlackList UrlStatus = 3
)

type RequestStatus string

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

type UrlSpec struct {
	Spec   *regexp.Regexp
	Status UrlStatus
}

type ApiSpec struct {
	ApiDefinition
	RxPaths          map[string][]UrlSpec
	WhiteListEnabled map[string]bool
}

type ApiDefinitionLoader struct {
	dbSession *mgo.Session
}

func (a *ApiDefinitionLoader) Connect() {
	var err error
	a.dbSession, err = mgo.Dial(config.AnalyticsConfig.MongoURL)
	if err != nil {
		log.Error("Mongo connection failed:")
		log.Panic(err)
	}
}

func (a *ApiDefinitionLoader) MakeSpec(thisAppConfig ApiDefinition) ApiSpec {
	newAppSpec := ApiSpec{}
	newAppSpec.ApiDefinition = thisAppConfig
	newAppSpec.RxPaths = make(map[string][]UrlSpec)
	newAppSpec.WhiteListEnabled = make(map[string]bool)
	for _, v := range thisAppConfig.VersionData.Versions {
		pathSpecs, whiteListSpecs := a.getPathSpecs(v)
		newAppSpec.RxPaths[v.Name] = pathSpecs
		newAppSpec.WhiteListEnabled[v.Name] = whiteListSpecs
	}

	return newAppSpec
}

func (a *ApiDefinitionLoader) LoadDefinitionsFromMongo() []ApiSpec {
	var ApiSpecs = []ApiSpec{}

	a.Connect()
	apiCollection := a.dbSession.DB("").C("raspberry_apis")

	search := bson.M{
		"active": true,
	}

	var ApiDefinitions = []ApiDefinition{}
	mongoErr := apiCollection.Find(search).All(&ApiDefinitions)

	if mongoErr != nil {
		log.Error("Could not find any application configs!")
		return ApiSpecs
	}

	for _, thisAppConfig := range ApiDefinitions {
		// Got the configuration, build the spec!
		newAppSpec := a.MakeSpec(thisAppConfig)
		ApiSpecs = append(ApiSpecs, newAppSpec)
	}
	return ApiSpecs
}

func (a *ApiDefinitionLoader) LoadDefinitions(dir string) []ApiSpec {
	var ApiSpecs = []ApiSpec{}
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
				thisAppConfig := ApiDefinition{}
				err := json.Unmarshal(appConfig, &thisAppConfig)
				if err != nil {
					log.Error("Couldn't unmarshal api configuration")
					log.Error(err)
				} else {
					// Got the configuration, build the spec!
					newAppSpec := a.MakeSpec(thisAppConfig)
					ApiSpecs = append(ApiSpecs, newAppSpec)
				}
			}
		}
	}

	return ApiSpecs
}

func (a *ApiDefinitionLoader) getPathSpecs(apiVersionDef VersionInfo) ([]UrlSpec, bool) {
	ignoredPaths := a.CompilePathSpec(apiVersionDef.Paths.Ignored, Ignored)
	blackListPaths := a.CompilePathSpec(apiVersionDef.Paths.BlackList, BlackList)
	whiteListPaths := a.CompilePathSpec(apiVersionDef.Paths.WhiteList, WhiteList)

	combinedPath := []UrlSpec{}
	combinedPath = append(combinedPath, ignoredPaths...)
	combinedPath = append(combinedPath, blackListPaths...)
	combinedPath = append(combinedPath, whiteListPaths...)

	if len(whiteListPaths) > 0 {
		return combinedPath, true
	}
	return combinedPath, false
}

func (a *ApiDefinitionLoader) CompilePathSpec(paths []string, specType UrlStatus) []UrlSpec {
	// transform a configuration URL into an array of URLSpecs
	// This way we can interate the whole array once, on match we break with status
	apiLandIdsRegex, _ := regexp.Compile("{(.*?)}")
	thisUrlSpec := []UrlSpec{}

	for _, stringSpec := range paths {
		asRegexStr := apiLandIdsRegex.ReplaceAllString(stringSpec, "{(.*?)}")
		asRegex, _ := regexp.Compile(asRegexStr)

		newSpec := UrlSpec{}
		newSpec.Spec = asRegex
		newSpec.Status = specType
		thisUrlSpec = append(thisUrlSpec, newSpec)
	}

	return thisUrlSpec
}

func (a *ApiSpec) IsUrlAllowedAndIgnored(url string, RxPaths []UrlSpec, WhiteListStatus bool) (bool, bool) {
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
			} else {
				// Should not occur, something has gone wrong
				log.Error("URL Status was not one of Ignored, BlackList or WhiteList! Blocking.")
				return false, false
			}
		}
	}

	// Nothing matched - should we still let it through?
	if WhiteListStatus {
		// We have a whitelist, nothing gets through unless specifically defined
		return false, false
	} else {
		// No whitelist, but also not in any of the other lists, let is through and filter
		return true, false
	}
}

func (a *ApiSpec) getVersionFromRequest(r *http.Request) string {
	if a.ApiDefinition.VersionDefinition.Location == "header" {
		versionHeaderVal := r.Header.Get(a.ApiDefinition.VersionDefinition.Key)
		if versionHeaderVal != "" {
			return versionHeaderVal
		} else {
			return ""
		}
	} else if a.ApiDefinition.VersionDefinition.Location == "url-param" {
		formParam := r.FormValue(a.ApiDefinition.VersionDefinition.Key)
		if formParam != "" {
			return formParam
		} else {
			return ""
		}
	} else {
		return ""
	}

	return ""
}

func (a *ApiSpec) IsThisApiVersionExpired(versionDef VersionInfo) bool {
	// Never expores
	if versionDef.Expires == "-1" {
		return false
	}

	// otherwise - calculate the time
	t, err := time.Parse("2006-01-02 15:04", versionDef.Expires)
	if err != nil {
		log.Error("Could not parse expiry date for API, dissallow")
		log.Error(err)
		return true
	} else {
		remaining := time.Since(t)
		if remaining < 0 {
			// It's in the future, keep going
			return false
		} else {
			// It's in the past, expire
			return true
		}
	}
}

func (a *ApiSpec) IsRequestValid(r *http.Request) (bool, RequestStatus) {
	versionMetaData, versionPaths, whiteListStatus, stat := a.GetVersionData(r)

	// Screwed up version info - fail and pass through
	if stat != StatusOK {
		return false, stat
	}

	// Is the API version expired?
	if a.IsThisApiVersionExpired(versionMetaData) {
		// Expired - fail
		return false, VersionExpired
	}

	// not expired, let's check path info
	allowURL, ignore := a.IsUrlAllowedAndIgnored(r.URL.Path, versionPaths, whiteListStatus)
	if !allowURL {
		return false, EndPointNotAllowed
	}

	if ignore {
		return true, StatusOkAndIgnore
	}

	return true, StatusOK
}

func (a *ApiSpec) GetVersionData(r *http.Request) (VersionInfo, []UrlSpec, bool, RequestStatus) {
	var thisVersion = VersionInfo{}
	var versionKey string
	var versionRxPaths = []UrlSpec{}
	var versionWLStatus bool

	// Are we versioned?
	if a.ApiDefinition.VersionData.NotVersioned {
		// Get the first one in the list
		for k, v := range a.ApiDefinition.VersionData.Versions {
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
	thisVersion, ok = a.ApiDefinition.VersionData.Versions[versionKey]
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
