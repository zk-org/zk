package tty

import (
	"fmt"
	"strings"

	"github.com/mickael-menu/zk/core/style"
)

// PromptOpt holds metadata about a possible prompt response.
type PromptOpt struct {
	// Default value for the response.
	Label string
	// Short description explaining this response.
	Description string
	// All recognized values for this response.
	AllowedResponses []string
}

// Prompt displays a message and waits for the user to input one of the
// available options.
// Returns the selected option index.
func (t *TTY) Prompt(msg string, defaultOpt int, options []PromptOpt) int {
	responses := ""
	for i, opt := range options {
		if i == len(options)-1 {
			responses += " or "
		} else if i > 0 {
			responses += ", "
		}
		responses += opt.Label
	}

	printHelp := func() {
		fmt.Println("\nExpected responses:")
		for _, opt := range options {
			fmt.Printf("  %v\t%v\n", opt.Label, opt.Description)
		}
		fmt.Println()
	}

	for {
		fmt.Printf("%s\n%s > ", msg, responses)

		// Don't prompt when --no-input is on.
		if t.NoInput {
			fmt.Println(options[defaultOpt].AllowedResponses[0])
			return defaultOpt
		}

		var response string
		_, err := fmt.Scan(&response)
		if err != nil {
			return defaultOpt
		}
		response = strings.ToLower(response)

		for i, opt := range options {
			for _, allowedResp := range opt.AllowedResponses {
				if response == strings.ToLower(allowedResp) {
					return i
				}
			}
		}

		printHelp()
	}
}

// Confirm is a shortcut to prompt a yes/no question to the user.
func (t *TTY) Confirm(msg string, yesDescription string, noDescription string) bool {
	return t.Prompt(msg, 1, []PromptOpt{
		{
			Label:            t.MustStyle("y", style.RuleEmphasis) + "es",
			Description:      yesDescription,
			AllowedResponses: []string{"yes", "y", "ok"},
		},
		{
			Label:            t.MustStyle("n", style.RuleEmphasis) + "o",
			Description:      noDescription,
			AllowedResponses: []string{"no", "n"},
		},
	}) == 0
}
