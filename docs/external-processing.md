# Send notes for processing by other programs

A great way to expand `zk` feature set is to explore a wealth of command-line tools available. You can use `zk`'s powerful [searching and filtering](note-filtering.md) capabilities to select notes before delegating further processing to other programs.

## Process file paths

Many programs expect file paths for input. You can interface with such program using the `path` list format and a space delimiter.

```sh
$ zk list --format path --delimiter " "
```

If the file paths can contain spaces, you may want to quote manually the paths, using the `{{path}}` [template variable](template-format.md) instead:

```sh
$ zk list --format "'{{path}}'" --delimiter " "
```

As always, this is such a useful [command alias](config-alias.md) to have:

```toml
paths = "zk list --format \"'{{path}}'\" --quiet --delimiter ' ' $@"
```

Some programs – such as `xargs` – work better when file paths are separated by the ASCII NUL character (`\0`). In this case, you can use the `--delimiter0` (or `-0`) option.

For example, this command prints the full Git history of the notes:

```sh
$ zk list --format path --delimiter0 | xargs -0 git log --patch --
```

### Feeding `zk` to itself

Some `zk` options such as `--exclude` also take file paths for parameters. Let's increase their flexibility by nesting `zk` calls. In this case, the delimiter will be `,`.

For example, this command lists the notes which are linked by at least one other note – so the notes which are *not* orphans.

```sh
$ zk list --exclude "`zk list -q -f path -d "," --orphan`"
```

And this one finds the notes which are linked by at least one note in `journal/`.

```sh
$ zk list --linked-by "`zk list -q -f path -d "," journal`"
```

## Process the content of a note

If you want to directly transform the content instead, you may use the `raw-content` template variable, which will print the full content of the note file.

In this particular case, we usually want to process only one note at a time. You can make sure that `zk list` will print only the first note from the result with `--limit 1` (or `-n1`).

```sh
$ zk list --format {{raw-content}} --limit 1
```

