package spec

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/pflag"
)

// Annotations a bound flag carries, for a front-end's help renderer to read
// back. They travel with the flag when flag sets are merged, which a lookup
// table on the side would not.
const (
	GroupAnnotation       = "zk_group"
	PlaceholderAnnotation = "zk_placeholder"
)

// BindFlags declares the given flags on a flag set, in order.
func BindFlags(fs *pflag.FlagSet, flags []Flag) {
	// zk lists its flags in the order they are declared, which groups the
	// related ones together far better than sorting them does.
	fs.SortFlags = false
	for _, flag := range flags {
		BindFlag(fs, flag)
	}
}

// BindFlag declares a single flag on a flag set.
func BindFlag(fs *pflag.FlagSet, flag Flag) {
	switch flag.Kind() {
	case KindBool:
		def, _ := flag.Default.(bool)
		fs.BoolVarP(flag.Target.(*bool), flag.Name, flag.Short, def, flag.Help)
	case KindInt:
		def, _ := flag.Default.(int)
		*flag.Target.(*int) = def
		fs.VarP(&intValue{target: flag.Target.(*int)}, flag.Name, flag.Short, flag.Help)
	case KindString:
		def, _ := flag.Default.(string)
		fs.StringVarP(flag.Target.(*string), flag.Name, flag.Short, def, flag.Help)
	case KindStringList:
		fs.VarP(&stringListValue{target: flag.Target.(*[]string)}, flag.Name, flag.Short, flag.Help)
	case KindStringMap:
		fs.VarP(&stringMapValue{target: flag.Target.(*map[string]string)}, flag.Name, flag.Short, flag.Help)
	}

	bound := fs.Lookup(flag.Name)
	bound.Hidden = flag.Hidden
	setAnnotation(bound, GroupAnnotation, flag.Group)
	setAnnotation(bound, PlaceholderAnnotation, PlaceholderFor(flag))
}

func setAnnotation(flag *pflag.Flag, key string, value string) {
	if value == "" {
		return
	}
	if flag.Annotations == nil {
		flag.Annotations = make(map[string][]string)
	}
	flag.Annotations[key] = []string{value}
}

// Annotation returns a bound flag's annotation, or "" when it has none.
func Annotation(flag *pflag.Flag, key string) string {
	if values, ok := flag.Annotations[key]; ok && len(values) > 0 {
		return values[0]
	}
	return ""
}

// PlaceholderFor returns the value standing for a flag's argument in the
// help, e.g. the PATH in `--notebook-dir=PATH`. A boolean flag takes no
// argument and gets an empty placeholder.
func PlaceholderFor(flag Flag) string {
	kind := flag.Kind()
	if kind == KindBool {
		return ""
	}

	repeatable := ""
	if kind == KindStringList || kind == KindStringMap {
		repeatable = ",..."
	}

	switch {
	case flag.Placeholder != "":
		return flag.Placeholder + repeatable
	case flag.DefaultString() != "":
		return fmt.Sprintf("%q", flag.DefaultString())
	case kind == KindStringMap:
		return "KEY=VALUE" + repeatable
	case kind == KindStringList:
		return strings.ToUpper(flag.Name) + repeatable
	case kind == KindInt:
		return "INT"
	default:
		return "STRING"
	}
}

// ParseFlags fills in the given flags from args and returns the positional
// arguments left over.
//
// This is how zk reads the filters a notebook's configuration defines, which
// mimic a command line, so that they mean the same thing wherever zk is
// driven from.
func ParseFlags(flags []Flag, args []string) ([]string, error) {
	fs := pflag.NewFlagSet("", pflag.ContinueOnError)
	// Errors are reported by the caller, with the context the user needs.
	fs.SetOutput(io.Discard)
	BindFlags(fs, flags)

	if err := fs.Parse(args); err != nil {
		return nil, FlagError(err)
	}
	return fs.Args(), nil
}

// pflagInvalidArgument matches pflag's wording for a value a flag rejected.
var pflagInvalidArgument = regexp.MustCompile(`^invalid argument "(?:.*)" for "(?:-\w, )?(--[^"]+)" flag: (.*)$`)

// FlagError rewrites a flag parsing error into the single line shape zk
// reports errors in, e.g. `--limit: expected a valid 64 bit int but got "a"`.
func FlagError(err error) error {
	if err == nil {
		return nil
	}
	message := err.Error()

	if matches := pflagInvalidArgument.FindStringSubmatch(message); matches != nil {
		return fmt.Errorf("%s: %s", matches[1], matches[2])
	}
	if rest, found := strings.CutPrefix(message, "unknown flag: "); found {
		return errors.New("unknown flag " + rest)
	}

	return err
}

// Value types for the flags zk accepts more than once.
//
// pflag's own StringSlice and StringToString parse their input as CSV, which
// would mangle a value containing double quotes, e.g.
// `--match '["न", "म"]'`. zk splits on commas verbatim instead.

// stringListValue is a repeatable flag whose values are also split on commas.
type stringListValue struct {
	target  *[]string
	changed bool
}

func (v *stringListValue) Set(value string) error {
	values := strings.Split(value, ",")
	if v.changed {
		*v.target = append(*v.target, values...)
	} else {
		*v.target = values
		v.changed = true
	}
	return nil
}

func (v *stringListValue) Type() string { return "stringList" }

func (v *stringListValue) String() string {
	if v.target == nil || len(*v.target) == 0 {
		return ""
	}
	return strings.Join(*v.target, ",")
}

// Append, Replace and GetSlice implement pflag.SliceValue, which shell
// completion relies on.
func (v *stringListValue) Append(value string) error {
	*v.target = append(*v.target, value)
	return nil
}

func (v *stringListValue) Replace(values []string) error {
	*v.target = values
	return nil
}

func (v *stringListValue) GetSlice() []string { return *v.target }

// stringMapValue is a repeatable `key=value` flag, whose pairs can also be
// delimited by commas.
type stringMapValue struct {
	target  *map[string]string
	changed bool
}

func (v *stringMapValue) Set(value string) error {
	if !v.changed || *v.target == nil {
		*v.target = make(map[string]string)
		v.changed = true
	}
	for _, pair := range strings.Split(value, ",") {
		key, val, found := strings.Cut(pair, "=")
		if !found {
			return fmt.Errorf("expected a KEY=VALUE pair but got %q", pair)
		}
		(*v.target)[key] = val
	}
	return nil
}

func (v *stringMapValue) Type() string { return "stringMap" }

func (v *stringMapValue) String() string {
	if v.target == nil || len(*v.target) == 0 {
		return ""
	}
	pairs := make([]string, 0, len(*v.target))
	for key, val := range *v.target {
		pairs = append(pairs, key+"="+val)
	}
	return strings.Join(pairs, ",")
}

// intValue is an int flag reporting invalid input the way zk always has.
type intValue struct {
	target *int
}

func (v *intValue) Set(value string) error {
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fmt.Errorf("expected a valid 64 bit int but got %q", value)
	}
	*v.target = int(parsed)
	return nil
}

func (v *intValue) Type() string { return "int" }

func (v *intValue) String() string {
	if v.target == nil || *v.target == 0 {
		return ""
	}
	return strconv.Itoa(*v.target)
}
