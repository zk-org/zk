# Notebook configuration

The `[notebook]` section from the [configuration file](config.md) is used to set
the default notebook directory. If the path starts with `~` it will be replaced
with the user home directory (`$HOME`). This property also supports environment
variables.

```toml
[notebook]
dir = "~/notebook" # same as "$HOME/notebook"
```

The following properties are customizable:

- `index` (string)
  - Strategy used to refresh the note index before running a command.
  - `"walk"` (default) walks the whole notebook directory to detect notes added,
    modified, removed or renamed outside of `zk`. Use it if you edit your notes
    with other tools.
  - `"mtime"` only reads the modification time of the notes already indexed,
    which is much faster on large or network-mounted notebooks, but notes added,
    renamed or removed outside of `zk` are not picked up until a full walk. Run
    `zk index --force` after editing, adding or removing notes with another tool.
- `dir` (string)
  - Path of the default notebook.
  - Only available in the global config file (`$ZK_CONFIG_DIR/config.toml` or `$XDG_CONFIG_HOME/zk/config.toml`).
