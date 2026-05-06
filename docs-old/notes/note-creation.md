# Creating a new note

You can add a new note to a [notebook](notebook.md) using
`zk new --title "An interesting concept" [<directory>]`.

`zk` automatically generates a filename and initial content according to rules
set in your [configuration file](../config/config.md). These settings can be customized
per [group of notes](../config/config-group.md) in your notebook, as illustrated in
[Maintaining a daily journal](../tips/daily-journal.md).

By default, `zk new` will start [your editor](../config/tool-editor.md) after creating the
note. You can choose instead to print the absolute path to the note with
`--print-path`, which is more useful for [automation](../tips/automation.md).

(test)=
## Search or create with a single command

If you are not sure whether a note already exists for a particular subject, the
"search or create" mode might be more appropriate than `zk new`. It is inspired
by [Notational Velocity](https://notational.net/) and enables searching for an
existing note or creating a new one in a single action.

This option is available when running `zk edit --interactive`, which spawns
[`fzf`](../config/tool-fzf.md) to filter selected notes. From `fzf`, press `Ctrl-E` to
create a new note using the current search query as title.

## Create a note with initial content

Initial content can be fed to the template through standard input using
`zk new --interactive`, which will be expandable with the `{{content}}`
[template variable](template-creation.md).

For example, to use the content of the macOS clipboard as the initial content
you can run:

```sh
$ pbpaste | zk new --interactive
```

Alternatively, you can use the content of a file:

```sh
$ zk new --interactive < file.txt
```
