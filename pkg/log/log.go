package log

import (
	"os"

	"go.elastic.co/apm/module/apmzap"
	"go.uber.org/zap"
)

func Init(version, description string) *zap.Logger {
	var log *zap.Logger
	var err error
	var wrappedCore = zap.WrapCore((&apmzap.Core{
		FatalFlushTimeout: 10000,
	}).WrapCore)

	if os.Getenv("DEBUG") != "" {
		cfg := zap.NewDevelopmentConfig()
		if os.Getenv("DEBUG") == "true" {
			cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
		}
		cfg.OutputPaths = []string{"stdout", description + "-" + version + ".log"}
		log, err = cfg.Build(wrappedCore)
	} else {
		cfg := zap.NewProductionConfig()
		cfg.OutputPaths = []string{"stdout", description + "-" + version + ".log"}
		log, err = cfg.Build(wrappedCore)
	}

	if err != nil {
		log = zap.NewExample()
		log.Warn("Unable to set up the logger. Replaced with example one which shouldn't fail", zap.Error(err))
	}
	zap.ReplaceGlobals(log)
	err = log.Sync()
	if err != nil {
		log.Warn("Logger sync fail", zap.Error(err))
	}
	log.Debug("Logger is ready")

	return log
}
