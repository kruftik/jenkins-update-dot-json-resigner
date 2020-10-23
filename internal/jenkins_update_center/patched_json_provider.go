package jenkins_update_center

import (
	"github.com/pkg/errors"
	"strings"
	"sync"
	"time"
)

type cacheEntry struct {
	mu sync.RWMutex

	ttl   time.Duration
	timer *time.Timer

	isExpired bool
}

type PatchedJSONProvider struct {
	orig     JSONProvider
	metadata *JSONMetadataT

	cacheTtl  time.Duration
	patchOpts JenkinsPatchOpts

	signingOpts *SigningInfoT

	//isPatched bool

	metadataCache *cacheEntry
}

//type patchedCacheEntry struct {
//	content  *UpdateJSON
//	metadata *JSONMetadataT
//}

func NewMetadataCache(cacheTtl time.Duration) *cacheEntry {
	log.Debugf("Setting metadata cache (expiration at %s)", time.Now().Add(cacheTtl))

	c := &cacheEntry{
		ttl: cacheTtl,
	}

	if c.timer != nil {
		c.timer.Stop()
	}

	c.timer = time.AfterFunc(cacheTtl, func() {
		log.Debug("In-memory metadata cache is expired, pruning it")
		c.isExpired = true
	})

	return c
}

func (c *cacheEntry) SetTimer() {
	c.mu.Lock()
	defer func() {
		c.mu.Unlock()
	}()

	c.isExpired = false

	c.timer.Stop()
	c.timer.Reset(c.ttl)
}

func (c *cacheEntry) IsExpired() bool {
	c.mu.RLock()
	defer func() {
		c.mu.RUnlock()
	}()

	return c.isExpired
}

func NewPatchedJSONProvider(orig JSONProvider, cacheTtl time.Duration, patchOpts JenkinsPatchOpts, signingOpts *SigningInfoT) (*PatchedJSONProvider, error) {
	p := &PatchedJSONProvider{
		orig:          orig,
		cacheTtl:      cacheTtl,
		patchOpts:     patchOpts,
		signingOpts:   signingOpts,
		metadataCache: NewMetadataCache(cacheTtl),
	}

	// Warm up the cache...
	//p.cache = NewEntryCache(nil, cacheTtl, func() (interface{}, error) {
	//	log.Info("Updating In-memory cache...")
	//
	//	//data, meta, err := p.GetFreshContent()
	//	//
	//	//return patchedCacheEntry{
	//	//	content:  data,
	//	//	metadata: meta,
	//	//}, err
	//	return p.GetFreshMetadata()
	//})

	return p, nil
}

func (p *PatchedJSONProvider) patchContent(signedOrig *UpdateJSON) (*InsecureUpdateJSON, error) {
	log.Info("Patching JSONp content...")

	c := InsecureUpdateJSON(*signedOrig)

	// Patch URL in Core section
	c.Core.URL = strings.ReplaceAll(c.Core.URL, p.patchOpts.From, p.patchOpts.To)
	log.Debug("Core URL patched")

	// and plugins download URLs
	for pluginName, pluginInfo := range c.Plugins {
		pluginInfo.URL = strings.ReplaceAll(c.Plugins[pluginName].URL, p.patchOpts.From, p.patchOpts.To)

		c.Plugins[pluginName] = pluginInfo

		//log.Debugf("New Plugin %s data: %s", pluginName, pluginInfo.URL)
	}
	log.Debug("Plugin URLs patched")

	log.Debug("Patching JSONp content [done]")
	return &c, nil
}

func (p PatchedJSONProvider) signContent(c *InsecureUpdateJSON) (*UpdateJSON, error) {
	log.Info("Signing JSONp content...")

	signature, err := p.signingOpts.SignJSONData(c)
	if err != nil {
		return nil, err
	}

	signedObj := UpdateJSON(*c)
	signedObj.Signature = *signature

	log.Debug("Signing JSONp content [done]")
	return &signedObj, nil
}

func (p *PatchedJSONProvider) GetFreshContent() (*UpdateJSON, *JSONMetadataT, error) {
	c, meta, err := p.orig.GetContent()
	if err != nil {
		return nil, nil, err
	}

	patched, err := p.patchContent(c)
	if err != nil {
		return nil, nil, err
	}

	signed, err := p.signContent(patched)
	if err != nil {
		return nil, nil, err
	}

	err = signed.VerifySignature()
	if err != nil {
		log.Error(err)
	}

	return signed, meta, nil
}

func (p PatchedJSONProvider) GetFreshMetadata() (*JSONMetadataT, error) {
	return p.orig.GetFreshMetadata()
}

func (p *PatchedJSONProvider) RefreshMetadata(meta *JSONMetadataT) (*JSONMetadataT, error) {
	var err error

	if meta == nil {
		meta, err = p.GetFreshMetadata()
		if err != nil {
			return nil, errors.Wrap(err, "cannot get origin metadata")
		}
	}

	p.metadata = meta

	p.metadataCache.SetTimer()

	return meta, err
}

func (p PatchedJSONProvider) IsContentUpdated() (bool, error) {
	if p.metadata == nil {
		return true, nil
	}

	if !p.metadataCache.IsExpired() {
		return false, nil
	}

	isUpdated, err := p.orig.IsContentUpdated()
	if err != nil {
		return false, errors.Wrap(err, "cannot check if origin content has updated")
	}

	if !isUpdated {
		p.metadataCache.SetTimer()
	}

	return isUpdated, nil

	//
	//meta, err := p.GetFreshMetadata()
	//if err != nil {
	//	return false, err
	//}
	//
	//if p.metadata.Size != meta.Size || p.metadata.LastModified != meta.LastModified {
	//	return true, nil
	//}

}

func (p PatchedJSONProvider) GetContent() (*UpdateJSON, *JSONMetadataT, error) {
	//cEntry, err := p.cache.Get()
	//if err != nil {
	//	return nil, nil, err
	//}
	//
	//c, ok := cEntry.(patchedCacheEntry)
	//if !ok {
	//	return nil, nil, fmt.Errorf("cannot desirialize cacheEntry")
	//}
	//
	//return c.content, c.metadata, nil
	return p.GetFreshContent()
}
