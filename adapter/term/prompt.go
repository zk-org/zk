package term

import (
	survey "github.com/AlecAivazis/survey/v2"
)

// Confirm is a shortcut to prompt a yes/no question to the user.
func (t *Terminal) Confirm(msg string) bool {
	confirmed := false
	prompt := &survey.Confirm{
		Message: msg,
		Default: true,
	}
	survey.AskOne(prompt, &confirmed)
	return confirmed
}
