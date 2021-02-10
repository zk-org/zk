package term

import (
	survey "github.com/AlecAivazis/survey/v2"
)

// Confirm is a shortcut to prompt a yes/no question to the user.
func (t *Terminal) Confirm(msg string, defaultAnswer bool) (confirmed, skipped bool) {
	if !t.IsInteractive() {
		return defaultAnswer, true
	}

	confirmed = false
	prompt := &survey.Confirm{
		Message: msg,
		Default: defaultAnswer,
	}
	survey.AskOne(prompt, &confirmed)
	return confirmed, false
}
