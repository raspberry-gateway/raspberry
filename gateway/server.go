package gateway

import (
	"context"
	"io/ioutil"
	"os"
	"sync"

	cli "github.com/raspberry-gateway/raspberry/cli"
	"github.com/raspberry-gateway/raspberry/config"
	logger "github.com/raspberry-gateway/raspberry/log"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
)

var (
	log       = logger.Get()
	mainLog   = logger.Get().WithField("prefix", "main")
	pubSubLog = logger.Get().WithField("prefix", "pub-sub")
	rawLog    = logger.GetRaw()

	// confPaths is the series of paths to try use as config files.
	// The first one to exist will be used. If none exists, a default config
	// will be written to the first path in the list.
	//
	// When --conf=foo is used, this will be replaced by []string{"foo"}.
	confPaths = []string{
		"raspberry.conf",
		"~/.config/raspberry/respberry.conf",
		"/etc/respberry/respberry.conf",
	}

	muNodeID sync.Mutex // guards NodeID
	// NodeID for current app
	NodeID string

	runningTestsMu sync.RWMutex
	testMode       bool
)

// Start The function Raspberry Gateway entry.
func Start() {
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	cli.Init(VERSION, confPaths)

	// Stop gateway process if not running in "start" mode:
	if !cli.DefaultMode {
		os.Exit(0)
	}

	SetNodeID("solo-" + uuid.NewV4().String())

}

// SetNodeID writes NodeID safely.
func SetNodeID(nodeID string) {
	muNodeID.Lock()
	NodeID = nodeID
	muNodeID.Unlock()
}

func initialliseSystem(ctx context.Context) error {
	if isRunningTests() && os.Getenv(logger.LogLevel) == "" {
		// `go test` without RASPBERRY_LOGLEVEL set defaults to no log output
		log.Level = logrus.ErrorLevel
		log.Out = ioutil.Discard
	} else if *cli.DebugMode {
		log.Level = logrus.DebugLevel
		mainLog.Debug("Enabling debug-level output")
	}

	if *cli.Conf != "" {
		mainLog.Debugf("Using %s fro configuration", *cli.Conf)
		confPaths = []string{*cli.Conf}
	} else {
		mainLog.Debug("No configuration file defined, with try to use default (raspberry.conf)")
	}

	mainLog.Infof("Raspberry API gateway %s", VERSION)

	if !isRunningTests() {
		globalConf := config.Config{}
	}
}

func isRunningTests() bool {
	runningTestsMu.RLock()
	v := testMode
	runningTestsMu.RUnlock()
	return v
}
