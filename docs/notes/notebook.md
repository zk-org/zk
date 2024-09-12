# Notebook

A _notebook_ is a directory containing a collection of notes managed by `zk`.
Notebooks cannot be nested, but you are free to organize your notes in
subdirectories.

To create a new notebook, simply run `zk init [<directory>]`.

Most `zk` commands are operating "Git-style" on the notebook containing the
current working directory (or one of its parents). However, you can explicitly
set which notebook to use with `--notebook-dir` or the `ZK_NOTEBOOK_DIR`
environment variable. Setting `ZK_NOTEBOOK_DIR` in your shell configuration
(e.g. `~/.profile`) can be used to define a default notebook which `zk` commands
will use when the working directory is not in another notebook.

If the [default notebook](../config/config-notebook.md) is set it will be used as
`ZK_NOTEBOOK_DIR`, unless this environment variable is not already set.

## Anatomy of a notebook

Similarly to Git, a notebook is identified by the presence of a `.zk` directory
at its root. This directory contains the only `zk`-specific files in your
notebook:

- `.zk/config.toml` is the user [configuration file](../config/config.md)
- `.zk/templates/` contains [user templates](template.md) used when
  [creating new notes](note-creation.md)
- `.zk/notebook.db` is the SQLite database enabling
  [powerful search features](note-filtering.md).
