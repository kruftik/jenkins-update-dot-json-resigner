package main

import (
	"sync"
	"time"
)

type cachedEntryT struct {
	mu sync.RWMutex

	data []byte
	eol time.Time
}

func (c *cachedEntryT) IsExpired() bool {
	return time.Now().After(c.eol)
}

func (c *cachedEntryT) IsValid() bool {
	return !c.IsExpired()
}



