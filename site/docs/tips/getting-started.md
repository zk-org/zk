# Getting started with `zk`

A short introduction showing how to use `zk`.

## Create a new notebook

Create a [notebook](../notes/notebook.md) to host your notes. You are free to organize
your notebook as you want, adding subdirectories if needed.

```sh
$ zk init my-notes
Initialized a notebook in my-notes

$ cd my-notes
```

## Create your first notes

Now you are ready to write your very first note. Pick a subject,
[create a new note](../notes/note-creation.md) and write on!

```sh
$ zk new --title "An interesting concept"
```

You can customize your experience using [custom templates](../notes/template.md) to
generate many kind of notes.

![Create a note](../assets/media/new1.svg)

If you are not sure whether a note already exists for a particular subject, the
"search or create" mode might be more appropriate than `zk new`. It is inspired
by [Notational Velocity](https://notational.net/) and enables searching for an
existing note or creating a new one in a single action.

From `zk`'s interactive edit screen, press `Ctrl-E` to create a new note using
the current search query as title.

![Create a note](../assets/media/new2.svg)

## List existing notes

After some time, hopefully you will have enough notes to be lost in it.

To help structure your notebook, you can add [metadata](../notes/note-frontmatter.md)
(e.g. keywords/tags) to your notes. You can then use `zk`'s powerful
[filtering capabilities](../notes/note-filtering.md) to find the notes you need.

```sh
$ zk list --tag "recipe" --match "pizza -pineapple"
```

![List notes](../assets/media/list.svg)

Sort the results however you need with `--sort`.

![Sort notes](../assets/media/list-sort.svg)

`--format` and `--delimiter` offer some versatile formatting options to
customize the output.

![Note list format](../assets/media/list-format.svg)

`zk` is aware of the links you set between your notes. Backlinks or outbound
links of a note can be revealed by using the link filtering options. It even
supports listing indirect links thanks to `--recursive`.

![Note list links](../assets/media/list-link.svg)

`zk` supports an interactive mode powered by
[`fzf`](https://github.com/junegunn/fzf) to further filter notes manually.

![Note list interactive](../assets/media/list-interactive.svg)

## Edit existing notes

To edit notes with your default editor, use `zk edit`. It supports the same
[filtering options](../notes/note-filtering.md) as `zk list`.

```sh
$ zk edit --interactive --match "recipe pizza -pineapple"

# or with short flags
$ zk edit -i -m "recipe pizza -pineapple"
```

![Note edit](../assets/media/edit.svg)

## Edit the configuration file

To customize your experience with `zk`, you may want to edit the
[user configuration file](../config/config.md).

```sh
$ vim .zk/config.toml
```

Declaring your own [aliases](../config/config-alias.md) is a great way to make your
experience with `zk` easier and more familiar.

![Note alias](../assets/media/alias.svg)
