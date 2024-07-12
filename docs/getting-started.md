# Getting started with `zk`

A short introduction showing how to use `zk`.

## Create a new notebook

Create a [notebook](notebook.md) to host your notes. You are free to organize your notebook as you want, adding subdirectories if needed.

```sh
$ zk init my-notes
Initialized a notebook in my-notes

$ cd my-notes
```

## Create your first notes

Now you are ready to write your very first note. Pick a subject, [create a new note](note-creation.md) and write on!

```sh
$ zk new --title "An interesting concept"
```

You can customize your experience using [custom templates](template.md) to generate many kind of notes.

<div align="center"><img alt="Create a note" width="85%" src="assets/media/new1.svg"/></div>

If you are not sure whether a note already exists for a particular subject, the "search or create" mode might be more appropriate than `zk new`. 
It is inspired by [Notational Velocity](https://notational.net/) and enables searching for an existing note or creating a new one in a single action.

From `zk`'s interactive edit screen, press `Ctrl-E` to create a new note using the current search query as title.

<div align="center"><img alt="Create a note" width="85%" src="assets/media/new2.svg"/></div>

## List existing notes

After some time, hopefully you will have enough notes to be lost in it. 

To help structure your notebook, you can add [metadata](note-frontmatter.md) (e.g. keywords/tags) to your notes. 
You can then use `zk`'s powerful [filtering capabilities](note-filtering.md) to find the notes you need.

```sh
$ zk list --tag "recipe" --match "pizza -pineapple"
```
<div align="center"><img alt="Format the list output" width="85%" src="assets/media/list.svg"/></div>

Sort the results however you need with `--sort`.

<div align="center"><img alt="Format the list output" width="85%" src="assets/media/list-sort.svg"/></div>

`--format` and `--delimiter` offer some versatile formatting options to customize the output.

<div align="center"><img alt="Format the list output" width="85%" src="assets/media/list-format.svg"/></div>

`zk` is aware of the links you set between your notes. 
Backlinks or outbound links of a note can be revealed by using the link filtering options. 
It even supports listing indirect links thanks to `--recursive`.

<div align="center"><img alt="Format the list output" width="85%" src="assets/media/list-link.svg"/></div>

`zk` supports an interactive mode powered by [`fzf`](https://github.com/junegunn/fzf) to further filter notes manually.

<div align="center"><img alt="Format the list output" width="85%" src="assets/media/list-interactive.svg"/></div>

## Edit existing notes

To edit notes with your default editor, use `zk edit`. It supports the same [filtering options](note-filtering.md) as `zk list`.

```sh
$ zk edit --interactive --match "recipe pizza -pineapple"

# or with short flags
$ zk edit -i -m "recipe pizza -pineapple"
```

<div align="center"><img alt="Format the list output" width="85%" src="assets/media/edit.svg"/></div>

## Edit the configuration file

To customize your experience with `zk`, you may want to edit the [user configuration file](config.md).

```sh
$ vim .zk/config.toml
```

Declaring your own [aliases](config-alias.md) is a great way to make your experience with `zk` easier and more familiar.

<div align="center"><img alt="Format the list output" width="85%" src="assets/media/alias.svg"/></div>

