# LSP configuration

The `[lsp]` [configuration file](config.md) section provides settings to fine-tune the [LSP editors integration](editors-integration.md).

## Completion

Customize how completion items appear in your editor when auto-completing links with the `[lsp.completion]` sub-section.

| Setting                     | Type       | Description                                                                           |
|-----------------------------|------------|---------------------------------------------------------------------------------------|
| `note-label`                | `template` | Label displayed in the completion pop-up for each note                                |
| `note-filter-text`          | `template` | Text used as a source when filtering the completion pop-up with keystrokes            |
| `note-detail`               | `template` | Additional information about a completion item                                        |
| `use-additional-text-edits` | `boolean`  | Indicates whether `additionalTextEdits` will be used to remove the trigger characters |

Each key accepts a [template](template.md) with the following context:

| Variable        | Type     | Description                                                        |
|-----------------|----------|--------------------------------------------------------------------|
| `filename`      | string   | Filename of the note, including its extension                      |
| `filename-stem` | string   | Filename of the note without the file extension                    |
| `path`          | string   | File path to the note, relative to the notebook root               |
| `abs-path`      | string   | Absolute file path to the note                                     |
| `rel-path`      | string   | File path to the note, relative to the current directory           |
| `title`         | string   | Note title                                                         |
| `title-or-path` | string   | Note title or path if empty                                        |
| `metadata`      | map      | YAML frontmatter metadata, e.g. `metadata.description`<sup>1</sup> |

1. YAML keys are normalized to lower case.


## Diagnostics

Use the `[lsp.diagnostics]` sub-section to configure how LSP diagnostics are reported to your editors. Each diagnostic setting can be:

* An empty string or `none` to ignore this diagnostic.
* `hint`, `info`, `warning` or `error` to enable and set the severity of the diagnostic.

| Setting      | Default   | Description                                                               |
|--------------|-----------|---------------------------------------------------------------------------|
| `wiki-title` | `"none"`  | Report titles of wiki-links, which is useful if you use IDs for filenames |
| `dead-link`  | `"error"` | Warn for dead links between notes                                         |

## Complete example

```toml
[lsp]

[lsp.diagnostics]
# Report titles of wiki-links as hints.
wiki-title = "hint"
# Warn for dead links between notes.
dead-link = "error"

[lsp.completion]
# Show the note title in the completion pop-up, or fallback on its path if empty.
note-label = "{{title-or-path}}"
# Filter out the completion pop-up using the note title or its path.
note-filter-text = "{{title}} {{path}}"
# Show the note filename without extension as detail.
note-detail = "{{filename-stem}}"
```
