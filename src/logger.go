package src

import (
	"log"

	"go.uber.org/zap"
)

// NewLogger return new zap.Logger instance with configuration based on the config
func NewLogger(verbose bool) *zap.Logger {
	var logger *zap.Logger

	switch verbose {
	case true:
		devLogger, err := zap.NewDevelopment()
		if err != nil {
			log.Panic("can't initialize logger")
		}
		logger = devLogger
	default:
		prodLogger, err := zap.NewProduction()
		if err != nil {
			log.Panic("can't initialize logger")
		}
		logger = prodLogger
	}
	return logger
}
