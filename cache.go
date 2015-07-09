package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	// CacheDir is directory for caching LICENSE files
	CacheDirName = ".lcns"

	// CacheDuration is duration for storing cache
	CacheDuration = 30 * 24 * time.Hour
)

// setCache saves body as file (named key + Unix time) in provided
// path. Any errors that occur are returned.
func setCache(body, key, path string) error {
	// Create cache directory if not exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(path, 0777)
		if err != nil {
			return err
		}
	}

	// Delete old caches before create new
	cleanCache(key, path)

	// Create new cache file
	timeStr := strconv.Itoa(int(time.Now().Unix()))
	cacheFileName := filepath.Join(path, key+"-"+timeStr)
	Debugf("Cache filename: %s", cacheFileName)

	cacheWriter, err := os.Create(cacheFileName)
	if err != nil {
		return err
	}

	r := strings.NewReader(body)
	_, err = io.Copy(cacheWriter, r)
	return err
}

// getCache read cache contents from provided path.
// Any errors that occur are returned.
func getCache(key, path string) (string, error) {
	cacheFiles, err := filepath.Glob(filepath.Join(path, key+"-"+"*"))

	// Check cache file is exist or not
	if len(cacheFiles) == 0 {
		return "", fmt.Errorf("cache file is not exist in %s", path)
	}

	// Check cache is latest or not
	cache := cacheFiles[0]
	createdUnix, err := strconv.Atoi(strings.Split(cache, "-")[len(strings.Split(cache, "-"))-1])
	if err != nil {
		return "", fmt.Errorf("invalid cache file name: %s", cache)
	}
	createdTime := time.Unix(int64(createdUnix), 0)
	Debugf("Cache was created at %s", createdTime.String())

	if time.Now().Sub(createdTime) > CacheDuration {
		return "", fmt.Errorf("cache file is old")
	}

	b, err := ioutil.ReadFile(cache)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// cleanCache deletes old cache files. Any errors that occur are returned.
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
