package main

import (
	"encoding/gob"
	"os"
)

const cacheFile = "cache.gob"

func loadCache() (map[int64]bool, error) {
	cache := map[int64]bool{}
	if _, err := os.Stat(cacheFile); os.IsNotExist(err) {
		return cache, nil
	}

	file, err := os.Open(cacheFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if err := gob.NewDecoder(file).Decode(&cache); err != nil {
		return nil, err
	}
	return cache, nil
}

func saveCache(cache map[int64]bool) error {
	file, err := os.Create(cacheFile)
	if err != nil {
		return err
	}
	defer file.Close()

	return gob.NewEncoder(file).Encode(cache)
}
