package jenkins_update_center

import (
	//"bytes"
	"fmt"
	"jenkins-resigner-service/jenkins_update_center/json_schema"
	"net/http"
	"net/url"
	"os"
	"time"

	//"github.com/tidwall/gjson"
	"go.uber.org/zap"
	//"sync"
)

var (
	log *zap.SugaredLogger
)

func Init() {
	log = zap.S()
}

func ValidateUpdateJSONLocation(jsonURL, jsonPath string) (isRemoteJSON bool, err error) {
	if jsonURL != "" && jsonPath != "" {
		return false, fmt.Errorf("update.json URL and path cannot be used simultaneously")
	} else if jsonPath != "" {
		log.Info("update.json location: ", jsonPath)
		//updateJSON, err = NewUpdateJSONFromFile(jsonPath)

		f, err := os.Open(jsonPath)
		if err != nil {
			return false, err
		}
		defer func() {
			err = f.Close()
			log.Info("Error closing file: ", err)
		}()

		return false, nil
	} else if jsonURL != "" {
		log.Info("update.json location: ", jsonURL)
		//updateJSON, err = NewUpdateJSONFromURL(jsonURL)

		resp, err := http.Head(jsonURL)
		if err != nil {
			return false, err
		}
		defer func() {
			err = resp.Body.Close()
			log.Info("Error closing http body: ", err)
		}()

		return true, nil
	}

	return false, fmt.Errorf("neither URL no path of update.json not specified")
}

func NewJenkinsUC(jsonURL, jsonPath string, cacheTTL time.Duration) (*JenkinsUCJSONT, error) {
	isRemoteJSON, err := ValidateUpdateJSONLocation(jsonURL, jsonPath)
	if err != nil {
		return nil, err
	}

	var juc *JenkinsUCJSONT

	if isRemoteJSON {
		rURL, err := url.ParseRequestURI(jsonURL)
		if err != nil {
			return nil, err
		}
		func(src url.URL) {
			juc = &JenkinsUCJSONT{
				src: src,
				get: func() (*json_schema.UpdateJSON, error) {
					return downloadUpdateJSONFromURL(src.String())
				},
			}
		}(*rURL)
	} else {
		func(src string) {
			juc = &JenkinsUCJSONT{
				src: src,
				get: func() (*json_schema.UpdateJSON, error) {
					return readUpdateJSONFromFile(src)
				},
			}
		}(jsonPath)
	}

	juc.cacheTTL = cacheTTL

	return juc, nil
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
