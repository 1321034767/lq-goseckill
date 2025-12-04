package log

import (
	"os"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var logger *zap.Logger

func InitLogger() {
	// Defines log file output configuration
	logFile := zapcore.AddSync(&lumberjack.Logger{
		Filename:   "./logs/app.log",
		MaxSize:    500, // Megabytes
		MaxBackups: 3,
		MaxAge:     28, // Days
		Compress:   true,
	})

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "time"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Determines the log level based on the configuration
	logLevel := zap.InfoLevel
	isDeveloperMode := viper.GetBool("log.developerMode")
	if isDeveloperMode {
		logLevel = zap.DebugLevel
		// Includes the caller function information in the log
		encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	}

	cores := []zapcore.Core{
		zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), logFile, logLevel),
	}

	// Adds console output if configured
	if viper.GetBool("log.outputToConsole") {
		consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
		consoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(zapcore.Lock(os.Stdout)), logLevel)
		cores = append(cores, consoleCore)
	}

	// Initializes the logger with all cores
	if isDeveloperMode {
		logger = zap.New(zapcore.NewTee(cores...), zap.AddCaller())
	} else {
		logger = zap.New(zapcore.NewTee(cores...))
	}

	// Replaces the global logger with this configured one
	zap.ReplaceGlobals(logger)
}
