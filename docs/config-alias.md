# Command aliases

A command alias is a custom `zk` command which can run another `zk` command or an external program.

Declaring your own aliases is a great way to make your experience with `zk` easier and more familiar. With aliases, `zk` becomes a hub capable of launching all the programs you need to manage your [notebook](notebook.md).

## Configuring aliases

Command aliases are declared in your [configuration file](config.md), under the `[alias]` section. They are executed with [your default shell](tool-shell.md), which allows you to:

* expand arguments with `$@` or `$*`
    * [it is recommended to wrap `$@` in quotes](https://github.com/zk-org/zk/issues/316#issuecomment-1543564168)
* expand environment variables
* run several commands with `&&`
* pipe several commands with `|`

An alias can call other aliases but cannot call itself. This enables you to override the default options of native commands, for example:

```toml
[alias]
edit = 'zk edit --interactive "$@"'
```

When running an alias, the `ZK_NOTEBOOK_DIR` environment variable is set to the absolute path of the current notebook. You can use it to run commands working no matter the location of the working directory.

```toml
journal = 'zk new "$ZK_NOTEBOOK_DIR/journal"'
```

If you need to surround the path with quotes, make sure you use double quotes, otherwise environment variables will not be expanded.

### `xargs` formula

Calling an external program with a list of note paths using `xargs` is such a common use case that we can extract a reusable alias pattern.

```toml
alias = "zk list --quiet --format path --delimiter0 $@ | xargs -0 <EXTERNAL COMMAND>"
```

Find more details about these options in [Send notes for processing by other programs](external-processing.md).

## Collection of useful aliases

Here are a few aliases to get you started writing your own.

### Shortcuts for native commands

```toml
ls = "zk list $@"
ed = "zk edit $@"
n = "zk new $@"
```

### Edit the last modified note

Suffixing the `modified` sort criterion with `-` orders the notes by *descendent* modification date.

```toml
edlast = "zk edit --limit 1 --sort modified- $@"
```

### Edit the notes created during the last two weeks

This command uses `--interactive` to let the user select which notes to actually edit among the recent ones. Note the use of human friendly language for `--created-after`'s argument. 

In this case, additional arguments do not necessarily make sense, so we omit the trailing `$@`.

```toml
recent = "zk edit --sort created- --created-after 'last two weeks' --interactive"
```

This kind of alias might be more useful as a [named filter](config-filter.md).

### Edit the configuration file

Here's a concrete example using environment variables, in particular `ZK_NOTEBOOK_DIR`. Note the double quotes around the path.

```toml
conf = '$EDITOR "$ZK_NOTEBOOK_DIR/.zk/config.toml"'
```

### List paths in a command-line friendly fashion

Use this alias to send a list of space-separated file paths matching the given [filtering criteria](note-filtering.md) to another program. See [send notes for processing by other programs](external-processing.md) for more details.

```toml
paths = "zk list --quiet --format \"'{{path}}'\" --delimiter ' ' $@"
```

### List paths to be used in a parent `zk` command

Similarly, use this alias to expand filtered note paths inside a parent `zk` command taking a comma-separated paths list.

```toml
inline = "zk list --quiet --format {{path}} --delimiter , $@"
```

Examples of use:

```sh
# List the notes which have at least one link pointing to them (i.e. not orphans).
$ zk list --exclude "`zk inline --orphan`"

# List the notes which are linked by at least one note from the journal/ directory.
$ zk list --linked-by "`zk inline journal`"
```

### Print a random note

Increasing serendipity while using your notebook is important to spark new ideas. The `random` sort criterion is the key to this alias.

```toml
lucky = "zk list --quiet --format full --sort random --limit 1"
```

### Create a note from a free title

If you often create notes with `zk new --title "An interesting concept"`, you will like this alias. Using `"$*"`, you do not need to quote the arguments anymore.

```toml
nt = 'zk new --title "$*"'
```

Usage: `zk nt An interesting concept`

No more forgotten quotes!

### Create a note and save its path into the clipboard (macOS)

Build upon the previous alias, but instead of editing the created note it will copy the created note's path into the macOS clipboard.

```toml
ntc = 'zk new --print-path --title "$*" | pbcopy'
```

### Print and sort the word count of selected notes

This will list the notes and their word count sorted by increasing word count. It is useful to spot flimsy notes that you could flesh out.

```toml
wc = "zk list --format '{{word-count}}\t{{title}}' --sort word-count $@"
```

Usage:

```sh
$ zk wc
4       Integration with fzf
5       Searching and filtering notes
63      Setting your default editor
86      Anatomy of a notebook
...
```

### Print the backlinks of a note

This is such a useful command, that an alias might be helpful.

```toml
bl = "zk list --link-to $@"
```

### Locate unlinked mentions in a note

This alias can help you look for potential new links to establish, by listing every note whose title is mentioned in the note you are working on but which are not already linked to it.

Note that we are using a single argument `$1` which is repeated for both options.

```toml
unlinked-mentions = "zk list --mentioned-by $1 --no-linked-by $1"
```

### Browse the Git history of selected notes

This example showcases the "`xargs` formula" with a concrete example.

```toml
log = "zk list --quiet --format path --delimiter0 $@ | xargs -0 git log --patch --"
```

### Saving the changes in the Git repository

This alias does not call `zk` at all! This shows how you can use `zk` as a hub for everything related to your notes.

```toml
save = 'git add . && git commit -m "$*"'
```

Usage: `zk save Expand the note on command aliases`

### Copy/backup selected notes

A more complex example backing up the notes matching the given filtering criteria in a target directory. It creates intermediate directories if needed.

`$1` and `${@:2}` are used to split the arguments between the first one which will be the destination directory, and the remaining arguments which will be used as filtering options.

```toml
# macOS
cp = 'zk list --quiet --format path --delimiter0 ${@:2} | xargs -t -0 -I % ditto "%" "$1/%"'

# Linux
cp = 'mkdir -p "$1" && zk list --quiet --format path --delimiter0 ${@:2} | xargs -t -0 -I % cp --parents "%" "$1"'
```

Usage: `zk cp output/ --created-after 'last two weeks'`

