package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	// CacheDir is directory for caching LICENSE files
	CacheDirName = ".licns"

	// CacheDuration is duration for storing cache
	CacheDuration = 24 * time.Hour
)

func newCache(key, path string) (io.WriteCloser, error) {
	// Create cache directory if not exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(path, 0777)
		if err != nil {
			return nil, err
		}
	}

	// Delete old caches before create new
	cleanCache(key, path)

	// Create new cache file
	timeStr := strconv.Itoa(int(time.Now().Unix()))
	cacheFileName := filepath.Join(path, key+"-"+timeStr)
	Debugf("Cache filename: %s", cacheFileName)

	return os.Create(cacheFileName)
}

// getCache read cache contents from path
func getCache(key, path string) (io.Reader, error) {
	cacheFiles, err := filepath.Glob(filepath.Join(path, key+"-"+"*"))

	// Check cache file is exist or not
	if len(cacheFiles) == 0 {
		return nil, fmt.Errorf("cache file is not exist in %s", path)
	}

	// Check cache is latest or not
	cache := cacheFiles[0]
	createdUnix, err := strconv.Atoi(strings.Split(cache, "-")[1])
	if err != nil {
		return nil, fmt.Errorf("invalid cache file name: %s", cache)
	}
	createdTime := time.Unix(int64(createdUnix), 0)
	Debugf("Cache was created at %s", createdTime.String())

	if time.Now().Sub(createdTime) > CacheDuration {
		return nil, fmt.Errorf("cache file is old")
	}

	Debugf("Use Cache: %s", cache)
	return os.OpenFile(cache, os.O_RDONLY, 0777)
}

// cleanCache deletes old cache files
func cleanCache(key, path string) {
	oldCacheFiles, err := filepath.Glob(filepath.Join(path, key+"-"+"*"))
	if err != nil {
		Debugf("Failed to glob cache files: %s", err.Error())
	}

	for _, of := range oldCacheFiles {
		Debugf("Delete old cache file: %s", of)
		err := os.Remove(of)
		if err != nil {
			Debugf("Failed to delete old cache file %s: %s", of, err.Error())
		}
	}
}
