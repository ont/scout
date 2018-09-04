package main

import "time"

func ConfiguredCache(cachePath string) Cache {
	if cachePath == "" {
		return NewCacheMap()
	} else {
		cache := NewCacheBadger(cachePath)
		go cache.CleanEvery(5 * time.Minute) // start autoclean in parallel
		return cache
	}
}
