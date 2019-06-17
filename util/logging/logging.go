package logging

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LogLevel - user-settable environment key name
const LogLevel = "LOG_LEVEL"

var logger *zap.Logger

func init() {
	var err error
	// "encoding": "json",
	//
	rawJSON := []byte(`{
			"level": "debug",
			"encoding": "json",
			"outputPaths": ["stdout"],
			"errorOutputPaths": ["stderr"],
			"encoderConfig": {
			  "messageKey": "msg",
			  "levelKey": "level",
			  "levelEncoder": "capital",
			  "timeKey": "time",
			  "timeEncoder": "iso8601"
			}
		  }`)

	var cfg zap.Config
	if err = json.Unmarshal(rawJSON, &cfg); err != nil {
		panic(err)
	}

	logger, err = cfg.Build()
	if err != nil {
		panic(err)
	}

	logLevelStr, found := os.LookupEnv(LogLevel)
	if found {
		logLevel := aToLogLevel(logLevelStr, zapcore.DebugLevel)
		cfg.Level.SetLevel(logLevel)
	}

	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
}

// Logger -
func Logger() *zap.Logger {
	return logger
}

// after we've used 'LookupEnv' and it found something under the key passed (second return value)
// convert the string to zapcore.Level; if no match is found, default to what was passed in as default
func aToLogLevel(levelStr string, defaultLevel zapcore.Level) zapcore.Level {
	if len(levelStr) > 0 {
		switch strings.ToLower(levelStr) {
		case "debug":
			return zap.DebugLevel
		// InfoLevel is the default logging priority.
		case "info":
			return zap.InfoLevel
		// WarnLevel logs are more important than Info, but don't need individual
		// human review.
		case "warn":
			return zap.WarnLevel
		// ErrorLevel logs are high-priority. If an application is running smoothly,
		// it shouldn't generate any error-level logs.
		case "error":
			return zap.ErrorLevel
		// DPanicLevel logs are particularly important errors. In development the
		// logger panics after writing the message.
		case "dpanic":
			return zap.DPanicLevel
		// PanicLevel logs a message, then panics.
		case "panic":
			return zap.PanicLevel
		// FatalLevel logs a message, then calls os.Exit(1).
		case "fatal":
			return zap.FatalLevel
		default:
			return defaultLevel
		}
	}
	return defaultLevel
}
