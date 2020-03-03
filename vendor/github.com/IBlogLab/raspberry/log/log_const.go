package log

// LogLevel defines a alice of string.
var LogLevel string

const (
	// RaspberryLogLevel Constants defined for the system log level.
	// this variable will be set in the os env.
	RaspberryLogLevel LogLevel = "RASPBERRY_LOGLEVEL"

	// Error can fetch by os.GetEnv(RASPBERRY_LOGLEVEL)
	Error LogLevel = "error"
	// Wran can fetch by os.GetEnv(RASPBERRY_LOGLEVEL)
	Wran LogLevel = "warn"
	// Debug can fetch by os.GetEnv(RASPBERRY_LOGLEVEL)
	Debug LogLevel = "debug"
)
