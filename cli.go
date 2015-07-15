package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/github"
	"github.com/mitchellh/go-homedir"
	"github.com/olekukonko/tablewriter"
	"github.com/tcnksm/go-gitconfig"
)

// Exit codes are int values that represent an exit code for a particular error.
const (
	ExitCodeOK    int = 0
	ExitCodeError int = 1 + iota
	ExitCodeErrorCache
)

const (
	// DefaultOutput is default output file name
	DefaultOutput = "LICENSE"

	// Default Values. If it's not changed, licence will ask them
	DefaultValue = "*"

	// Default value for user input
	DoNothing = "(no replacement)"
)

// CLI is the command line object
type CLI struct {
	// outStream and errStream are the stdout and stderr
	// to write message from the CLI.
	outStream, errStream io.Writer
}

// Run invokes the CLI with the given arguments.
func (cli *CLI) Run(args []string) int {

	var (
		output string

		optionYear    string
		optionAuthor  string
		optionEmail   string
		optionProject string

		force   bool
		noCache bool
		raw     bool
	)

	// Define option flag parse
	flags := flag.NewFlagSet(Name, flag.ContinueOnError)
	flags.SetOutput(cli.errStream)
	flags.Usage = func() {
		fmt.Fprintf(cli.errStream, helpText)
	}

	flags.StringVar(&output, "output", DefaultOutput, "")
	flags.BoolVar(&noCache, "no-cache", false, "")
	flags.BoolVar(&force, "force", false, "")
	flags.BoolVar(&raw, "raw", false, "")

	// Replacement values
	flags.StringVar(&optionYear, "year", DefaultValue, "")
	flags.StringVar(&optionAuthor, "author", DefaultValue, "")
	flags.StringVar(&optionEmail, "email", DefaultValue, "")
	flags.StringVar(&optionProject, "project", DefaultValue, "")

	flList := flags.Bool("list", false, "")
	flChoose := flags.Bool("choose", false, "")

	flDebug := flags.Bool("debug", false, "")
	flVersion := flags.Bool("version", false, "")

	// This is only for dev (and test)
	flListkeys := flags.Bool("list-keys", false, "")

	// Parse commandline flag
	if err := flags.Parse(args[1:]); err != nil {
		return ExitCodeError
	}

	// Show version
	if *flVersion {

		fmt.Fprintf(cli.errStream, "%s version %s\n", Name, Version)
		select {
		case <-time.After(CheckTimeout):
			// Do nothing
		case res := <-verCheckCh:
			if res != nil {
				msg := fmt.Sprintf("Latest version of license is %s, please update it\n", res.Current)
				fmt.Fprint(cli.errStream, msg)
			}
		}

		return ExitCodeOK
	}

	// Set Debug environmental variable
	if *flDebug {
		os.Setenv(EnvDebug, "1")
		Debugf("Run as DEBUG mode")
	}

	// Show list of LICENSE and quit
	if *flList || *flListkeys {
		Debugf("Show list of LICENSE")

		// Create default client
		client := github.NewClient(nil)

		// Fetch list from Github API
		list, res, err := client.Licenses.List()
		if err != nil {
			fmt.Fprintf(cli.errStream, "Failed to fetch LICENSE list: %s\n", err.Error())
			return ExitCodeError
		}

		if res.StatusCode != http.StatusOK {
			fmt.Fprintf(cli.errStream, "Invalid status code from GitHub\n %s\n", res.String())
			return ExitCodeError
		}

		// List LICENSE keys (name used when fetching)
		// This is only for dev(testing)
		if *flListkeys {
			Debugf("List LICENSE keys")
			for _, l := range list {
				fmt.Fprintf(cli.outStream, "%s\n", *l.Key)
			}
			return ExitCodeOK
		}

		// Write LICENSE list as a table
		outBuffer := new(bytes.Buffer)
		table := tablewriter.NewWriter(outBuffer)

		header := []string{"Key", "Name"}
		table.SetHeader(header)
		for _, l := range list {
			Debugf("%s (%s)", *l.Name, *l.Key)
			table.Append([]string{*l.Key, *l.Name})
		}
		table.Render()

		outBuffer.WriteString("See more about these LICENSE at http://choosealicense.com/licenses/\n")
		fmt.Fprintf(cli.outStream, outBuffer.String())

		return ExitCodeOK
	}

	// Check file exist or not
	if _, err := os.Stat(output); !os.IsNotExist(err) && !force {
		fmt.Fprintf(cli.errStream, "Cannot create file %q: file exists\n", output)
		return ExitCodeError
	}

	parsedArgs := flags.Args()
	if len(parsedArgs) > 1 {
		fmt.Fprintf(cli.errStream, "Invalid arguments\n")
		return ExitCodeError
	}

	var key string
	if len(parsedArgs) == 1 {
		key = parsedArgs[0]
		// Every key must be lower case
		key = strings.ToLower(key)
	}

	// Choose a LICENSE like http://choosealicense.com/
	if len(key) == 0 && *flChoose {
		Debugf("Choose a LICENSE like http://choosealicense.com/")
		var err error
		key, err = cli.Choose()
		if err != nil {
			fmt.Fprintf(cli.errStream, "Failed to choose a LICENSE: %s\n", err.Error())
			return ExitCodeError
		}
	}

	// Show all LICENSE available and ask user to select.
	if len(key) == 0 {
		Debugf("Show all LICENSE available and ask user to select")

		list, err := fetchLicenseList()
		if err != nil {
			fmt.Fprintf(cli.errStream, "Failed to show LICENSE list: %s", err.Error())
			return ExitCodeError
		}

		var buf bytes.Buffer
		buf.WriteString("Which of the following do you want to use?\n")
		for i, l := range list {
			fmt.Fprintf(&buf, "  %2d) %s\n", i+1, *l.Name)
		}
		fmt.Fprintf(cli.errStream, buf.String())

		// Use MIT as default, it may change in future
		// So should fix it
		defaultNum := 13

		num, err := cli.AskNumber(len(list), defaultNum)
		if err != nil {
			fmt.Fprintf(cli.errStream, "Failed to scan user input: %s\n", err.Error())
			return ExitCodeError
		}

		key = *(list[num-1]).Key
	}

	home, err := homedir.Dir()
	if err != nil {
		Debugf("Failed to get home directory: %s", err.Error())
		noCache = true
		home = "."
	}
	cacheDir := filepath.Join(home, CacheDirName)

	// By default noCache is false (useCache) and check cache is exist or not
	var body string
	if !noCache {
		var err error
		body, err = getCache(key, cacheDir)
		if err != nil {
			Debugf("Failed to get cache: %s", err.Error())
		}
	}

	// If cache is not exist, fetch it from GitHub
	var fetched bool = false
	if len(body) == 0 {
		var err error
		body, err = fetchLicense(key)
		if err != nil {
			fmt.Fprintf(cli.errStream, "Failed to get LICENSE file: %s\n", err.Error())
			return ExitCodeError
		}

		if !noCache {
			err := setCache(body, key, cacheDir)
			if err != nil {
				Debugf("Failed to save cache: %s", err.Error())
			}
		}

		fetched = true
	}

	// Create output path if it is not exist
	dir, _ := filepath.Split(output)
	if len(dir) != 0 {
		os.MkdirAll(dir, 0777)
	}

	licenseWriter, err := os.Create(output)
	if err != nil {
		fmt.Fprintf(cli.errStream, "Failed to create file %s: %s\n", output, err.Error())
		return ExitCodeError
	}
	defer licenseWriter.Close()
	Debugf("Output filename: %s", output)

	// Replace place holders
	if !raw {

		// Replace year if needed
		var year string
		if optionYear != DefaultValue {
			year = optionYear
		} else {
			year = strconv.Itoa(time.Now().Year())
		}

		yearFolders := findPlaceholders(body, yearKeys)
		for _, f := range yearFolders {
			fmt.Fprintf(cli.errStream, "----> Replace placeholder %q to %q in LICENSE body\n", f, year)
			body = strings.Replace(body, f, year, -1)
		}

		// Replace author name if needed
		defaultAuthor, _ := gitconfig.GithubUser()
		if len(defaultAuthor) == 0 {
			defaultAuthor = DoNothing
		}
		body = cli.ReplacePlaceholder(body, nameKeys, "Input author name", defaultAuthor, optionAuthor)

		// Replace email if needed
		defaultEmail, _ := gitconfig.Email()
		if len(defaultEmail) == 0 {
			defaultEmail = DoNothing
		}
		body = cli.ReplacePlaceholder(body, nameKeys, "Input email", defaultEmail, optionEmail)

		// Replace project name if needed
		body = cli.ReplacePlaceholder(body, projectKeys, "Input project name", DoNothing, optionProject)
	}

	// Write LICENSE body to file
	_, err = io.Copy(licenseWriter, strings.NewReader(body))
	if err != nil {
		fmt.Fprintf(cli.errStream, "Failed to write license body to %q: %s\n", output, err.Error())
		return ExitCodeError
	}

	// Output message to user
	var msg bytes.Buffer
	msg.WriteString(fmt.Sprintf("====> Successfully generated %q LICENSE", key))
	if !noCache && !fetched {
		msg.WriteString(" (Use cache)")
	}

	fmt.Fprintf(cli.errStream, msg.String()+"\n")

	return ExitCodeOK
}

var helpText = `Usage: license [option] [KEY]

  Generate LICENSE file. If you provide KEY, it will try to get LICENSE by
  it. If you don't provide it, it will ask you to choose from avairable list.
  You can check avairable LICESE list by '-list' option.

Options:

  -list               Show all avairable LICENSE list and quit.
                      It will fetch information from GitHub.

  -choose             Choose LICENSE like http://choosealicense.com/
                      It shows you which LICENSE is useful for you.

  -output=NAME        Change output file name.
                      By default, output file name is 'LICENSE'

  -force              Replace LICENSE file if exist.
                      By default, it stop generating if file is alreay
                      exist

  -no-cache           Disable using local cache.
                      By default, it uses local cache file which
                      is saved in ~/.lcns folder. 

  -raw                Generate raw LICENSE file.
                      By default, it replace year, name, or email

`
