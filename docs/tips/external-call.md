# Call `zk` from other programs

Calling `zk` from other programs can be useful in a number of situations, such as:

* creating notes from your text editor using a custom shortcut
* creating a reference note from the text selected in your web browser
* automating periodical maintenance tasks on your [notebook](../notes/notebook.md)
* displaying the backlinks of a note in a GUI wrapper around `zk`

The following options can be useful to make sure `zk` behaves properly in a background context:

<!-- TODO: --color=none, --json -->
* `--no-input` disables all user prompts and ignores `--interactive`
* `--quiet` reduces unnecessary output


## Linking into `zk` from other programs

To go the other way — from another program back into a note — use a
[`zk://` URI](../notes/note-uri.md). `zk uri` prints one for a given note, line
and column, and `zk open` opens it in your editor. `zk` can be registered as
the system handler for the scheme, so that such links become clickable
anywhere.
