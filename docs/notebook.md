# Notebook

A *notebook* is a directory containing a collection of notes managed by `zk`. Notebooks cannot be nested, but you are free to organize your notes in subdirectories.

To create a new notebook, simply run `zk init [<directory>]`.

Most `zk` commands are operating "Git-style" on the notebook containing the current working directory (or one of its parents).

## Anatomy of a notebook

Similarly to Git, a notebook is identified by the presence of a `.zk` directory at its root. This directory contains the only `zk`-specific files in your notebook:

* `.zk/config.toml` is the user [configuration file](config.md)
* `.zk/templates/` contains [user templates](template.md) used when [creating new notes](note-creation.md)
* `.zk/notebook.db` is the SQLite database enabling [powerful search features](note-filtering.md).
