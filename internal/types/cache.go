package types

import (
	"log"
	"net/http"
	"sync"
)

type CacheEntry struct {
	Body   interface{}
	header http.Header
}

type ActionCache struct {
	mutex   sync.RWMutex
	entries map[string]*CacheEntry
}

func NewActionCache() *ActionCache {
	res := &ActionCache{}
	res.entries = make(map[string]*CacheEntry)
	return res
}

func (cache *ActionCache) Entry(key string, fetcher func() interface{}) *CacheEntry {
	cache.mutex.Lock()
	resp, ok := cache.entries[key]
	if !ok {
		log.Printf("Cache miss, fetching...\n")
		out := fetcher()
		resp = &CacheEntry{Body: out}
		cache.entries[key] = resp
	} else {
		log.Printf("Cache hit\n")
	}
	cache.mutex.Unlock()
	return resp
}
