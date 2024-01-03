# Integration with `fzf`

[`fzf`](https://github.com/junegunn/fzf) is an awesome and versatile fuzzy finder powering `zk`'s [interactive filtering mode](note-filtering.md).

Besides the standard [`fzf` configuration options](https://github.com/junegunn/fzf) documented on its website, `zk` offers additional options you can set in the `[tool]` [configuration section](config.md).

If you wish to customize more of `fzf` behavior, [please post a feature request](https://github.com/zk-org/zk/issues).

## Preview command

You can customize the command used to preview a note with `fzf-preview`. The special placeholder `{-1}` will be expanded to the note file path.

By default, `zk` uses `cat` for preview, which is a bit boring. A much better option would be to use [`bat`](https://github.com/sharkdp/bat) which supports syntax highlighting.

```toml
[tool]
fzf-preview = "bat -p --color always {-1}"
```

Or, if you prefer to preview more metadata, you can use a nested `zk` command.

```toml
[tool]
fzf-preview = "zk list --quiet --format full --limit 1 {-1}"
```

## Line format

With the `fzf-line` setting property, you can provide your own [template](template.md) to customize the format of each `fzf` line. The lines are used by `fzf` for the fuzzy matching, so if you want to search in the full note content, do not forget to add `{{body}}` in your custom template.

The default line template is `{{style "title" title-or-path}} {{style "understate" body}} {{style "understate" (json metadata)}}`.

Here's an example using different colors and showing the list of tags as #hashtags:

```toml
[tool]
fzf-line = "{{style 'blue' rel-path}}{{#each tags}} #{{this}}{{/each}} {{style 'black' body}}"
```

### Template context

The following variables are available in the line template.

| Variable        | Type     | Description                                                        |
|-----------------|----------|--------------------------------------------------------------------|
| `filename`      | string   | Filename of the note, including its extension                      |
| `filename-stem` | string   | Filename of the note without the file extension                    |
| `path`          | string   | File path to the note, relative to the notebook root               |
| `abs-path`      | string   | Absolute file path to the note                                     |
| `rel-path`      | string   | File path to the note, relative to the current directory           |
| `title`         | string   | Note title                                                         |
| `title-or-path` | string   | Note title or path if empty                                        |
| `body`          | string   | All of the note content, minus the heading                         |
| `raw-content`   | string   | The full raw content of the note file                              |
| `word-count`    | int      | Number of words in the note                                        |
| `tags`          | [string] | List of tags found in the note                                     |
| `metadata`      | map      | YAML frontmatter metadata, e.g. `metadata.description`<sup>1</sup> |
| `created`       | date     | Date of creation of the note                                       |
| `modified`      | date     | Last date of modification of the note                              |
| `checksum`      | string   | SHA-256 checksum of the note file                                  |

1. YAML keys are normalized to lower case.


## `fzf` options

You can override the default `fzf` options used by `zk` with `fzf-options`. Look at `man fzf` for the list of available options.

```toml
[tool]
fzf-options = "--height 40% --border"
```

Note that this overrides all the default options used by `zk`, you might want to keep some of them:

* `--tiebreak begin` Prefer matches located at the beginning of the line
* `--exact` Look for exact matches instead of fuzzy ones by default
* `--tabstop 4` Length of tab characters
* `--height 100%` Height of the list relative to the terminal window
* `--layout reverse` Display the input field at the top
* `--no-hscroll` Make sure the path and titles are always visible
* `--color hl:-1,hl+:-1` Don't highlight search terms
* `--preview-window wrap` Enable line wrapping in the preview window

## Key bindings

When running `fzf` with `zk edit --interactive`, you can [create a new note with the `Ctrl-E` key binding](note-creation.md#search-or-create-with-a-single-command). This binding is customizable with `fzf-bind-new`. You can also disable it by setting it to an empty string (`""`).

```toml
[tool]
fzf-bind-new = "Ctrl-C"
```
