package patcher

import (
	"strings"

	"go.uber.org/zap"

	"github.com/kruftik/jenkins-update-dot-json-resigner/internal/config"
	"github.com/kruftik/jenkins-update-dot-json-resigner/internal/jenkins/types"
)

var (
	_ types.Patcher = Service{}
)

type Service struct {
	log *zap.SugaredLogger

	from, to string
}

func NewPatcher(log *zap.SugaredLogger, cfg config.PatchConfig) Service {
	return Service{
		log: log,

		from: cfg.OriginDownloadURL,
		to:   cfg.NewDownloadURL,
	}
}

func (s Service) Patch(insecureJSON *types.InsecureUpdateJSON) error {
	// Patch URL in Core section
	insecureJSON.Core.URL = strings.ReplaceAll(insecureJSON.Core.URL, s.from, s.to)

	// and plugins download URLs
	for pluginName, pluginInfo := range insecureJSON.Plugins {
		pluginInfo.URL = strings.ReplaceAll(insecureJSON.Plugins[pluginName].URL, s.from, s.to)

		insecureJSON.Plugins[pluginName] = pluginInfo
	}

	return nil
}
