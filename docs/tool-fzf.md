# Integration with `fzf`

[`fzf`](https://github.com/junegunn/fzf) is an awesome and versatile fuzzy finder powering `zk`'s [interactive filtering mode](note-filtering.md).

Besides the standard [`fzf` configuration options](https://github.com/junegunn/fzf) documented on its website, `zk` offers additional options you can set in the `[tool]` [configuration section](config.md).

If you wish to customize more of `fzf` behavior, [please post a feature request](https://github.com/mickael-menu/zk/issues).

## Preview command

You can customize the command used to preview a note with `fzf-preview`. The special placeholder `{-1}` will be expanded to the note file path.

By default, `zk` uses `cat` for preview, which is a bit boring. A much better option would be to use [`bat`](https://github.com/sharkdp/bat) which supports syntax highlighting.

```toml
[tool]
fzf-preview = "bat -p --color always {-1}"
```

Or, if you prefer to preview more metadata, you can use a nested `zk` command.

```toml
[tool]
fzf-preview = "zk list --quiet --format full --limit 1 {-1}"
```
