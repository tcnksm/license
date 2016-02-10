package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"

	"github.com/mitchellh/colorstring"
)

// AskNumber asks user to choose number from 1 to max
func (cli CLI) AskNumber(max int, defaultNum int) (int, error) {

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	defer signal.Stop(sigCh)

	result := make(chan int, 1)
	go func() {
		for {

			fmt.Fprintf(cli.errStream, "Your choice? [default: %d] ", defaultNum)
			reader := bufio.NewReader(os.Stdin)
			line, err := reader.ReadString('\n')
			if err != nil {
				Debugf("Failed to scan stdin: %s", err.Error())
			}

			Debugf("Input: %q", line)

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

// AskString asks user to input some string
func (cli CLI) AskString(query string, defaultStr string) (string, error) {

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	defer signal.Stop(sigCh)

	result := make(chan string, 1)
	go func() {
		fmt.Fprintf(cli.errStream, "%s [default: %s] ", query, defaultStr)

		reader := bufio.NewReader(os.Stdin)
		b, _, err := reader.ReadLine()
		if err != nil {
			Debugf("Failed to scan stdin: %s", err.Error())
		}
		line := string(b)
		Debugf("Input: %q", line)

		// Use Default value
		if line == "" {
			result <- defaultStr
		}

		result <- line
	}()

	select {
	case <-sigCh:
		return "", fmt.Errorf("interrupted")
	case str := <-result:
		return str, nil
	}
}

func (cli *CLI) ReplacePlaceholder(body string, keys []string, query, defaultReplace, optionValue string) string {
	// Repalce name if needed
	folders := findPlaceholders(body, keys)

	if len(folders) > 0 {
		var ans string
		if optionValue != DefaultValue {
			ans = optionValue
		} else {
			// Ask or Confirm default value from user
			ans, _ = cli.AskString(query, defaultReplace)
		}

		if ans != DoNothing {
			for _, f := range folders {
				fmt.Fprintf(cli.errStream, "----> Replace placeholder %q to %q in LICENSE body\n", f, ans)
				body = strings.Replace(body, f, ans, -1)
			}
		}
	}

	return body
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

	// If user selects 3, should ask user GPL V2 or V3
	if num == 3 {
		var buf bytes.Buffer
		buf.WriteString("\n")
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
