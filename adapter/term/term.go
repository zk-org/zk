package term

import (
	"os"

	"github.com/mattn/go-isatty"
)

// Terminal offers utilities to interact with the terminal.
type Terminal struct {
	NoInput bool
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
