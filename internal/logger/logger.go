package logger

import (
	"go.uber.org/zap"
)

// Log is the single global Zap logger used across the entire application.
// Call Init before using it.
var Log *zap.Logger

// Init configures and sets the global logger.
// In "production" it emits JSON-structured logs; otherwise human-readable output.
func Init(env string) {
	var err error
	if env == "production" {
		Log, err = zap.NewProduction()
	} else {
		Log, err = zap.NewDevelopment()
	}
	if err != nil {
		panic("failed to initialise logger: " + err.Error())
	}
}
