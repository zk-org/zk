# Note group

A _group_ is a named [configuration section](config.md) used to override
[note creation rules](config-note.md) for specific directories. This allows you
to use your [notebook](../notes/notebook.md) very differently depending on the type of
note created. For a practical example, take a look at
[maintaining a daily journal](../tips/daily-journal.md).

## Declaring a new group

To add a new group to your configuration file, declare a new `[group.<name>]`
section. It takes a single optional property `paths`, which is the list of
directories belonging to this group.

```toml
[group.journal]
paths = [
    "journal/daily",
    "journal/weekly"
]
```

You can also use
[glob patterns](https://en.wikipedia.org/wiki/Glob_(programming)) in `paths`.

```toml
[group.journal]
paths = ["journal/*"]
```

If you omit `paths`, the directory named after the group will be inferred. Note
the double quotes when using spaces or slashes for subdirectories.

```toml
# This will automatically apply to the `citations/web` directory
[group."citations/web"]
```

## Overriding note configuration and extra variables

You can override the global [note configuration](config-note.md) and
[extra user variables](config-extra.md) for a given group.

```toml
[group.journal.note]
filename = "{{format-date now}}"
template = "journal.md"

[group.journal.extra]
author = "Mickaël"
```

## Choose a group dynamically

If you prefer to keep multiple groups in a single directory, you can specify
which group to use when creating a new note explicitly.

```sh
$ zk new --group journal
```
