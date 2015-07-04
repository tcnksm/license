package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"

	"github.com/google/go-github/github"
	"github.com/olekukonko/tablewriter"
)

// Exit codes are int values that represent an exit code for a particular error.
const (
	ExitCodeOK    int = 0
	ExitCodeError int = 1 + iota
)

const (
	// DefaultOutput is default output file name
	DefaultOutput = "LICENSE"
)

// CLI is the command line object
type CLI struct {
	// outStream and errStream are the stdout and stderr
	// to write message from the CLI.
	outStream, errStream io.Writer
}

// Run invokes the CLI with the given arguments.
func (cli *CLI) Run(args []string) int {

	var LicenseInfo LicenseInfo
	var output string

	// Define option flag parse
	flags := flag.NewFlagSet(Name, flag.ContinueOnError)
	flags.SetOutput(cli.errStream)
	flags.Usage = func() {
		fmt.Fprintf(cli.errStream, helpText)
	}

	flags.StringVar(&LicenseInfo.Author, "author", "", "")
	flags.StringVar(&output, "output", DefaultOutput, "")

	flList := flags.Bool("list", false, "")
	// flChoose := flags.Bool("choose", false, "")

	flDebug := flags.Bool("debug", false, "")
	flVersion := flags.Bool("version", false, "")

	// Parse commandline flag
	if err := flags.Parse(args[1:]); err != nil {
		return ExitCodeError
	}

	// Show version
	if *flVersion {
		fmt.Fprintf(cli.errStream, "%s version %s\n", Name, Version)
		return ExitCodeOK
	}

	// Set Debug environmental variable
	if *flDebug {
		if err := os.Setenv(EnvDebug, "1"); err != nil {
			// Should not reach here
			panic(err)
		}
	}

	if *flList {

		client := github.NewClient(nil)

		Debugf("Show list of LICENSE")
		list, res, err := client.Licenses.List()
		if err != nil {
			fmt.Fprintf(cli.errStream, "Failed to fetch LICENSE list: %s\n", err.Error())
			return ExitCodeError
		}

		if res.StatusCode != http.StatusOK {
			fmt.Fprintf(cli.errStream, "Invalid status code from GitHub\n %s\n", res.String())
			return ExitCodeError
		}

		outBuffer := new(bytes.Buffer)
		table := tablewriter.NewWriter(outBuffer)

		header := []string{"Key", "Name"}
		table.SetHeader(header)
		for _, l := range list {
			Debugf("%s (%s)", *l.Name, *l.Key)
			table.Append([]string{*l.Key, *l.Name})
		}
		table.Render()

		fmt.Fprintf(cli.outStream, outBuffer.String())
		return ExitCodeOK
	}

	// Check file exist or not
	// TODO: -force option
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		fmt.Fprintf(cli.errStream, "Cannot create file %q: file exists\n", output)
		return ExitCodeError
	}

	parsedArgs := flags.Args()
	if len(parsedArgs) > 1 {
		fmt.Fprintf(cli.errStream, "Invalid arguments")
		return ExitCodeError
	}

	var key string
	if len(parsedArgs) == 1 {
		key = parsedArgs[0]
		// Every key must be lower case
		key = strings.ToLower(key)
	}

	// Ask user to select license
	if len(key) == 0 {
		client := github.NewClient(nil)

		Debugf("Show list of LICENSE")
		list, res, err := client.Licenses.List()
		if err != nil {
			fmt.Fprintf(cli.errStream, "Failed to fetch LICENSE list: %s\n", err.Error())
			return ExitCodeError
		}

		if res.StatusCode != http.StatusOK {
			fmt.Fprintf(cli.errStream, "Invalid status code from GitHub\n %s\n", res.String())
			return ExitCodeError
		}

		var buf bytes.Buffer
		buf.WriteString("Choose LICENSE\n")
		for i, l := range list {
			fmt.Fprintf(&buf, "  %2d) %s\n", i+1, *l.Name)
		}

		fmt.Fprintf(cli.errStream, buf.String())

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt)
		defer signal.Stop(sigCh)

		result := make(chan int, 1)
		go func() {
			for {

				defaultChoice := 1
				fmt.Fprintf(cli.errStream, "Your choice? [default: %d] ", defaultChoice)

				var line string
				if _, err := fmt.Fscanln(os.Stdin, &line); err != nil {
					Debugf("Failed to scan stdin: %s", err.Error())
				}

				Debugf("Input: %s", line)

				// Use Default value
				if line == "" {
					result <- defaultChoice
					break
				}

				n, err := strconv.Atoi(line)
				if err != nil {
					fmt.Fprintf(cli.errStream, " is not a valid choice. Choose by number.\n")
					continue
				}

				if n < 1 || len(list) < n {
					fmt.Fprintf(cli.errStream, " is not a valid choice. Choose from 1 to %d\n", len(list))
					continue
				}

				result <- n
				break
			}
		}()

		select {
		case <-sigCh:
			fmt.Fprintf(cli.errStream, "Interrupted\n")
			return ExitCodeError
		case num := <-result:
			key = *(list[num-1]).Key
			Debugf("Select: %s", key)
		}
	}

	Debugf("Get license by key: %s", key)
	client := github.NewClient(nil)
	license, res, err := client.Licenses.Get(key)
	if err != nil {
		fmt.Fprintf(cli.errStream, "Failed to fetch LICENSE: %s\n", err.Error())
		return ExitCodeError
	}

	if res.StatusCode != http.StatusOK {
		fmt.Fprintf(cli.errStream, "Invalidd status code from GitHub\n %s\n", res.String())
		return ExitCodeError
	}

	licenseName := *license.Name
	Debugf("Fetched license name: %s", licenseName)

	licenseBody := *license.Body
	Debugf("Fetched license body:\n\n%s", licenseBody)

	r := strings.NewReader(licenseBody)
	w, err := os.Create(output)
	if err != nil {
		panic(err)
	}
	defer w.Close()
	Debugf("Output filename: %s", output)

	i, err := io.Copy(w, r)
	if i < 0 || err != nil {
		fmt.Fprintf(cli.errStream, "Failed to write license body to %q: %s\n", output)
		return ExitCodeError
	}
	Debugf("Written: %d", i)

	fmt.Fprintf(cli.outStream, "Successfully generated %q LICENSE file\n", licenseName)
	return ExitCodeOK
}

var helpText = `Usage: license [option] [KEY]

  Generate LICENSE file. If you provide KEY, it will try to get LICENSE by
  it. If you don't provide it, it will ask you to choose from avairable list.
  You can check avairable LICESE list by '-list' option. 

Options:

  -list               Show all avairable LICENSE list and quit.
                      It will fetch information from GitHub. 

  -output=NAME        Change output file name.
                      By default, output file name is 'LICENSE'

`
