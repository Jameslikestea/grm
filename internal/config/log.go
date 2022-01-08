package config

import (
	"strings"

	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

const (
	logPath  = "log.path"
	logLevel = "log.level"
	logFile  = "log.file"
)

func GetLogPath() string {
	return viper.GetString(logPath)
}

func SetLogPath(p string) {
	viper.Set(logPath, p)
}

func GetLogLevel() zerolog.Level {
	switch strings.ToUpper(viper.GetString(logLevel)) {
	case "TRACE":
		return zerolog.TraceLevel
	case "DEBUG":
		return zerolog.DebugLevel
	case "INFO":
		return zerolog.InfoLevel
	case "WARN":
		return zerolog.WarnLevel
	case "ERROR":
		return zerolog.ErrorLevel
	case "FATAL":
		return zerolog.FatalLevel
	case "PANIC":
		return zerolog.PanicLevel
	}

	return zerolog.InfoLevel
}

func SetLogLevel(l zerolog.Level) {
	viper.Set(logLevel, l.String())
}

func GetLogFile() bool {
	return viper.GetBool(logFile)
}

func SetLogFile(f bool) {
	viper.Set(logFile, f)
}
