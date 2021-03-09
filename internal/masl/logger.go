package masl

import (
	"fmt"
	"os"
	"os/user"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var cfg zap.Config
var zapLogger *zap.Logger
var once sync.Once

func GetLogger(level string) *zap.Logger {
	once.Do(func() {
		usr, err := user.Current()
		if err != nil {
			fmt.Printf("\n%s", err.Error())
			os.Exit(1)
		}
		var zapLevel zap.AtomicLevel
		if level == "debug" {
			zapLevel = zap.NewAtomicLevelAt(zapcore.DebugLevel)
		} else {
			zapLevel = zap.NewAtomicLevelAt(zapcore.InfoLevel)
		}
		cfg = zap.Config{
			Encoding:         "console",
			ErrorOutputPaths: []string{"stderr"},
			EncoderConfig:    zap.NewDevelopmentEncoderConfig(),
			Level:            zapLevel,
			OutputPaths: []string{
				usr.HomeDir + string(os.PathSeparator) + ".masl" + string(os.PathSeparator) + "masl.log",
			},
		}
		zapLogger, err = cfg.Build()
		if err != nil {
			zapLogger, _ = zap.NewProduction()
		}
		defer func() {
			_ = zapLogger.Sync()
		}()
	})
	return zapLogger
}
