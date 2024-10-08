package input

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// ShortOrEmpty reads a single line from stdin, or nothing if enter is pressed without any text being
// entered. If a pattern is supplied, the input is matched against that pattern and will be re-requested
// if the input doesn't match.
func ShortOrEmpty(prompt string, matches *regexp.Regexp) (*string, error) {
	reader := bufio.NewReader(os.Stdin)

	for {
		// Display the prompt to the user
		fmt.Print(prompt)

		// Read input from stdin
		input, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("failed to read input: %s", err)
		}

		// Trim any surrounding whitespace/newlines
		input = strings.TrimSpace(input)

		// If the input is empty, return nil (user pressed Enter without input)
		if input == "" {
			return nil, nil
		}

		// If a pattern is provided, validate the input
		if matches != nil {
			if !matches.MatchString(input) {
				fmt.Println("Input does not match the expected pattern, please try again.")
				continue // Re-prompt if the input does not match
			}
		}

		// Return the input as a pointer
		return &input, nil
	}
}
