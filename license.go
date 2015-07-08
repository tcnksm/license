package main

import (
	"regexp"
	"strings"
)

var yearKeys = []string{
	"{year}",
	"<year>",
	"{yyyy}",
	"[year]",
}

var nameKeys = []string{
	"{name of author}",
	"<name of author>",
	"[name of author]",
	"{fullname}",
	"[fullname]",
	"{name of copyright owner}",
}

var emailKeys = []string{
	"[email]",
}

var miscKeys = []string{
	"<one line to give the program's name and a brief idea of what it does.>",
	"{one line to give the program's name and a brief idea of what it does.}",
	"{signature of Ty Coon}",
	"[project]",
	"{description}",
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
