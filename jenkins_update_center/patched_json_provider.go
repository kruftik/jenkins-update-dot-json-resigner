package jenkins_update_center

import (
	"jenkins-resigner-service/jenkins_update_center/json_schema"
	"strings"
	"time"
)

var (
	patchedJSONCacheTTL = 30 * time.Minute
)

type PatchedJSONProvider struct {
	origin   JSONProvider
	metadata *JSONMetadataT
}

func (p *PatchedJSONProvider) patchContent(from, to string) (*json_schema.InsecureUpdateJSON, error) {
	signedObj, err := p.origin.GetContent()
	if err != nil {
		return nil, err
	}

	c := json_schema.InsecureUpdateJSON(*signedObj)

	// Patch URL in Core section
	c.Core.URL = strings.ReplaceAll(c.Core.URL, from, to)

	log.Debug("New Core URL: " + c.Core.URL)

	// and plugins download URLs
	for pluginName, pluginInfo := range c.Plugins {
		pluginInfo.URL = strings.ReplaceAll(c.Plugins[pluginName].URL, from, to)

		c.Plugins[pluginName] = pluginInfo

		log.Debugf("New Plugin %s data: %s", pluginName, pluginInfo.URL)
	}

	return &c, nil
}

func (p *PatchedJSONProvider) signContent(c *json_schema.InsecureUpdateJSON) (*json_schema.UpdateJSON, error) {

}

func (p *PatchedJSONProvider) GetContent() (*json_schema.UpdateJSON, error) {
	return nil, nil
}

func (p *PatchedJSONProvider) GetMetadata() (*JSONMetadataT, error) {
	return nil, nil
}

func (p *PatchedJSONProvider) IsContentUpdated() (bool, error) {
	return true, nil
}
