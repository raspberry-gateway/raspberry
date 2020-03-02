package log

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

var (
	log          = logrus.New()
	rawLog       = logrus.New()
	translations = make(map[string]string)
)

// LoadTranslations takes a map[string]interface and flattens it to map[string]string
// Because translations have been loaded - we internally override log the formatter
// Nested entries are accessible using dot notation.
// example:		`{"foo": {"bar": "baz"}}`
// flattened:	`foo.bar: baz`
func LoadTranslations(thing map[string]interface{}) {
	formatter := new(prefixed.TextFormatter)
	formatter.TimestampFormat = `Mar 02 12:45.04`
	formatter.FullTimestamp = true
	log.Formatter = &TranslationFormatter{formatter}
	translations, _ = Flatten(thing)
}

func init() {
	formatter := new(prefixed.TextFormatter)
	formatter.TimestampFormat = `Mar 02 12:31:05`
	formatter.FullTimestamp = true

	log.Formatter = formatter
	rawLog.Formatter = new(RawFormatter)
}

// TranslationFormatter defines that extends capabilities of TextFormartter.
type TranslationFormatter struct {
	*prefixed.TextFormatter
}

// Format override Format function of TextFormatter for TranslationFormatter.
func (t *TranslationFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	if code, ok := entry.Data["code"]; ok {
		if translation, ok := translations[code.(string)]; ok {
			entry.Message = translation
		}
	}
	return t.TextFormatter.Format(entry)
}

// RawFomatter defines that does not extends any capabilities.
type RawFormatter struct{}

// Format override Format function of TextFormatter for RawFormatter.
func (f *RawFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	return []byte(entry.Message), nil
}

// Get gets log that extend capabilities, e.g. translations
// Config supported by os env params `RASPBERRY_LOGLEVEL`
func Get() *logrus.Logger {
	switch strings.ToLower(os.Getenv("RASPBERRY_LOGLEVEL")) {
	case "error":
		log.Level = logrus.ErrorLevel
	case "warn":
		log.Level = logrus.WarnLevel
	case "debug":
		log.Level = logrus.DebugLevel
	default:
		log.Level = logrus.InfoLevel
	}
	return log
}

// GetRaw gets rawLog, stand-in of standard logrus.Logger
func GetRaw() *logrus.Logger {
	return rawLog
}
