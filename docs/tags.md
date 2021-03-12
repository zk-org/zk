# Tags

Tags are a useful way to organize and filter your notes with `zk`, which supports most syntaxes:

* `#hashtags`
* `:colon:separated:tags:` ([opt-in](note-format.md))
* Bear's `#multi-word tags#` ([opt-in](note-format.md))
* YAML frontmatter (`tags` and `keywords` keys).

You can filter your notes by their tags using the `--tags` option, as demonstrated in [Searching and filtering notes](note-filtering.md).

```sh
$ zk list --tag "inbox OR todo, NOT done"
```
