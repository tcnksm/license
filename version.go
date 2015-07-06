package main

import (
	"time"

	"github.com/tcnksm/go-latest"
)

const Name string = "license"
const Version string = "0.1.0"

// verCheckCh is channel which gets go-latest.Response
var verCheckCh = make(chan *latest.CheckResponse)

// CheckTimeout is default timeout of go-latest.Check execution.
var CheckTimeout = 2 * time.Second

func init() {

	go func() {
		githubTag := &latest.GithubTag{
			Owner:      "tcnksm",
			Repository: "license",
		}

		// Ignore error, because it's not important
		res, _ := latest.Check(githubTag, Version)
		verCheckCh <- res
	}()
}
