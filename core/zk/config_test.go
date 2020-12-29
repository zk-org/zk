package zk

import (
	"testing"

	"github.com/mickael-menu/zk/util/assert"
)

func TestParseMinimal(t *testing.T) {
	conf, err := parseConfig([]byte(""))

	assert.Nil(t, err)
	assert.Equal(t, conf, &config{})
}

func TestParseComplete(t *testing.T) {
	conf, err := parseConfig([]byte(`
		// Comment
		editor = "vim"
		filename = "{{id}}.note"
		template = "default.note"
		id {
			charset = "alphanum"
			length = 4
			case = "lower"
		}
		extra = {
			hello = "world"
			salut = "le monde"
		}
		dir "log" {
			filename = "{{date}}.md"
			template = "log.md"
			id {
				charset = "letters"
				length = 8
				case = "mixed"
			}
			extra = {
				log-ext = "value"
			}
		}
	`))

	assert.Nil(t, err)
	assert.Equal(t, conf, &config{
		Filename: "{{id}}.note",
		Template: "default.note",
		ID: &idConfig{
			Charset: "alphanum",
			Length:  4,
			Case:    "lower",
		},
		Editor: "vim",
		Dirs: []dirConfig{
			{
				Dir:      "log",
				Filename: "{{date}}.md",
				Template: "log.md",
				ID: &idConfig{
					Charset: "letters",
					Length:  8,
					Case:    "mixed",
				},
				Extra: map[string]string{"log-ext": "value"},
			},
		},
		Extra: map[string]string{
			"hello": "world",
			"salut": "le monde",
		},
	})
}

func TestParseInvalidConfig(t *testing.T) {
	conf, err := parseConfig([]byte("unknown = 'value'"))

	assert.NotNil(t, err)
	assert.Nil(t, conf)
}
