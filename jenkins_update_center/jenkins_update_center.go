package jenkins_update_center

import (
	"bufio"
	//"bytes"
	"fmt"
	json "github.com/gibson042/canonicaljson-go"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"jenkins-resigner-service/jenkins_update_center/json_schema"
	"net/http"
	"os"
	"strings"
	//"sync"
)

var (
	log *zap.SugaredLogger

	origJSON SyncedByteBuffer
	//PatchedJSONCache cachedEntryT
)

func Init (){
	log = zap.S()
}

func ParseUpdateJSONLocation(jsonURL, jsonPath string) error {
	if jsonURL != "" && jsonPath != "" {
		return fmt.Errorf("update.json URL and path cannot be used simultaneously")
	} else if jsonPath != "" {
		log.Info("update.json location: ", jsonPath)
		//updateJSON, err = NewUpdateJSONFromFile(jsonPath)

		f, err := os.Open(jsonPath)
		if err != nil {
			return err
		}
		defer func(){
			err = f.Close()
			log.Info("Error closing file: ", err)
		}()

		JenkinsUCJSON.Get = func() (r io.Reader, err error) {
			err = downloadUpdateJSON(jsonURL)
		}
	} else if jsonURL != "" {
		log.Info("update.json location: ", jsonURL)
		//updateJSON, err = NewUpdateJSONFromURL(jsonURL)

		resp, err := http.Head(jsonURL)
		if err != nil {
			return err
		}
		defer func(){
			err = resp.Body.Close()
			log.Info("Error closing http body: ", err)
		}()

		JenkinsUCJSON.Get = func() (r io.Reader, err error) {
			origJSON.mu.RLock()
			defer func() {
				origJSON.mu.RUnlock()
			}()
			err = downloadUpdateJSON(jsonURL)
			return bufio.NewReader(&origJSON.data), nil
		}
	} else {
		//log.Debug("Using default update.json URL: ", Opts.UpdateJSONURL)
		//updateJSON, err = NewUpdateJSONFromURL(Opts.OriginDownloadURI + "update.json")
		return fmt.Errorf("either URL or path of update.json MUST be specified")
	}

	return nil
}

func downloadUpdateJSON(downloadURL string) error {
	log.Infof("Downloading %s...", downloadURL)

	origJSON.mu.Lock()
	defer func() {
		origJSON.mu.Unlock()
	}()

	resp, err := http.Get(downloadURL)
	if err != nil {
		log.Errorf("Cannot GET %s: %s", downloadURL, err)
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	origJSON.Reset()
	//n, err := io.Copy(bufio.NewWriter(&origJSON.data), )
	n, err := origJSON.data.ReadFrom(resp.Body)
	if err != nil {
		return fmt.Errorf("cannot save update.json content to buffer: %s", err)
	}
	log.Debugf("Successfully written %d bytes to buffer", n)

	return nil
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

	UpdateJSON := &json_schema.UpdateJSON{}
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






//func (uj *UpdateJSONT) RefreshPatchedJSONCache() ([]byte, error) {
//	PatchedJSONCache.mu.Lock()
//	defer func(){
//		PatchedJSONCache.mu.Unlock()
//	}()
//
//	return nil, nil
//}
//
//func (uj *UpdateJSONT) GetPatchedJSONP() ([]byte, error) {
//	PatchedJSONCache.mu.RLock()
//
//	if PatchedJSONCache.IsValid(){
//		defer func(){
//			PatchedJSONCache.mu.RUnlock()
//		}()
//		return PatchedJSONCache.data, nil
//	} else {
//		PatchedJSONCache.mu.RUnlock()
//	}
//
//	return nil, nil
//}

