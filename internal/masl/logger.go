package masl

import (
	"fmt"
	"os"
	"os/user"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var zaplogger *zap.Logger
var once sync.Once

// start loggeando
func GetInstance() *zap.Logger {
	once.Do(func() {
		usr, err := user.Current()
		if err != nil {
			fmt.Printf("\n%s", err.Error())
			os.Exit(1)
		}
		cfg := zap.Config{
			Encoding:         "console",
			ErrorOutputPaths: []string{"stderr"},
			EncoderConfig:    zap.NewDevelopmentEncoderConfig(),
			Level:            zap.NewAtomicLevelAt(zapcore.InfoLevel),
			OutputPaths: []string{
				usr.HomeDir + string(os.PathSeparator) + ".masl" + string(os.PathSeparator) + "masl.log",
			},
		}
		zaplogger, err = cfg.Build()
		if err != nil {
			zaplogger, _ = zap.NewProduction()
		}

	})
	return zaplogger
}
