package init

import (
	"flag"
	"path/filepath"
	"strings"

	"github.com/mattn/go-colorable"
	"github.com/momper14/rotatefilehook"
	"github.com/momper14/viperfix"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// init viper
func init() {
	configfile := flag.String("config", "config.yml", "file with config")
	flag.Parse()

	viper.SetConfigName(*configfile)
	viper.SetConfigType(filepath.Ext(*configfile)[1:])
	viper.AddConfigPath(".")
	err := viper.ReadInConfig() // Find and read the config file
	if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		logrus.Warn(err)
	} else if err != nil {
		logrus.Fatal(err)
	}

	viper.SetEnvPrefix("MSW")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	viper.SetTypeByDefaultValue(true)
}

func strToLogrusLevel(str string) logrus.Level {

	var m = map[string]logrus.Level{
		"panic": logrus.PanicLevel,
		"fatal": logrus.FatalLevel,
		"error": logrus.ErrorLevel,
		"warn":  logrus.WarnLevel,
		"info":  logrus.InfoLevel,
		"debug": logrus.DebugLevel,
		"trace": logrus.TraceLevel,
	}

	if val, ok := m[str]; ok {
		return val
	}

	return logrus.InfoLevel
}

// define config
func init() {
	viper.SetDefault("log.filename", "logs/latest.log")
	viper.SetDefault("log.compress", true)
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.timestampformat", "02-01-2006 15:04:05")
	viper.SetDefault("log.max.size", 50)
	viper.SetDefault("log.max.backups", 5)
	viper.SetDefault("log.max.age", 31)
}

// init logrus
func init() {
	var c struct {
		Filename        string
		Compress        bool
		Level           string
		Timestampformat string
		Max             struct {
			Size    int
			Backups int
			Age     int
		}
	}

	if err := viperfix.UnmarshalKey("log", &c); err != nil {
		logrus.Fatal(err)
	}

	var logLevel = strToLogrusLevel(c.Level)
	var timestampFormat = c.Timestampformat

	rotateFileHook, err := rotatefilehook.NewRotateFileHook(rotatefilehook.RotateFileConfig{
		Filename:   c.Filename,
		MaxSize:    c.Max.Size,
		MaxBackups: c.Max.Backups,
		MaxAge:     c.Max.Age,
		Compress:   c.Compress,
		Level:      logLevel,
		Formatter: &logrus.TextFormatter{
			DisableColors:   true,
			FullTimestamp:   true,
			TimestampFormat: timestampFormat,
		},
	})

	if err != nil {
		logrus.Fatalf("Failed to initialize file rotate hook: %v", err)
	}

	logrus.SetLevel(logLevel)
	logrus.SetOutput(colorable.NewColorableStdout())
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: timestampFormat,
	})
	logrus.AddHook(rotateFileHook)
}
