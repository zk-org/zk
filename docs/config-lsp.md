# LSP configuration

The `[lsp]` [configuration file](config.md) section provides settings to fine-tune the [LSP editors integration](editors-integration.md).

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
```