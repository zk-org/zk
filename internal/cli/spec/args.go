package spec

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/zk-org/zk/internal/util/paths"
)

// BindArgs fills in a command's positional arguments, and reports a missing
// or unexpected one.
func (c *Command) BindArgs(args []string) error {
	for i, arg := range c.Args {
		if arg.Variadic {
			values, err := bindValues(arg, args[min(i, len(args)):])
			if err != nil {
				return err
			}
			*arg.Target.(*[]string) = append(*arg.Target.(*[]string), values...)
			return nil
		}

		if i >= len(args) {
			if !arg.Optional {
				return fmt.Errorf("expected %q", "<"+arg.Name+">")
			}
			if arg.Default == "" {
				continue
			}
			values, err := bindValues(arg, []string{arg.Default})
			if err != nil {
				return err
			}
			*arg.Target.(*string) = values[0]
			continue
		}

		values, err := bindValues(arg, []string{args[i]})
		if err != nil {
			return err
		}
		*arg.Target.(*string) = values[0]
	}

	if len(args) > len(c.Args) {
		return fmt.Errorf("unexpected argument %s", args[len(c.Args)])
	}
	return nil
}

// MaxArgs returns how many positional arguments the command accepts, and
// whether it accepts any number of them.
func (c *Command) MaxArgs() (int, bool) {
	for _, arg := range c.Args {
		if arg.Variadic {
			return 0, true
		}
	}
	return len(c.Args), false
}

func bindValues(arg Arg, values []string) ([]string, error) {
	if !arg.Path {
		return values, nil
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		expanded, err := paths.ExpandPath(value)
		if err != nil {
			return nil, err
		}
		absolute, err := filepath.Abs(expanded)
		if err != nil {
			return nil, err
		}
		out = append(out, absolute)
	}
	return out, nil
}

// ValidateFlags reports a required flag which was left out, and a flag set to
// a value it doesn't accept. isSet tells whether the user set a given flag,
// which only the front-end's parser knows.
func ValidateFlags(flags []Flag, isSet func(name string) bool) error {
	missing := []string{}
	for _, flag := range flags {
		if flag.Required && !isSet(flag.Name) {
			missing = append(missing, "--"+flag.Name+"="+PlaceholderFor(flag))
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing flags: %s", strings.Join(missing, ", "))
	}

	for _, flag := range flags {
		if len(flag.OneOf) == 0 || !isSet(flag.Name) {
			continue
		}
		value := *flag.Target.(*string)
		if !contains(flag.OneOf, value) {
			quoted := make([]string, 0, len(flag.OneOf))
			for _, candidate := range flag.OneOf {
				quoted = append(quoted, fmt.Sprintf("%q", candidate))
			}
			return fmt.Errorf("--%s must be one of %s but got %q", flag.Name, strings.Join(quoted, ","), value)
		}
	}

	return nil
}

func contains(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}
