# Note formats

To keep your notebooks [future-proof](future-proof.md), `zk` uses a simple plain text format for your notes. Only Markdown is supported at the moment, but more formats may be added in the future.

You can set up some features of `zk`'s Markdown parser from your [configuration file](config.md), under the `[format.markdown]` section.

| Setting          | Default | Description                                                            |
|------------------|---------|------------------------------------------------------------------------|
| `hashtags `      | `true`  | Enable `#hashtags` support                                             |
| `colon-tags`     | `false` | Enable `:colon:separated:tags:` support                                |
| `multiword-tags` | `false` | Enable Bear's [`#multi-word tags#`][1]. Hashtags must also be enabled. |

[1]: https://blog.bear.app/2017/11/bear-tips-how-to-create-multi-word-tags/
