package main

import (
	"fmt"
	"net/http"

	"github.com/google/go-github/github"
)

func fetchLicenseList() ([]*github.License, error) {
	// Create default client
	client := github.NewClient(nil)

	// Fetch list of LICENSE from Github API
	list, res, err := client.Licenses.List()
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid status code from GitHub\n %s\n", res.String())
	}

	return list, nil
}

// fetchLicense fetches LICENSE file from Github API.
// if something wrong returns error.
func fetchLicense(key string) (string, error) {

	// Create default client
	client := github.NewClient(nil)

	// Fetch a LICENSE from Github API
	Debugf("Fetch license from GitHub API by key: %s", key)
	license, res, err := client.Licenses.Get(key)
	if err != nil {
		return "", err
	}

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("invalid status code from GitHub\n %s\n", res.String())
	}
	Debugf("Fetched license name: %s", *license.Name)

	return *license.Body, nil
}
