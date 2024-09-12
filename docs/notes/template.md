# Template syntax

`zk` uses the [Handlebars template syntax](https://handlebarsjs.com/guide) for
its templates. The list of variables available depends of the running command:

- [Template context when creating notes](template-creation.md) (i.e. `zk new`)
- [Template context when formatting a note](template-format.md) (i.e.
  `zk list --format <template>`)

## Additional helpers

Besides the default Handlebars helpers, `zk` ships with additional helpers which
you might find useful. They are available to all templates.

### Format Link helper

The `{{format-link}}` helper renders an internal link to another note, according
to the user preferences set in the [note formats configuration](note-format.md).

```
{{format-link "path/to note.md" "An interesting note"}}

can generate (depending on the user config):

[An interesting note](path/to%20note.md)
[[path/to note]]
```

The second parameter `title` is optional.

### String helpers

There are a couple of template helpers operating on strings.

#### Concat helper

The `{{concat s1 s2}}` helper concatenates two strings together. For example
`{{concat '> ' 'A quote'}}` produces `> A quote`.

#### Substring helper

- The `{{substring s index length}}` helper extracts a portion of the given
  string. For example:
  - `{{substring 'A full quote' 2 4}}` outputs `full`
  - `{{substring 'A full quote' -5 5}}` outputs `quote`

### Date helpers

#### Date from natural string helper

You can get a date object from a natural human date (e.g. `tomorrow`,
`2 weeks ago`, `2022-03-24`) using the `{{date}}` helper. It is most useful when
paired with the `{{format-date}}` helper.

```
{{date "tomorrow"}}

{{format-date (date "last week") "timestamp"}}
```

#### Date formatting helper

The `{{format-date}}` helper formats the given date for display.

Template contexts usually provide a `now` variable which can be used to print
the current date.

The default format output by `{{format-date <variable>}}` looks like
`2009-11-17`, but you can choose a different format by providing a second
argument, e.g. `{{format-date now "medium"}}`.

| Format           | Output                     | Notes                                            |
| ---------------- | -------------------------- | ------------------------------------------------ |
| `short`          | 11/17/2009                 |                                                  |
| `medium`         | Nov 17, 2009               |                                                  |
| `long`           | November 17, 2009          |                                                  |
| `full`           | Tuesday, November 17, 2009 |                                                  |
| `year`           | 2009                       |                                                  |
| `time`           | 20:34                      |                                                  |
| `timestamp`      | 200911172034               | Useful for sortable filenames                    |
| `timestamp-unix` | 1258490098                 | Number of seconds since January 1, 1970          |
| `elapsed`        | 12 years ago               | Time elapsed since then in human-friendly format |

If none of the provided formats suit you, you can use a custom format using
`strftime`-style placeholders, e.g. `{{format-date now "%m-%d-%Y"}}`. See
`man strftime` for a list of placeholders.

### Slug helper

The `{{slug}}` helper generates a URL friendly version of a text. For example,
`{{slug "This will be slugified!"}}` becomes `this-will-be-slugified`.

This is mostly useful to generate a safe filename containing the title passed to
`zk new --title "An interesting note"`. With the [`filename`](../config/config-note.md)
template `{{slug title}}`, it becomes `an-interesting-note.md`.

### Prepend helper

The `{{prepend}}` helper adds a prefix to every line of the given text or block.
You can use it to generate a Markdown quote, for example:

```
{{prepend "> " "A quote"}}

{{#prepend "> "}}
A multiline
quote.
{{/prepend}}
```

### Shell helper

The `{{sh}}` helper will call the given shell command and insert its output in
the template. Your imagination is the limit!

```
Get today's events from your calendar:
{{sh "icalBuddy -b '* ' -nc eventsToday"}}

Insert a random quote:
{{prepend '> ' (sh 'fortune')}}

Download today's weather:
{{sh 'curl http://wttr.in/?0'}}
```

When used as a block helper, the block content will be passed to the command
through a standard input pipe.

```
Will output "HELLO, WORLD!":
{{#sh "tr '[a-z]' '[A-Z]'"}}
Hello, world!
{{/sh}}
```

### Style helper

The `{{style}}` helper is mostly useful when formatting content for the
command-line. See the [styling rules](../tips/style.md) for more information.

```
{{style 'red bold' 'A text'}}

{{#style 'underline'}}Another text{{/style}}
```

### JSON helper

The `{{json}}` helper serializes its argument to a JSON value. This is useful to
generate valid JSON objects, for example:

```
{ "title": {{json title}}, "tags": {{json tags}} }
->
{ "title": "A \"quoted\" title", "tags": ["example", "json"] }
```

**Warning**: The template parser trips on `}}}`, so make sure to add an extra
space before the third `}`.

You can serialize the whole template context as a JSON object with `{{json .}}`,
which is how `zk list --format json` produces its output.
