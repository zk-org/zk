# Template context when formatting a note

The following variables are available in the templates used when formatting notes, for example with `zk list --format <template>`.

| Variable        | Type     | Description                                                              |
|-----------------|----------|--------------------------------------------------------------------------|
| `filename`      | string   | Filename of the note, including its extension                            |
| `filename-stem` | string   | Filename of the note without the file extension                          |
| `path`          | string   | File path to the note, relative to the current directory                 |
| `abs-path`      | string   | File path to the note, absolute path including the notebook directory    |
| `title`         | string   | Note title                                                               |
| `link`          | string   | Markdown link to the note, relative to the current directory<sup>1</sup> |
| `lead`          | string   | First paragraph extracted from the note content                          |
| `body`          | string   | All of the note content, minus the heading                               |
| `snippets`      | [string] | List of context-sensitive relevant excerpts from the note                |
| `raw-content`   | string   | The full raw content of the note file                                    |
| `word-count`    | int      | Number of words in the note                                              |
| `tags`          | [string] | List of tags found in the note                                           |
| `metadata`      | map      | YAML frontmatter metadata, e.g. `metadata.description`<sup>2</sup>       |
| `created`       | date     | Date of creation of the note                                             |
| `modified`      | date     | Last date of modification of the note                                    |
| `checksum`      | string   | SHA-256 checksum of the note file                                        |

1. The format of the generated Markdown links can be customized in the [note format configuration](note-format.md).
2. YAML keys are normalized to lower case.
