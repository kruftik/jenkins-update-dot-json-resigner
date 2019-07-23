package jenkins_update_center

import (
	"fmt"
	"time"
)

type cacheUpdateFnT func() (interface{}, error)

type cachedEntryT struct {
	data      interface{}
	Ttl       time.Duration
	UpdatedAt time.Time
	Eol       time.Time

	updateFn cacheUpdateFnT
}

func NewEntryCache(data interface{}, ttl time.Duration, updateFn cacheUpdateFnT) *cachedEntryT {
	c := cachedEntryT{
		data:      data,
		Ttl:       ttl,
		UpdatedAt: time.Now(),
		Eol:       time.Now().Add(ttl),

		updateFn: updateFn,
	}

	return &c
}

func (c *cachedEntryT) IsExpired() bool {
	return time.Now().After(c.Eol)
}

func (c *cachedEntryT) IsValid() bool {
	return !c.IsExpired()
}

func (c *cachedEntryT) Update() error {
	if c.updateFn == nil {
		return fmt.Errorf("updateFn not provided")
	}

	data, err := c.updateFn()
	if err != nil {
		return err
	}

	c.Set(data)

	return nil
}

func (c *cachedEntryT) Get() (interface{}, error) {
	if c.IsExpired() {
		log.Info("cache has expired")

		if err := c.Update(); err != nil {
			return nil, err
		}
	}
	return c.data, nil
}

func (c *cachedEntryT) Set(data interface{}) {
	c.data = data
	c.UpdatedAt = time.Now()
	c.Eol = time.Now().Add(c.Ttl)
}
