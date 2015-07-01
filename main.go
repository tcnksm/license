package main

import (
	"fmt"
	"os"
)

const (
	EnvDebug = "LI_DEBUG"
)

func main() {
	cli := &CLI{outStream: os.Stdout, errStream: os.Stderr}
	os.Exit(cli.Run(os.Args))
}

func Debugf(format string, args ...interface{}) {
	if os.Getenv(EnvDebug) != "" {
		fmt.Fprintf(os.Stdout, "[DEBUG] "+format+"\n", args...)
	}
}
