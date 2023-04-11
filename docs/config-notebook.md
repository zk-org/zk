# Notebook configuration

The `[notebook]` section from the [configuration file](config.md) is used to set the default notebook directory.
If the path starts with `~` it will be replaced with the user home directory (`$HOME`). This property also supports environment variables.

```toml
[notebook]
dir = "~/notebook" # same as "$HOME/notebook"
```

 The following properties are customizable:

* `dir` (string)
    * Path of the default notebook.
    * Only available in the global config file (`~/.config/zk/config.toml`).
