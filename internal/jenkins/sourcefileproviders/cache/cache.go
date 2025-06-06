package cache

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/kruftik/jenkins-update-dot-json-resigner/internal/jenkins/sourcefileproviders"
)

var (
	_ sourcefileproviders.Provider = (*Cache)(nil)
)

type Cache struct {
	log *zap.SugaredLogger
	p   sourcefileproviders.Provider

	dataFile string
	metadata sourcefileproviders.FileMetadata

	mu sync.RWMutex
}

func NewCacheWrapper(ctx context.Context, log *zap.SugaredLogger, p sourcefileproviders.Provider, cacheDuration time.Duration) (*Cache, error) {
	fData, err := os.CreateTemp("", "cache-wrapper-*.data")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp data file: %w", err)
	}
	_ = fData.Close()

	log.Debugf("%s temp data file created", fData.Name())

	c := &Cache{
		log:      log,
		p:        p,
		dataFile: fData.Name(),
	}

	if err := c.refreshContent(ctx); err != nil {
		return nil, err
	}

	go c.runCacheWorker(ctx, cacheDuration)

	return c, nil
}

func (c *Cache) refreshContent(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	metadata, err := c.p.GetMetadata(ctx)
	if err != nil {
		return fmt.Errorf("failed to get JSONP metadata: %w", err)
	}

	_, err = os.Stat(c.dataFile)
	if err != nil {
		c.log.Infof("data file %s does not exist, force update", c.dataFile)
	}

	if metadata == c.metadata && err == nil {
		c.log.Debugf("cached JSONP body is up-to-date, skipping update")
		return nil
	}

	metadata, signedJSON, err := c.p.GetBody(ctx)
	if err != nil {
		return fmt.Errorf("failed to get JSONP body: %w", err)
	}
	defer signedJSON.Close()

	f, err := os.CreateTemp("", "cache-wrapper-*.jsonp")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer f.Close()

	if _, err = io.Copy(f, signedJSON); err != nil {
		return fmt.Errorf("failed to write JSONP body: %w", err)
	}

	if err := os.Rename(f.Name(), c.dataFile); err != nil {
		return fmt.Errorf("failed to move data file %s to %s: %w", f.Name(), c.dataFile, err)
	}

	c.metadata = metadata

	return nil
}

func (c *Cache) runCacheWorker(ctx context.Context, cacheDuration time.Duration) {
	c.log.Infow("starting cache refresh worker")
	defer c.log.Infow("cache refresh worker stopped")

	ticker := time.NewTicker(cacheDuration)
	defer ticker.Stop()

FOR:
	for {
		select {
		case <-ctx.Done():
			break FOR
		case <-ticker.C:
			c.log.Info("refreshing cache content")

			if err := c.refreshContent(ctx); err != nil {
				c.log.Errorf("failed to refresh cache content: %v", err)
			}
		}
	}

	if err := os.Remove(c.dataFile); err != nil {
		c.log.Warnf("cannot remove temp data file: %v", err)
	}
}

func (c *Cache) GetBody(_ context.Context) (sourcefileproviders.FileMetadata, io.ReadCloser, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	f, err := os.Open(c.dataFile)
	if err != nil {
		return sourcefileproviders.FileMetadata{}, nil, fmt.Errorf("failed to open data file: %w", err)
	}

	return c.metadata, f, nil
}

func (c *Cache) GetMetadata(_ context.Context) (sourcefileproviders.FileMetadata, error) {
	return c.metadata, nil
}
