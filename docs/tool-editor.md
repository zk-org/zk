# Setting your default editor

`zk` is not a text editor. Instead, it is designed to interface with your favorite editor to write your notes.

You can customize which editor to use either from the [configuration file](config.md) or environment variables. In order of precedence, `zk` will use:

1. `ZK_EDITOR` environment variable
2. `editor` configuration property
    ```toml
    [tool]
    editor = "vim"
    ```
3. `VISUAL` environment variable
4. `EDITOR` environment variable
