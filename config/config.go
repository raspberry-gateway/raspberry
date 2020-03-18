package config

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kelseyhightower/envconfig"
	logger "github.com/raspberry-gateway/raspberry/log"
)

type IPsHandleStrategy string

var (
	log      = logger.Get()
	global   atomic.Value
	globalMu sync.Mutex

	Default = Config{
		ListenPort:   8080,
		Secret:       "352d20ee67be67f6340b4c0605b044b7",
		TemplatePath: "templates",
	}
)

const (
	envPrefix = "RASPBERRY_GW"

	dnsCacheDefaultTtl           = 3600
	dnsCacheDefaultCheckInterval = 60

	PickFirstStrategy IPsHandleStrategy = "pick_first"
	RandomStrategy    IPsHandleStrategy = "random"
	NoCacheStrategy   IPsHandleStrategy = "no_cache"

	DefalutDashPolicySource     = "service"
	DefaultDashPolicyRecordName = "raspberry_policies"
)

type PoliciesConfig struct {
	PolicySource           string `json:"policy_source"`
	PolicyConnectionString string `json:"policy_connection_string"`
	PolicyRecordName       string `json:"policy_record_name"`
	AllowExplicitPolicyID  string `json:"allow_explicit_policy_id"`
}

// DBAppConfOptionsConfig definited for DB
type DBAppConfOptionsConfig struct {
	ConnectionString string   `json:"connection_string"`
	NodeIsSegmented  bool     `json:"node_is_segmented"`
	Tags             []string `json:"tags"`
}

// StorageOptionsConf definited for storage
type StorageOptionsConf struct {
	Type                  string            `json:"type"`
	Host                  string            `json:"host"`
	Port                  int               `json:"port"`
	Hosts                 map[string]string `json:"hosts"`
	Addrs                 []string          `json:"addrs"`
	MasterName            string            `json:"master_name"`
	Username              string            `json:"username"`
	Password              string            `json:"password"`
	Database              int               `json:"database"`
	MaxIdle               int               `json:"optimisation_max_idle"`
	MaxActive             int               `json:"optimisation_max_active"`
	Timeout               int               `json:"timeout"`
	EnableCluster         bool              `json:"enable_cluster"`
	UseSSL                bool              `json:"use_ssl"`
	SSLInsecureSkipVerify bool              `json:"ssl_insecure_skip_verify"`
}

type NormalisedURLConfig struct {
	Enabled            bool                 `json:"enabled"`
	NormaliseUUIDs     bool                 `json:"normalise_uuids"`
	NormaliseNumbers   bool                 `json:"normalise_numbers"`
	Custom             []string             `json:"custom_patterns"`
	CompiledPatternSet NormaliseURLPatterns `json:"-"` // see analytics.go
}

type NormaliseURLPatterns struct {
	UUIDs  *regexp.Regexp
	IDs    *regexp.Regexp
	Custom []*regexp.Regexp
}

type AnalyticsConfigConfig struct {
	Type                    string              `json:"type"`
	IgnoredIPs              []string            `json:"ignored_ips"`
	EnableDetailedRecording bool                `json:"enable_detailed_recording"`
	EnableGeoIP             bool                `json:"enable_geo_ip"`
	NormaliseUrls           NormalisedURLConfig `json:"normalise_urls"`
	PoolSize                int                 `json:"pool_size"`
	RecordsBufferSize       uint64              `json:"records_buffer_size"`
	StorageExpirationTime   int                 `json:"storage_expiration_time"`
	ignoredIPsCompiled      map[string]bool
}

type HealthCheckConfig struct {
	EnableHealthChecks      bool  `json:"enable_health_checks"`
	HealthCheckValueTimeout int64 `json:"health_check_value_timeouts"`
}

type LivenessCheckConfig struct {
	CheckDuration time.Duration `json:"check_duration"`
}

type DnsCacheConfig struct {
	Enbaled bool  `json:"enabled"`
	TTL     int64 `json:"ttl"`
	// CheckInterval controls cache cleanup interval. By convention shouldn't be exposed to config or env_variable_setup
	CheckInterval            int64             `json:"check_interval"`
	MultipleIPsHandleStategy IPsHandleStrategy `json:"multiple_ips_handle_stategy"`
}

type MonitorConfig struct {
	EnableTriggerMonitors bool               `json:"enable_trigger_monitors"`
	Config                WebHookHandlerConf `json:"configuration"`
	GlobalTriggerLimit    float64            `json:"global_trogger_limit"`
	MonitorUserKeys       bool               `json:"monitor_user_keys"`
	MonitorOrgKeys        bool               `json:"monitor_org_keys"`
}

type WebHookHandlerConf struct {
	Method       string            `bson:"method" json:"method"`
	TargetPath   string            `bson:"target_path" json:"target_path"`
	TemplatePath string            `bson:"template_path" josn:"template_path"`
	HeaderList   map[string]string `bson:"header_map" json:"header_map"`
	EventTimeout int64             `bson:"event_timeout" json:"event_timeout"`
}

type SlaveOptionsConfig struct {
	UseRPC                          bool   `json:"use_roc"`
	UseSSL                          bool   `json:"use_ssl"`
	SSLInsecureSkipVerify           bool   `json:"ssl_insecure_skip_vevify"`
	ConnectionString                string `json:"connection_string"`
	RPCKey                          string `json:"rpc_key"`
	APIKey                          string `json:"api_key"`
	EnableRPCCache                  bool   `json:"enable_rpc_cache"`
	BindToSlugsInsteadOfListenPaths bool   `json:"bind_to_slugs"`
	DisableKeySpaceSync             bool   `json:"disable_key_space_sync"`
	GroupID                         string `json:"group_id"`
	CallTimeout                     int    `json:"call_timeout"`
	PingTimeout                     int    `json:"ping_timeout"`
	RPCPoolSize                     int    `json:"rpc_pool_size"`
}

type LocalSessionCacheConf struct {
	DisableCacheSessionState bool `json:"disable_cache_session_state"`
	CachedSessionTimeout     int  `json:"cached_session_timeout"`
	CacheSessionEviction     int  `json:"cache_session_evication"`
}

type HttpServerOptionsConfig struct {
	OverrideDefaults       bool       `json:"override_defaults"`
	ReadTimeout            int        `json:"read_timeout"`
	WriteTimeout           int        `json:"write_timeout"`
	UseSSL                 bool       `json:'use_ssl'`
	UseLe_SSL              bool       `json:"use_ssl_le"`
	EnableHttp2            bool       `json:"enable_http2"`
	SSLInsecureSkipVerify  bool       `json:"ssl_insecure_skip_verify"`
	EnableWebSockets       bool       `json:"enable_web_sockets"`
	Certificates           []CertData `json:"certificates"`
	SSLCertificates        []string   `json:"ssl_certificates"`
	ServerName             string     `json:"server_name"`
	MinVersion             uint16     `json:"min_version"`
	FlushIntercal          int        `json:"flush_interval"`
	SkipURLCleaning        bool       `json:"skip_url_cleaning"`
	SkipTargetPathEscaping bool       `json:"skip_target_path_escaping"`
	Ciphers                []string   `json:"cliphers"`
}

type AuthOverrideConf struct {
	ForceAuthProvider    bool `json:"force_auth_provider"`
	ForceSessionProvider bool `json:"force_session_provider"`
}

type CertData struct {
	Name     string `json:"name"`
	CertFile string `json:"cert_file"`
	KeyFile  string `key_file`
}

// Config is the configuration object used by raspberry to set up various parameters.
type Config struct {
	// OriginalPath is the path to the config file that was read.
	// If none was found, it's the path to the default config file
	// that was written.
	OriginalPath string `json:"-"`

	HostName             string `json:"host_name"`
	ListenAddress        string `json:"listen_address"`
	ListenPort           int    `json:"listen_port"`
	ControlAPIHostname   string `json:"control_api_hostname"`
	ControlAPIPort       int    `json:"control_api_port"`
	Secret               string `json:"secret"`
	NodeSecret           string `json:"node_secret"`
	PIDFileLocation      string `json:"pid_file_location"`
	AllowInsecureConfigs bool   `json:"allow_insecure_configs"`
	PublicKeyPath        string `json:"public_key_path"`
	AllowRemoteConfig    bool   `json:"allow_remote_config"`
}

// WriteDefault will set conf to the default config and write it to disk
// in path, if the path is non-empty.
func WriteDefault(path string, conf *Config) error {
	_, b, _ := runtime.Caller(0)
	configPath := filepath.Dir(b)
	rootPath := filepath.Dir(cinfigPath)
	Default.TemplatePath = filepath.Join(rootPath, "templates")

	*conf = Default
	if err := envconfig.Process(envPrefix, conf); err != nil {
		return err
	}
	if path == "" {
		return nil
	}
	return WriteConf(path, conf)
}

// WriteConf will write conf file to specified path.
func WriteConf(path string, conf *Config) error {
	bs, err := json.MarshalIndent(conf, "", "    ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, bs, 0644)
}

// Load will load a configuration file, trying each of the paths given
// and using the first one that is a regular file and can be opened.
//
// If none exists, a default config will be written to the first path in
// the list.
//
// An error will be returned only if any of the paths existed but was
// not a valid config file.
func Load(paths []string, conf *Config) error {
	var r io.Reader
	for _, path := range paths {
		f, err := os.Open(path)
		if err == nil {
			r = f
			conf.OriginalPath = path
			break
		}
		if os.IsNotExist(err) {
			continue
		}
		return err
	}
	if r == nil {
		path := paths[0]
		log.Warnf("No config file found, writing default to %s", path)
		if err := WriteDefault(path, conf); err != nil {
			return err
		}
		log.Info("Loading default configuration...")
		return Load([]string{path}, conf)
	}
	if err := json.NewDecoder(r).Decode(&conf); err != nil {
		return fmt.Errorf("couldn't unmarshal config: %v", err)
	}
}
