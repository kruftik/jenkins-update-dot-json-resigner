package jenkins_update_center

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func (juc *JenkinsUCJSONT) GetOriginalJSON() (*[]bytes.Buffer, error) {
}

func (uj *UpdateJSONT) SaveJSON(path string, savePatched bool) (err error) {
	var (
		jsonData []byte
	)

	if !savePatched {
		jsonData, err = json.Marshal(uj.json)
		if err != nil {
			return fmt.Errorf("cannot marshal struct to JSON: %s", err)
		}
		jsonData = replaceSymbolsByTrickyMap(jsonData)
	} else {
		jsonData, err = uj.GetUnsignedJSON()
		if err != nil {
			return fmt.Errorf("cannot get unsigned JSON: %s", err)
		}
	}

	err = ioutil.WriteFile(path, append(jsonData, []byte("\n")[:]...), os.ModePerm)
	if err != nil {
		return fmt.Errorf("cannot save JSON to %s: %s", path, err)
	}

	log.Debugf("Successfully saved JSON (savePatched=%t) to %s", savePatched, path)

	return nil
}

func (uj *UpdateJSONT) SaveJSONP(path string, savePatched bool) (err error) {
	var (
		jsonData []byte
	)

	if !savePatched {
		jsonData, err = json.Marshal(uj.json)
		if err != nil {
			return fmt.Errorf("cannot marshal struct to JSON: %s", err)
		}
		jsonData = replaceSymbolsByTrickyMap(jsonData)
	} else {
		jsonData, err = uj.GetUnsignedJSON()
		if err != nil {
			return fmt.Errorf("cannot get unsigned JSON: %s", err)
		}
	}

	jsonp := []byte(wrappedJSONPrefix + "\n")

	jsonp = append(jsonp, jsonData...)
	jsonp = append(jsonp, []byte("\n"+wrappedJSONPostfix)...)

	err = ioutil.WriteFile(path, jsonp, os.ModePerm)
	if err != nil {
		return fmt.Errorf("cannot save JSON to %s: %s", path, err)
	}

	log.Debugf("Successfully saved JSON (savePatched=%t) to %s", savePatched, path)

	return nil
}

func (uj *UpdateJSONT) PatchUpdateCenterURLs() error {
	uj.mu.Lock()
	defer func() {
		uj.mu.Unlock()
	}()

	if uj.isPatched {
		return fmt.Errorf("update.json data have been already patched")
	}

	uj.isPatched = true

	// Patch URL in Core section
	uj.json.Core.URL = strings.ReplaceAll(uj.json.Core.URL, Opts.OriginDownloadURI, Opts.NewDownloadURI)

	log.Debug("New Core URL: " + uj.json.Core.URL)

	// and plugins download URLs
	for pluginName, pluginInfo := range uj.json.Plugins {
		pluginInfo.URL = strings.ReplaceAll(uj.json.Plugins[pluginName].URL, Opts.OriginDownloadURI, Opts.NewDownloadURI)

		log.Debugf("New Plugin %s data: %s", pluginName, pluginInfo.URL)

		uj.json.Plugins[pluginName] = pluginInfo
	}

	return nil
}
