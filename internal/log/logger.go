package logger

import (
	"go.uber.org/zap"
)

var Log *zap.Logger

func Init() {
	config := zap.NewProductionConfig()
    config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)

    logger, err := config.Build()
    if err != nil {
        panic(map[string]any{
			"code": "FAILED_LOGGER",
			"message": err.Error(),
		})
    }

	Log = logger
}

func Close() {
	Log.Sync()
}