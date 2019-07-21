package main

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	json "github.com/gibson042/canonicaljson-go"
	"io"
	"io/ioutil"
	"jenkins-update-dot-json-resigner/update_json_schema"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/tidwall/gjson"
)

const (
	wrappedJSONPrefix  = "updateCenter.post("
	wrappedJSONPostfix = ");"
)

type jsonSymbolReplacementRuleT struct {
	from []byte
	to   []byte
}

type signingInfoT struct {
	roots *x509.CertPool
	cert  *x509.Certificate
	priv  *rsa.PrivateKey
	set   bool
}

var (
	jsonSymbolReplacementsMap = []jsonSymbolReplacementRuleT{
		{[]byte("\\u0026"), []byte("&")},
		{[]byte("\\u003c/"), []byte("<\\/")},
		{[]byte("\\u003c"), []byte("<")},
		{[]byte("\\u003e"), []byte(">")},
	}
)

type UpdateJSONT struct {
	m sync.RWMutex

	json      *update_json_schema.UpdateJSON
	Signature *update_json_schema.Signature
	//data map[string]interface{}

	signingInfo signingInfoT

	isPatched bool
}

func parseUpdateJSONLocation() (updateJSON *UpdateJSONT, err error) {
	if Opts.UpdateJSONURL != "" && Opts.UpdateJSONPath != "" {
		return nil, fmt.Errorf("update.json URL and path cannot be used simultaneously")
	} else if Opts.UpdateJSONPath != "" {
		log.Debug("update.json location: ", Opts.UpdateJSONPath)
		updateJSON, err = NewUpdateJSONFromFile(Opts.UpdateJSONPath)
	} else if Opts.UpdateJSONURL != "" {
		log.Debug("Upstream update.json URL: ", Opts.UpdateJSONURL)
		updateJSON, err = NewUpdateJSONFromURL(Opts.UpdateJSONURL)
	} else {
		//log.Debug("Using default update.json URL: ", Opts.UpdateJSONURL)
		//updateJSON, err = NewUpdateJSONFromURL(Opts.OriginDownloadURI + "update.json")
		return nil, fmt.Errorf("either URL or path of update.json MUST be specified")
	}

	return updateJSON, nil
}

func downloadUpdateJSON(downloadURL string) (*os.File, error) {
	updateJSON, err := ioutil.TempFile(os.TempDir(), "update.json")
	if err != nil {
		log.Error("Cannot create temp file for update.json: ", err)
		return nil, err
	}
	//defer func () {
	//	_ = f.Close()
	//}()

	log.Debug("Created temp file: ", updateJSON.Name())

	log.Infof("Downloading %s...", downloadURL)
	resp, err := http.Get(downloadURL)
	if err != nil {
		log.Errorf("Cannot GET %s: %s", downloadURL, err)
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	n, err := io.Copy(updateJSON, resp.Body)
	if err != nil {
		log.Errorf("Cannot save update.json content to %s", updateJSON.Name(), err)
		return nil, err
	}
	log.Debugf("Successfully written %d bytes to %s", n, updateJSON.Name())

	return updateJSON, nil
}

func extractJSONDocument(s string) (string, error) {
	idxFrom := strings.Index(s, `{`)
	idxTo := strings.LastIndex(s, `}`)

	if idxFrom == -1 || idxTo == -1 {
		return "", fmt.Errorf("cannot find a valid JSON document in the provided string")
	}

	return s[idxFrom : idxTo+1], nil

	//sLen := len(s)

	//prefixLen := len(wrappedJSONPrefix)
	//postfixLen := len(wrappedJSONPostfix)
	//if s[:prefixLen] != wrappedJSONPrefix {
	//	return "", fmt.Errorf("given JSON-wrapped string does not begin with '%s' prefix", wrappedJSONPrefix)
	//}
	//
	//if s[sLen-postfixLen:] != wrappedJSONPostfix {
	//	return "", fmt.Errorf("given JSON-wrapped string does not end with '%s' postfix", wrappedJSONPostfix)
	//}

	//return s[prefixLen : sLen-postfixLen], nil
}

func NewUpdateJSONFromURL(downloadURL string) (*UpdateJSONT, error) {
	f, err := downloadUpdateJSON(downloadURL)
	if err != nil {
		return nil, err
	}
	defer func() {
		tempFileName := f.Name()
		_ = f.Close()

		if err := os.Remove(tempFileName); err != nil {
			log.Errorf("Cannot delete obsolete temp file %s: %s", tempFileName, err)
		}
		log.Infof("Temp file %s successfully deleted", tempFileName)
	}()

	return NewUpdateJSONFromFile(f.Name())
}

func NewUpdateJSONFromFile(path string) (*UpdateJSONT, error) {
	sbytes, err := ioutil.ReadFile(path)
	if err != nil {
		log.Error("Cannot read update.json content: ", err)
		return nil, err
	}

	jsonStr, err := extractJSONDocument(string(sbytes))
	if err != nil {
		log.Error("Cannot strip json wrapping trailers: ", err)
		return nil, err
	}

	if !gjson.Valid(jsonStr) {
		log.Error("update.json is not valid")
		return nil, fmt.Errorf("update.json is not valid")
	}

	UpdateJSON := &update_json_schema.UpdateJSON{}
	err = json.Unmarshal([]byte(jsonStr), UpdateJSON)
	if err != nil {
		log.Error("Cannot unmarshal update.json: ", err)
		return nil, err
	}

	updateJSON := &UpdateJSONT{
		json:      UpdateJSON,
		Signature: &UpdateJSON.Signature,
	}

	return updateJSON, nil
}

func (uj *UpdateJSONT) SetSigningData(roots *x509.CertPool, cert *x509.Certificate, priv *rsa.PrivateKey) error {
	uj.m.Lock()
	defer func() {
		uj.m.Unlock()
	}()

	uj.signingInfo = signingInfoT{
		roots,
		cert,
		priv,
		true,
	}

	return nil
}

func (uj *UpdateJSONT) GetCertificates() ([]x509.Certificate, error) {
	var (
		sign = uj.Signature

		err      error
		crtBytes []byte
	)

	uj.m.RLock()
	defer func() {
		uj.m.RUnlock()
	}()

	certs := make([]x509.Certificate, len(sign.Certificates))

	for idx, crtBase64 := range sign.Certificates {
		crtBytes, err = base64.StdEncoding.DecodeString(crtBase64)
		if err != nil {
			return nil, fmt.Errorf("cannot decode '%s' as base64: %s", crtBase64, err)
		}

		crt, err := x509.ParseCertificate(crtBytes)
		if err != nil {
			return nil, fmt.Errorf("cannot parse '%s' as x509 cert: %s", crtBase64, err)
		}

		log.Debugf("Cert valid before %s and %s for %s", crt.NotBefore, crt.NotAfter, crt.Subject)

		certs[idx] = *crt
	}

	return certs, nil
}

//func (json *UpdateJSONT) LoadCertificates() error {
//	var (
//		err      error
//		crtBytes []byte
//
//		sign = json.Signature
//
//		roots = x509.NewCertPool()
//	)
//
//	return nil
//}

func replaceSymbolsByTrickyMap(data []byte) []byte {
	for _, r := range jsonSymbolReplacementsMap {
		data = bytes.ReplaceAll(data, r.from, r.to)
	}

	return data
}

func (uj *UpdateJSONT) GetUnsignedJSON() ([]byte, error) {
	var insecureUpdateJSON update_json_schema.InsecureUpdateJSON

	uj.m.RLock()
	defer func() {
		uj.m.RUnlock()
	}()

	insecureUpdateJSON = update_json_schema.InsecureUpdateJSON(*updateJSON.json)

	data, err := json.Marshal(insecureUpdateJSON)
	if err != nil {
		return nil, err
	}

	return replaceSymbolsByTrickyMap(data), nil
}

func (uj *UpdateJSONT) PatchUpdateCenterURLs() error {
	uj.m.Lock()
	defer func() {
		uj.m.Unlock()
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
	}

	return nil
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

func (uj *UpdateJSONT) VerifySignature() (bool, error) {
	uj.m.RLock()
	defer func() {
		uj.m.RUnlock()
	}()

	rawData, err := uj.GetUnsignedJSON()
	if err != nil {
		return false, fmt.Errorf("cannot get unsigned JSON data: %s", err)
	}

	certs, err := uj.GetCertificates()
	if err != nil {
		return false, err
	}

	return VerifySignature(
		rawData,
		certs,
		uj.Signature.CorrectDigest,
		uj.Signature.CorrectSignature,
		uj.Signature.CorrectDigest512,
		uj.Signature.CorrectSignature512,
	), nil
}

func (uj *UpdateJSONT) SignPatchedJSON() error {
	if !uj.isPatched {
		return fmt.Errorf("JSON is not patched yes")
	}

	if !uj.signingInfo.set {
		return fmt.Errorf("signing info is not configured yes")
	}

	rawData, err := uj.GetUnsignedJSON()
	if err != nil {
		return fmt.Errorf("cannot get unsigned JSON data: %s", err)
	}

	digest1str, signature1str, digest512str, signature512str, err := SignJSON(rawData, uj.signingInfo.priv)

	var resignedJSON update_json_schema.UpdateJSON
	resignedJSON = update_json_schema.UpdateJSON(rawData)
}
