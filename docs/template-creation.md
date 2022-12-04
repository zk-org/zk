# Template context when creating notes

The following variables are available in the templates used when [creating new notes](note-creation.md) – both for the filename and the note content.

| Variable      | Type   | Description                                                                           |
|---------------|--------|---------------------------------------------------------------------------------------|
| `id`          | string | Random ID generated for this note                                                     |
| `title`       | string | Note title given to `--title`                                                         |
| `content`     | string | Any text piped through the standard input                                             |
| `dir`         | string | Parent directory in the notebook                                                      |
| `extra.<key>` | string | [Additional variables](config-extra.md) provided through the config file or `--extra` |
| `now`         | date   | Current date and time, useful when paired with [`{{format-date now}}`](template.md)   |
| `env`         | map    | Dictionary of case-sensitive environment variables, e.g. `{{env.PATH}}`.              |

These additional variables are available only to the note content template, once the filename is generated.

| Variable        | Type   | Description                                                    |
|-----------------|--------|----------------------------------------------------------------|
| `filename`      | string | Filename generated for this note, including the file extension |
| `filename-stem` | string | Filename without the file extension                            |

