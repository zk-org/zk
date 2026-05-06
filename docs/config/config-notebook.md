# Global notebook

A `[notebook]` section can be added to the [configuration file](config.md) to
set a global notebook directory.

`~` and environment variables will be expanded.

```toml
[notebook]
dir = "~/notebook" # same as "$HOME/notebook"
```

The following properties are customizable:

- `dir` (string)
  - Path of the default notebook.
  - Only available in the global config file (`$ZK_CONFIG_DIR/config.toml` or
    `$XDG_CONFIG_HOME/zk/config.toml`).

The global notebook can also be set via the
[environment variable `$ZK_NOTEBOOK_DIR`](../notes/notebook.md).
