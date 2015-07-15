package main

import (
	"regexp"
	"strings"
)

var yearKeys = []string{
	"[year]",
}

var nameKeys = []string{
	"[fullname]",
}

var emailKeys = []string{
	"[email]",
}

var projectKeys = []string{
	"[project]",
}

func findPlaceholders(body string, keys []string) (folders []string) {
	for _, k := range keys {
		if strings.Contains(body, k) {
			folders = append(folders, k)
		}
	}
	return
}

var reg = regexp.MustCompile("[{}<>\\[\\]]")

func constructQuery(raw string) string {
	return strings.TrimRight(reg.ReplaceAllString(raw, ""), ".")
}
