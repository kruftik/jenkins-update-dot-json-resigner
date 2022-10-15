package logging

import (
	"go.uber.org/zap"
	"jenkins-resigner-service/internal/config"
)

var (
	logger *zap.Logger
	log *zap.SugaredLogger
)

func Configure(cfg config.Config) (*zap.SugaredLogger, error) {
	if config.Opts.Dbg {
		logger, _ = zap.NewDevelopment()
	} else {
		logger, _ = zap.NewProduction()
	}
	//defer func() {
	//	_ = logger.Sync()
	//}()

	zap.ReplaceGlobals(logger)
	log = zap.S()

	return log, nil
}

func GetLogger() *zap.SugaredLogger {
	return log
}
