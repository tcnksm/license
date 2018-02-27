package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/go-github/github"
)

func fetchLicenseList() ([]*github.License, error) {
	// Create default client
	client := github.NewClient(nil)

	// Fetch a list of LICENSEs from GitHub API
	list, res, err := client.Licenses.List(context.Background())
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid status code from GitHub\n %s\n", res.String())
	}

	return list, nil
}

// fetchLicense fetches a LICENSE file from GitHub API.
// If something is wrong, it returns the error.
func fetchLicense(key string) (string, error) {

	// Create default client
	client := github.NewClient(nil)

	// Fetch a LICENSE from GitHub API
	Debugf("Fetch license from GitHub API by key: %s", key)
	license, res, err := client.Licenses.Get(context.Background(), key)
	if err != nil {
		return "", err
	}

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("invalid status code from GitHub\n %s\n", res.String())
	}
	Debugf("Fetched license name: %s", *license.Name)

	return *license.Body, nil
}
