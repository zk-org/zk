# Notebook housekeeping

Tending to your notes does not only mean writing. You need to keep your [notebook](notebook.md) in great shape to make good use of it. For many maintenance tasks, `zk` can help!

## Find related notes

To surf your notebook with ease, make sure to link all related notes together. You can list notes which could be good candidates for a new link with the `--related` [filtering option](note-filtering.md).

```sh
$ zk list --related note.md
```

This returns notes which are not connected to the given note, but with at least one linked note in common.

## Find flimsy notes

To find flimsy notes needing to be fleshed out, you can list the first few notes with the smallest word count from your notebook with the following command:

```sh
$ zk list --format '{{word-count}}\t{{title}}' --sort word-count --limit 20
4       Integration with fzf
5       Searching and filtering notes
63      Setting your default editor
86      Anatomy of a notebook
...
```
