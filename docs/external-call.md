# Call `zk` from other programs

Calling `zk` from other programs can be useful in a number of situations, such as:

* creating notes from your text editor using a custom shortcut
* creating a reference note from the text selected in your web browser
* automating periodical maintenance tasks on your [notebook](notebook.md)
* displaying the backlinks of a note in a GUI wrapper around `zk`

The following options can be useful to make sure `zk` behaves properly in a background context:

<!-- TODO: --color=none, --json -->
* `--no-input` disables all user prompts and ignores `--interactive`
* `--quiet` reduces unnecessary output

