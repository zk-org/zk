package zk

import (
	"testing"

	"github.com/mickael-menu/zk/util/assert"
)

// Parse a minimal configuration file.
func TestParseMinimal(t *testing.T) {
	config, err := parseConfig([]byte(""))

	assert.Nil(t, err)
	assert.Equal(t, config, &Config{rootConfig{}})
}

// Parse a complete configuration file.
func TestParseComplete(t *testing.T) {
	config, err := parseConfig([]byte(`
		// Comment
		editor = "vim"
		extension = "note"
		filename = "{{random-id}}.note"
		template = "default.note"
		random_id {
			charset = "alphanum"
			length = 4
			case = "lower"
		}
		ext = {
			hello = "world"
			salut = "le monde"
		}
		dir "log" {
			filename = "{{date}}.md"
			template = "log.md"
			ext = {
				log-ext = "value"
			}
		}
	`))

	assert.Nil(t, err)
	assert.Equal(t, config, &Config{rootConfig{
		Editor:    "vim",
		Extension: "note",
		Filename:  "{{random-id}}.note",
		Template:  "default.note",
		RandomID: &randomIDConfig{
			Charset: "alphanum",
			Length:  4,
			Case:    "lower",
		},
		Dirs: []dirConfig{
			dirConfig{
				Dir:      "log",
				Filename: "{{date}}.md",
				Template: "log.md",
				Ext:      map[string]string{"log-ext": "value"},
			},
		},
		Ext: map[string]string{
			"hello": "world",
			"salut": "le monde",
		},
	}})
}

// Parsing failure
func TestParseInvalidConfig(t *testing.T) {
	config, err := parseConfig([]byte("unknown = 'value'"))

	assert.NotNil(t, err)
	assert.Nil(t, config)
}
