package tty

type TTY struct {
	NoInput bool
}

func New() *TTY {
	return &TTY{}
}
