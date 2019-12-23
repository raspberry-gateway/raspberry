package main

import (
	"net/http"
	"regexp"
	"time"
)

type ApiDefinition struct {
	Name              string `json:"name"`
	VersionDefinition struct {
		Location string `json:"location"`
		Key      string `json:"key"`
	} `json:"version_definition"`
	VersionData struct {
		NotVersioned bool                   `json:"not_versioned"`
		Versions     map[string]VersionInfo `json:"versions"`
	} `json:"version_data"`
}

type VersionInfo struct {
	Name    string `json:"name"`
	Expires string `json:"expires"`
	Proxy   struct {
		ListenPath string `json:"listen_path"`
		TargetUrl  string `json:"target_url"`
	} `json:"proxy"`
	Auth struct {
		AuthHeaderName string `json:"auth_header_name"`
	} `json:"auth"`
	Paths struct {
		Igored    []string `json:"igored"`
		WhiteList []string `json:"white_list"`
		BlackList []string `json:"black_list"`
	}
}

type UrlStatus int

const (
	Igored    UrlStatus = 1
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

type ApiDefinitionLoader struct{}

func (a *ApiDefinitionLoader) LoadDefinitions(definitionFolder string) []ApiSpec {
	// Grab json files from directory
	// Check if white list and black list are being used (only one can be used at a time)
	// If White list is being used, set flag so we don't let requests through unless on list
	// If only black list and ignored are being used, all requests are filtered except these ones.

	return []ApiSpec{}
}

func (a *ApiDefinitionLoader) getPathSpecs(apiVersionDef VersionInfo) ([]UrlSpec, bool) {
	ignoredPaths := a.CompilePathSpec(apiVersionDef.Paths.Igored, Igored)
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
			if v.Status == Igored {
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
	} else if a.ApiDefinition.VersionDefinition.Location == "url" {
		// TODO - URL MATCH
		return ""
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

	return thisVersion, versionRxPaths, versionWLStatus, GeneralFailure
}

/*

Get /api/blah version=1
Extract Version Info
v1
Load Version Data for v1
Pull RxPaths for version info
Pull White-list bool for versionInfo
Copare route to RxPaths
Route matched? -> Check status (White, black, ignore)
Route not matched -> Check Whitelist status -> action
*/
