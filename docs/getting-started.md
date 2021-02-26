# Getting started with `zk`

A short introduction showing how to use `zk`.

## Create a new notebook

Create a [notebook](notebook.md) to host your notes. You are free to organize your notebook as you want, adding subdirectories if needed.

```sh
$ zk init my-notes
Initialized a notebook in my-notes

$ cd my-notes
```

## Edit the configuration file

To customize your experience with `zk`, you may want to edit the [user configuration file](config.md).

```sh
$ vim .zk/config.toml
```

## Create your first notes

Now you are ready to write your very first note. Pick a subject, [create a new note](note-creation.md) and write on!

```sh
$ zk new --title "An interesting concept"
```

## Edit existing notes

After some time, hopefully you will have enough notes to be lost in it. Use `zk`'s powerful [filtering capabilities](note-filtering.md) to find what you need.

```sh
$ zk edit --interactive --match "recipe pizza -pineapple"

# or with short flags
$ zk edit -i -m "recipe pizza -pineapple"
```

## List existing notes

If you do not need to edit a note, use `zk list` instead to print context-sensitive results.

```sh
$ zk list -m "recipe pizza -pineapple"
```
