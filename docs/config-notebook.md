# Notebook configuration

The `[notebook]` section from the [configuration file](config.md) is used to set the default notebook directory.
If the path starts with _~_ it will be replaced with the user home directory (_$HOME_), this configuration also supports any environment variable.
```toml
[notebook]
dir = "~/notebook" # same as "$HOME/notebook"
```

 The following properties are customizable:

* `dir` (string)
    * Path of the default notebook.
