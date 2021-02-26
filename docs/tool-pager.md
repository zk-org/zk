# Setting your default pager

When `zk`'s output exceeds a certain limit, it is automatically paginated by your system pager. By default, `less` is used but you may set up your own pager in the [configuration file](config.md) or environment variables. In order of precedence, `zk` will use:

1. `ZK_PAGER` environment variable
2. `pager` configuration property
    ```toml
    [tool]
    pager = "less -FIRX"
    ```
3. `PAGER` environment variable

## Disable the pager

If you need to disable paging, you can either:

* use `--no-pager`
* set the `pager` configuration property to an empty string `""`
