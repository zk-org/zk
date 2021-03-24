# Named filter

A named filter is a set of [note filtering options](note-filtering.md) used frequently together, declared in the [configuration file](config.md).

For example, if you use regularly the following command to list your most recent notes:

```sh
$ zk list --sort created- --created-after "last two weeks"
```

You can create a new named filter in the configuration file to avoid repeating yourself.

```toml
[filter]
recents = "--sort created- --created-after 'last two weeks'"
```

Then, you can use the name as an argument of `zk list`, with any additional option.

```sh
$ zk list recents --limit 10
```

Named filters are similar to [command aliases](config-alias.md), as they simplify frequent commands. However, named filters can be used with any command accepting filtering options.

```sh
$ zk edit recents --interactive
```

## Filter named after a directory

In filtering commands, named filters take precedence over path arguments. As a nice side effect, this means you can customize the default filtering options for a directory by naming a filter after it.

For example, by default `zk` sorts notes by their titles. However, if you keep daily notes under a `journal/` directory, you may want to sort them by creation date instead. You can use the following named filter for this:

```
[filter]
journal = "--sort created journal"
```

Named filters cannot call themselves recursively, so by adding the `journal` argument to the filter, we are actually selecting the `journal/` directory. This means that the following commands are equivalent:

```sh
# Without the filter
$ zk list --sort created journal

# With the filter
$ zk list journal
```
