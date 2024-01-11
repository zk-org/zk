package term

import (
	"testing"

	"github.com/fatih/color"
	"github.com/zk-org/zk/internal/core"
	"github.com/zk-org/zk/internal/util/test/assert"
)

func createTerminal() *Terminal {
	color.NoColor = false // Otherwise the color codes are not injected during tests
	return New()
}

func TestStyleNoRule(t *testing.T) {
	res, err := createTerminal().Style("Hello")
	assert.Nil(t, err)
	assert.Equal(t, res, "Hello")
}

func TestStyleOneRule(t *testing.T) {
	res, err := createTerminal().Style("Hello", core.Style("red"))
	assert.Nil(t, err)
	assert.Equal(t, res, "\033[31mHello\033[0m")
}

func TestStyleMultipleRule(t *testing.T) {
	res, err := createTerminal().Style("Hello", core.Style("red"), core.Style("bold"))
	assert.Nil(t, err)
	assert.Equal(t, res, "\033[31;1mHello\033[0m")
}

func TestStyleUnknownRule(t *testing.T) {
	_, err := createTerminal().Style("Hello", core.Style("unknown"))
	assert.Err(t, err, "unknown styling rule: unknown")
}

func TestStyleEmptyString(t *testing.T) {
	res, err := createTerminal().Style("", core.Style("bold"))
	assert.Nil(t, err)
	assert.Equal(t, res, "")
}

func TestStyleAllRules(t *testing.T) {
	styler := createTerminal()
	test := func(rule string, expected string) {
		res, err := styler.Style("Hello", core.Style(rule))
		assert.Nil(t, err)
		assert.Equal(t, res, "\033["+expected+"Hello\033[0m")
	}

	test("title", "1;33m")
	test("path", "4;36m")
	test("term", "31m")
	test("emphasis", "1;36m")
	test("understate", "2m")

	test("bold", "1m")
	test("faint", "2m")
	test("italic", "3m")
	test("underline", "4m")
	test("blink", "5m")
	test("reverse", "7m")
	test("hidden", "8m")
	test("strikethrough", "9m")

	test("black", "30m")
	test("red", "31m")
	test("green", "32m")
	test("yellow", "33m")
	test("blue", "34m")
	test("magenta", "35m")
	test("cyan", "36m")
	test("white", "37m")

	test("black-bg", "40m")
	test("red-bg", "41m")
	test("green-bg", "42m")
	test("yellow-bg", "43m")
	test("blue-bg", "44m")
	test("magenta-bg", "45m")
	test("cyan-bg", "46m")
	test("white-bg", "47m")

	test("bright-black", "90m")
	test("bright-red", "91m")
	test("bright-green", "92m")
	test("bright-yellow", "93m")
	test("bright-blue", "94m")
	test("bright-magenta", "95m")
	test("bright-cyan", "96m")
	test("bright-white", "97m")

	test("bright-black-bg", "100m")
	test("bright-red-bg", "101m")
	test("bright-green-bg", "102m")
	test("bright-yellow-bg", "103m")
	test("bright-blue-bg", "104m")
	test("bright-magenta-bg", "105m")
	test("bright-cyan-bg", "106m")
	test("bright-white-bg", "107m")
}
