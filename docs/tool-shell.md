# Setting your default shell

This is *currently* not supported on Windows (that defaults always to `cmd`).

You can customize which shell to use to run aliases and commands either from the [configuration file](config.md) or environment variables. In order of precedence, `zk` will use:

1. `ZK_SHELL` environment variable
2. `shell` configuration property
    ```toml
    [tool]
    shell = "/bin/bash"
    ```
3. `SHELL` environment variable
4. `sh` as fallback
