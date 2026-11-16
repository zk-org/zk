# Setting your default editor

`zk` is not a text editor. Instead, it is designed to interface with your
favorite editor to write your notes.

You can customize which editor to use either from the
[configuration file](config.md) or environment variables. In order of
precedence, `zk` will use:

1. `ZK_EDITOR` environment variable
2. `editor` configuration property
   ```toml
   [tool]
   editor = "vim"
   ```
3. `VISUAL` environment variable
4. `EDITOR` environment variable

## Opening a note at a given line

When opening a [`zk://` URI](../notes/note-uri.md) with `zk open`, the cursor is
moved to the line and column the URI points at. The command-line syntax for this
differs per editor, so `zk` selects one from the name of the editor binary — see
[note URIs](../notes/note-uri.md) for the list of editors it knows about. An
unrecognized editor is given the plain file path instead, and the position is
dropped.
