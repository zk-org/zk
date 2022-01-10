package term

import (
	"fmt"
	"os"
	"strings"

	survey "github.com/AlecAivazis/survey/v2"
	"github.com/mattn/go-isatty"
)

// Terminal offers utilities to interact with the terminal.
type Terminal struct {
	NoInput    bool
	ForceInput string
}

func New() *Terminal {
	return &Terminal{}
}

// IsInteractive returns whether the app is attached to an interactive terminal
// and can prompt the user.
func (t *Terminal) IsInteractive() bool {
	return !t.NoInput && t.IsTTY()
}

// IsTTY returns whether the app is attached to an interactive terminal.
func (t *Terminal) IsTTY() bool {
	return isatty.IsTerminal(os.Stdin.Fd())
}

// SupportsUTF8 returns whether the computer is configured to support UTF-8.
func (t *Terminal) SupportsUTF8() bool {
	lang := strings.ToUpper(os.Getenv("LANG"))
	lc := strings.ToUpper(os.Getenv("LC_ALL"))
	return strings.Contains(lang, "UTF") || strings.Contains(lc, "UTF")
}

// Confirm is a shortcut to prompt a yes/no question to the user.
func (t *Terminal) Confirm(msg string, defaultAnswer bool) (confirmed, skipped bool) {
	if !t.IsInteractive() {
		switch strings.ToLower(t.ForceInput) {
		case "y":
			return t.forceConfirm(msg, true)
		case "n":
			return t.forceConfirm(msg, false)
		default:
			return defaultAnswer, true
		}
	}

	confirmed = false
	prompt := &survey.Confirm{
		Message: msg,
		Default: defaultAnswer,
	}
	survey.AskOne(prompt, &confirmed)
	return confirmed, false
}

func (t *Terminal) forceConfirm(msg string, answer bool) (confirmed, skipped bool) {
	msg = "? " + msg + " ("
	if answer {
		msg += "Y/n"
	} else {
		msg += "y/N"
	}
	msg += ")"
	fmt.Println(msg)

	return answer, false
}
