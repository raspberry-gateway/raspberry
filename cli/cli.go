package cli

import (
	"fmt"
	"io/ioutil"
	"os"
	"tyk/cli/linter"

	logger "github.com/raspberry-gateway/raspberry/log"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

const (
	appName = "raspberry"
	appDesc = "Raspberry Gateway"
)

var (
	// Conf specifies the configuration file path.
	Conf *string
	// Port specifies the listen port.
	Port *string
	// MemProfile enables memory profiling.
	MemProfile *bool
	// CPUProfile enables CPU profiling.
	CPUProfile *bool
	// BlockProfile enables block profiling.
	BlockProfile *bool
	// MutexProfile enables mutex profiling.
	MutexProfile *bool
	// HTTPProfile exposes a HTTP endpoint for accessing profiling.
	HTTPProfile *bool
	// DebugMode sets the log level to debug mode.
	DebugMode *bool
	// LogInstrumentation outputs instrumentation data to stdout.
	LogInstrumentation *bool

	// DefaultMode is set when default command is used.
	DefaultMode bool

	app *kingpin.Application

	log = logger.Get()
)

// Init sets all flags and subcommands.
func Init(version string, confPaths []string) {
	app = kingpin.New(appName, appDesc)
	app.HelpFlag.Short('h')
	app.Version(version)

	// Start/default command:
	startCmd := app.Command("start", "Starts the Gateway")
	Conf = startCmd.Flag("conf", "load a named configuration file").PlaceHolder("FILE").String()
	Port = startCmd.Flag("port", "listen on PORT (override config file)").String()
	MemProfile = startCmd.Flag("memprofile", "generate a memory profile").Bool()
	CPUProfile = startCmd.Flag("cpuprofile", "generate a cpu profile").Bool()
	BlockProfile = startCmd.Flag("blockprofile", "generate a block profile").Bool()
	MutexProfile = startCmd.Flag("mutexprofile", "generate a mutex profile").Bool()
	HTTPProfile = startCmd.Flag("httpprofile", "expose runtime profiling data via HTTP").Bool()
	DebugMode = startCmd.Flag("debug", "enable debug mode").Bool()
	LogInstrumentation = startCmd.Flag("log-instrumentation", "output instrumentation output to stdout").Bool()

	startCmd.Action(func(ctx *kingpin.ParseContext) error {
		DefalutMode = true
		return nil
	})
	startCmd.Default()

	// Linter
	lintCmd := app.Command("lint", "Runs a linter on Raspberry configuration file")
	lintCmd.Action(func(c *kingpin.ParseContext) error {
		confSchema, err := ioutil.ReadFile("cli/linter/schema.json")
		if err != nil {
			return err
		}
		path, lines, err := linter.Run(string(confSchema), confPaths)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if len(lines) == 0 {
			fmt.Printf("found no issues in %s\n", path)
			os.Exit(0)
		}
		fmt.Printf("issues found in %s:\n", path)
		for _, line := range lines {
			fmt.Println(line)
		}
		os.Exit(1)
		return nil
	})

}
