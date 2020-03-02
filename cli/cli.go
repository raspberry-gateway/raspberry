package cli

import (
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

	log = logger.GET()
)
