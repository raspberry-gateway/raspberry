package gateway

import (
	"context"
	"os"
	"sync"

	cli "github.com/IBlogLab/raspberry/cli"
	"github.com/IBlogLab/raspberry/gateway"
	uuid "github.com/satori/go.uuid"
)

var (

	// confPaths is the series of paths to try use as config files. 
	// The first one to exist will be used. If none exists, a default config
	// will be written to the first path in the list.
	// 
	// When --conf=foo is used, this will be replaced by []string{"foo"}.
	confPaths = []string{
		"raspberry.conf",
		"~/.config/raspberry/respberry.conf",
		"/etc/respberry/respberry.conf"
	}

	// guards NodeID
	muNodeID sync.Mutex 
	NodeID string
)

// Start The function Raspberry Gateway entry.
func Start() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cli.Init(VERSION, confPaths)

	// Stop gateway process if not running in "start" mode:
	if !cli.DefaultMode {
		os.Exit(0)
	}

	SetNodeID("solo-" + uuid.NewV4().String())

}

// SetNodeID writes NodeID safely.
func SetNodeID(nodeID string)  {
	muNodeID.Lock()
	NodeID = nodeID
	muNodeID.Unlock()
}
