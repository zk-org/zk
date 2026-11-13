# Notebook configuration

The `[notebook]` section from the [configuration file](config.md) is used to set
the default notebook directory and to choose how the note index is refreshed.

```toml
[notebook]
dir = "~/notebook" # same as "$HOME/notebook"
scan = false
```

The following properties are customizable:

- `dir` (string)
  - Path of the default notebook.
  - Only available in the global config file (`$ZK_CONFIG_DIR/config.toml` or
    `$XDG_CONFIG_HOME/zk/config.toml`).
- `scan` (boolean)
  - Whether `zk` checks the notebook for changes before each command.
  - `true` (default) scans the whole notebook directory and updates the index
    for notes added, modified, removed or renamed outside of `zk`. Use this if
    you edit notes with other tools.
  - `false` skips this scan. `zk` updates the index when it writes a note; run
    `zk index` to pick up changes made by other tools. Until then, new or
    modified notes are missing from the index, and removed notes remain listed.
    A notebook that has not been indexed yet, or needs to be reindexed, is still
    scanned. This can make commands much faster on large or network-mounted
    notebooks.
  - With `false`, `zk new` and `zk edit` update edited notes after the editor
    exits. For blocking editors, use a blocking command (for example,
    `code --wait`) so edits are saved before it exits.
