package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"

	"github.com/mitchellh/colorstring"
)

var errInterrupt = errors.New("interrupted")

// AskNumber asks users to choose a number from 1 to max.
func (cli CLI) AskNumber(max int, defaultNum int) (int, error) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	defer signal.Stop(sigCh)

	result, errCh := cli.askNumber(max, defaultNum)

	select {
	case <-sigCh:
		fmt.Fprintln(cli.outStream)
		return -1, errInterrupt
	case num := <-result:
		return num, nil
	case err := <-errCh:
		return -1, err
	}
}

func (cli CLI) askNumber(max int, defaultNum int) (<-chan int, <-chan error) {
	result := make(chan int, 1)
	errCh := make(chan error, 1)
	go func() {
		for {
			n, err := cli.askNumber1(max, defaultNum)
			if err == nil {
				result <- n
				break
			}
			if !err.isInvalidInput() {
				errCh <- err
				break
			}
			cli.errorf("  is an invalid choice: %s.\n\n", err)
		}
	}()
	return result, errCh
}

type askError struct {
	kind askErrorKind
	err  error
}

type askErrorKind int

const (
	scanError askErrorKind = iota + 1
	invalidInput
)

func (a *askError) Error() string {
	switch a.kind {
	case scanError:
		return fmt.Sprintf("scanning from stdin: %s", a.err)
	case invalidInput:
		return a.err.Error()
	}
	panic("unreachable")
}

func (a *askError) isInvalidInput() bool {
	return a.kind == invalidInput
}

func (cli CLI) askNumber1(max int, defaultNum int) (int, *askError) {
	cli.errorf("Your choice? [default: %d] ", defaultNum)
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintln(cli.outStream)
		return 0, &askError{kind: scanError, err: err}
	}
	line = strings.TrimSpace(line)

	Debugf("Input: %q", line)

	// Use the default value
	if line == "" {
		return defaultNum, nil
	}

	// Convert string to int
	n, err := strconv.Atoi(line)
	if err != nil {
		return 0, &askError{kind: invalidInput, err: errors.New("choose by number")}
	}

	// Check the input is in range
	if n < 1 || max < n {
		return 0, &askError{kind: invalidInput, err: fmt.Errorf("choose from 1 to %d", max)}
	}

	return n, nil
}

// AskString asks users to input some string
func (cli CLI) AskString(query string, defaultStr string) (string, error) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	defer signal.Stop(sigCh)

	result, errCh := cli.askString(query, defaultStr)

	select {
	case <-sigCh:
		fmt.Fprintln(cli.outStream)
		return "", errInterrupt
	case str := <-result:
		return str, nil
	case err := <-errCh:
		return "", err
	}
}

func (cli CLI) askString(query string, defaultStr string) (<-chan string, <-chan error) {
	result := make(chan string, 1)
	errCh := make(chan error, 1)
	go func() {
		cli.errorf("%s [default: %s] ", query, defaultStr)

		reader := bufio.NewReader(os.Stdin)
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(cli.outStream)
			errCh <- &askError{kind: scanError, err: err}
			return
		}
		line = strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
		Debugf("Input: %q", line)

		// Use Default value
		if line == "" {
			line = defaultStr
		}

		result <- line
	}()
	return result, errCh
}

func (cli *CLI) ReplacePlaceholder(body string, keys []string, query, defaultReplace, optionValue string) (string, error) {
	// Replace name if needed
	folders := findPlaceholders(body, keys)

	if len(folders) > 0 {
		var ans string
		if optionValue != DefaultValue {
			ans = optionValue
		} else {
			// Ask or Confirm default value from user
			s, err := cli.AskString(query, defaultReplace)
			if err != nil {
				return "", err
			}
			ans = s
		}

		if ans != DoNothing {
			for _, f := range folders {
				cli.errorf("----> Replace placeholder %q to %q in LICENSE body\n", f, ans)
				body = strings.Replace(body, f, ans, -1)
			}
		}
	}

	return body, nil
}

// Choose shows LICENSE description from http://choosealicense.com/.
// And ask users to choose LICENSE. It returns key to fetch LICENSE file.
// If something is wrong, returns error.
func (cli *CLI) Choose() (string, error) {
	colorstring.Fprintf(cli.errStream, chooseText)

	num, err := cli.AskNumber(4, 1)
	if err != nil {
		return "", err
	}

	// If the user selects 3, should ask whether GPL V2 or V3
	if num == 3 {
		var buf bytes.Buffer
		buf.WriteString("\n")
		buf.WriteString("Which version do you want?\n")
		buf.WriteString("  1) V2\n")
		buf.WriteString("  2) V3\n")
		cli.errorf(buf.String())

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
