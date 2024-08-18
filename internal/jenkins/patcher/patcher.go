package patcher

import (
	"strings"

	"go.uber.org/zap"

	"github.com/kruftik/jenkins-update-dot-json-resigner/internal/jenkins/types"
)

type Patcher interface {
	Patch(insecureJSON *types.InsecureUpdateJSON) error
}

var (
	_ Patcher = Service{}
)

type Service struct {
	log *zap.SugaredLogger

	from, to string
}

func NewPatcher(log *zap.SugaredLogger, from, to string) Service {
	return Service{
		log: log,

		from: from,
		to:   to,
	}
}

func (s Service) Patch(insecureJSON *types.InsecureUpdateJSON) error {
	s.log.Debug("Patching JSON content...")

	// Patch URL in Core section
	insecureJSON.Core.URL = strings.ReplaceAll(insecureJSON.Core.URL, s.from, s.to)
	s.log.Debug("Core URL patched")

	// and plugins download URLs
	for pluginName, pluginInfo := range insecureJSON.Plugins {
		pluginInfo.URL = strings.ReplaceAll(insecureJSON.Plugins[pluginName].URL, s.from, s.to)

		insecureJSON.Plugins[pluginName] = pluginInfo
	}

	s.log.Debug("Plugin URLs patched")

	return nil
}
