# Notebook configuration

The `[notebook]` section from the [configuration file](config.md) is used to set
the default notebook directory and to choose how the note index is refreshed.
If the path starts with `~` it will be replaced with the user home directory
(`$HOME`). This property also supports environment variables.

```toml
[notebook]
dir = "~/notebook" # same as "$HOME/notebook"
```

The following properties are customizable:

- `scan` (boolean)
  - Whether the notebook is scanned before running a command.
  - `true` (default) walks the whole notebook directory to detect notes added,
    modified, removed or renamed outside of `zk`. Use it if you edit your notes
    with other tools.
  - `false` leaves the notebook alone: the index is then only refreshed by
    `zk index`, or by `zk` itself when it writes a note. This is much faster on
    large or network-mounted notebooks, but a note added, modified or removed
    outside of `zk` is not picked up until the next `zk index` — and a note
    removed outside of `zk` stays listed until then. A notebook that was never
    indexed, or that needs to be reindexed, is still scanned.
- `dir` (string)
  - Path of the default notebook.
  - Only available in the global config file (`$ZK_CONFIG_DIR/config.toml` or `$XDG_CONFIG_HOME/zk/config.toml`).
