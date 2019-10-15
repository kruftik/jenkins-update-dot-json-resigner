package jenkins_update_center

import (
	//"bytes"
	//"encoding/json"
	//"bytes"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"sync"

	//"jenkins-resigner-service/jenkins_update_center/json_schema"
	"time"

	//"github.com/tidwall/gjson"
	"go.uber.org/zap"
	//"sync"
)

type JenkinsLocationOpts struct {
	Src            string
	Timeout        time.Duration
	IsRemoteSource bool
}

type JenkinsPatchOpts struct {
	From string
	To   string
}

type patchedJSONP struct {
	mu sync.RWMutex
	f  string
}

type JenkinsUCOpts struct {
	Src      *JenkinsLocationOpts
	CacheTtl time.Duration

	PatchOpts JenkinsPatchOpts

	SigningInfo *SigningInfoT
}

type JenkinsUCJSONT struct {
	opts JenkinsUCOpts

	c *PatchedJSONProvider

	/*caches map[string]*cachedEntryT*/
	patchedJSONP patchedJSONP
}

var (
	log *zap.SugaredLogger
)

func Init() {
	log = zap.S()
}

func ValidateUpdateJSONLocation(jsonURL, jsonPath string) (opts *JenkinsLocationOpts, err error) {
	if jsonURL != "" && jsonPath != "" {
		return nil, fmt.Errorf("update.json URL and path cannot be used simultaneously")
	} else if jsonPath != "" {
		log.Info("update.json location: ", jsonPath)

		err = ValidateLocalFileJSONProviderSource(jsonPath)
		if err != nil {
			return nil, err
		}

		return &JenkinsLocationOpts{
			IsRemoteSource: false,
			Src:            jsonPath,
		}, nil
	} else if jsonURL != "" {
		log.Info("update.json location: ", jsonURL)

		err = ValidateURLJSONProviderSource(jsonURL)
		if err != nil {
			return nil, err
		}

		return &JenkinsLocationOpts{
			IsRemoteSource: true,
			Src:            jsonURL,
		}, nil
	}

	return nil, fmt.Errorf("neither URL no path of update.json not specified")
}

func NewJenkinsUC(opts JenkinsUCOpts) (*JenkinsUCJSONT, error) {
	var (
		err error
		juc = &JenkinsUCJSONT{
			opts: opts,
			//caches: make(map[string]*cachedEntryT),
		}
		origContentProvider JSONProvider
	)

	if opts.Src.IsRemoteSource {
		Timeout = opts.Src.Timeout
		origContentProvider, err = NewURLJSONProvider(opts.Src.Src)
		if err != nil {
			return nil, errors.Wrap(err, "cannot initialize URLJSONProvider")
		}
	} else {
		origContentProvider, err = NewLocalFileJSONProvider(opts.Src.Src)
		if err != nil {
			return nil, errors.Wrap(err, "cannot initialize LocalFileJSONProvider")
		}
	}

	tf, err := ioutil.TempFile("", "update-center.patched.json")
	if err != nil {
		return nil, errors.Wrap(err, "cannot create update-center.patched.json temp file")
	}

	log.Info("Created update-center.json temp file: ", tf.Name())

	juc.patchedJSONP = patchedJSONP{
		f: tf.Name(),
	}

	juc.patchedJSONP.mu.Lock()
	defer func() {
		if err := tf.Close(); err != nil {
			log.Error(errors.Wrap(err, "cannot close update-center.patched.json temp file"))
		}

		juc.patchedJSONP.mu.Unlock()
	}()

	juc.c, err = NewPatchedJSONProvider(origContentProvider, opts.CacheTtl, opts.PatchOpts, opts.SigningInfo)

	return juc, nil
}

func (juc *JenkinsUCJSONT) GetPatchedAndSigned() ([]byte, error) {
	isUpdated, err := juc.c.IsContentUpdated()
	if err != nil {
		return nil, err
	}

	if isUpdated {
		juc.patchedJSONP.mu.Lock()

		fd, err := os.Create(juc.patchedJSONP.f)
		if err != nil {
			return nil, err
		}

		c, meta, err := juc.c.GetContent()
		if err != nil {
			return nil, err
		}

		if _, err = juc.c.RefreshMetadata(meta); err != nil {
			log.Error(err)
			return nil, err
		}

		jsonp, err := GetJSONPString(c)
		if err != nil {
			log.Error(err)
			return nil, err
		}
		if _, err := fd.Write(jsonp); err != nil {
			return nil, err
		}

		if err := fd.Close(); err != nil {
			return nil, err
		}

		juc.patchedJSONP.mu.Unlock()
	} else {
		log.Debugf("Original content not changed, using on-disk cache")
	}

	juc.patchedJSONP.mu.RLock()
	defer func() {
		juc.patchedJSONP.mu.RUnlock()
	}()

	return ioutil.ReadFile(juc.patchedJSONP.f)
}

func (juc *JenkinsUCJSONT) Cleanup() {
	log.Info("Cleaning up temp file...")

	juc.patchedJSONP.mu.Lock()
	defer func() {
		juc.patchedJSONP.mu.Unlock()
	}()

	if err := os.Remove(juc.patchedJSONP.f); err != nil {
		log.Warn(err)
	}
}
