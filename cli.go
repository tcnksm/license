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
	"github.com/mitchellh/colorstring"
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

	var output string

	// Define option flag parse
	flags := flag.NewFlagSet(Name, flag.ContinueOnError)
	flags.SetOutput(cli.errStream)
	flags.Usage = func() {
		fmt.Fprintf(cli.errStream, helpText)
	}

	flags.StringVar(&output, "output", DefaultOutput, "")

	flList := flags.Bool("list", false, "")
	flChoose := flags.Bool("choose", false, "")

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
		os.Setenv(EnvDebug, "1")
		Debugf("Run as DEBUG mode")
	}

	// Show list of LICENSE and quit
	if *flList {
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
	if _, err := os.Stat(output); !os.IsNotExist(err) {
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

		// Create default client
		client := github.NewClient(nil)

		// Fetch list of LICENSE from Github API
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
		buf.WriteString("Which of the following do you want to use?\n")
		for i, l := range list {
			fmt.Fprintf(&buf, "  %2d) %s\n", i+1, *l.Name)
		}
		fmt.Fprintf(cli.errStream, buf.String())

		num, err := cli.AskNumber(len(list), 1)
		if err != nil {
			fmt.Fprintf(cli.errStream, "Failed to scan user input: %s\n", err.Error())
			return ExitCodeError
		}

		key = *(list[num-1]).Key
	}

	Debugf("Get license by key: %s", key)

	// Create default client
	client := github.NewClient(nil)

	// Fetch a LICENSE from Github API
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

	fmt.Fprintf(cli.outStream, "Successfully generated %q LICENSE\n", licenseName)

	return ExitCodeOK
}

// AskNumber asks user to choose number from 1 to max
func (cli CLI) AskNumber(max int, defaultNum int) (int, error) {

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	defer signal.Stop(sigCh)

	result := make(chan int, 1)
	go func() {
		for {

			fmt.Fprintf(cli.errStream, "Your choice? [default: %d] ", defaultNum)

			var line string
			if _, err := fmt.Fscanln(os.Stdin, &line); err != nil {
				Debugf("Failed to scan stdin: %s", err.Error())
			}

			Debugf("Input: %s", line)

			// Use Default value
			if line == "" {
				result <- defaultNum
				break
			}

			// Convert string to int
			n, err := strconv.Atoi(line)
			if err != nil {
				fmt.Fprintf(cli.errStream, "  is not a valid choice. Choose by number.\n\n")
				continue
			}

			// Check input is in range
			if n < 1 || max < n {
				fmt.Fprintf(cli.errStream, "  is not a valid choice. Choose from 1 to %d\n\n", max)
				continue
			}

			result <- n
			break
		}
	}()

	select {
	case <-sigCh:
		return -1, fmt.Errorf("interrupted")
	case num := <-result:
		return num, nil
	}
}

// Choose shows shows LICENSE description from http://choosealicense.com/
// And ask user to choose LICENSE. It returns key to fetch LICENSE file.
// If something is wrong, return error.
func (cli *CLI) Choose() (string, error) {
	colorstring.Fprintf(cli.errStream, chooseText)

	num, err := cli.AskNumber(4, 1)
	if err != nil {
		return "", err
	}
	fmt.Fprintf(cli.errStream, "\n")

	// If user selects 3, should ask user GPL V2 or V3
	if num == 3 {
		var buf bytes.Buffer
		buf.WriteString("Which version do you want?\n")
		buf.WriteString("  1) V2\n")
		buf.WriteString("  2) V3\n")
		fmt.Fprintf(cli.errStream, buf.String())

		num, err = cli.AskNumber(2, 1)
		if err != nil {
			return "", err
		}
		num += 4
	}

	var key string
	switch num {
	case 1:
		key = "mit"
	case 2:
		key = "apache-2.0"
	case 4:
		key = ""
	case 5:
		key = "gpl-2.0"
	case 6:
		key = "gpl-3.0"
	default:
		// Should not reach here
		panic("Invalid number")
	}

	return key, nil
}

var chooseText = `Choose LICENSE like http://choosealicense.com/

  [blue]Choosing an OSS license doesn't need to be scary[reset]

Which of the following best describes your situation?

  1) I want it simple and permissive.

    The [red][bold]MIT License[reset] is a permissive license that is short and to the
    point. It lets people do anything they want with your code as long as they
    provide attribution back to you and don't hold you liable.
    e.g., jQuery, Rails

  2) I'm concerned about patents.

    The [red][bold]Apache License[reset] (apache-2.0) is a permissive license similar to the MIT License,
    but also provides an express grant of patent rights from contributors to users.
    e.g., Apache, SVN, NuGet

  3) I care about sharing improvements.

    The [red][bold]GPL V2[reset] (gpl-2.0) or [red][bold]GPL V3[reset] (gpl-3.0) is a copyleft license that requires
    anyone who distributes your code or a derivative work to make the source available under
    the same terms. V3 is similar to V2, but further restricts use in hardware that forbids
    software alterations.
    e.g., Linux, Git, WordPress

  4) I want more choices.

`

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

`
