# Template context when formatting a note

The following variables are available in the templates used when formatting notes, for example with `zk list --format <template>`.

| Variable      | Type     | Description                                               |
|---------------|----------|-----------------------------------------------------------|
| `path`        | string   | File path to the note, relative to the current directory  |
| `title`       | string   | Note title                                                |
| `lead`        | string   | First paragraph extracted from the note content           |
| `body`        | string   | All of the note content, minus the heading                |
| `snippets`    | [string] | List of context-sensitive relevant excerpts from the note |
| `raw-content` | string   | The full raw content of the note file                     |
| `word-count`  | int      | Number of words in the note                               |
| `tags`        | [string] | List of tags found in the note                            |
| `created`     | date     | Date of creation of the note                              |
| `modified`    | date     | Last date of modification of the note                     |
| `checksum`    | string   | SHA-256 checksum of the note file                         |

