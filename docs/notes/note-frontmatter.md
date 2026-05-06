# YAML frontmatter

Markdown being a simple format, it does not offer any way to attach additional
metadata to a note. The community came up with a solution by inserting a YAML
header at the top of each note to contain its metadata. This method is widely
supported among Zettelkasten softwares, including `zk`.

```yaml
---
title: Improve the structure of essays by rewriting
date: 2011-05-16 09:58:57
keywords: [writing, essay, practice]
---
```

`zk` supports the following metadata:

| Key        | Description                                                 |
| ---------- | ----------------------------------------------------------- |
| `title`    | Title of the note – takes precedence over the first heading |
| `date`     | Creation date – takes precedence over the file date         |
| `tags`     | List of tags attached to this note                          |
| `keywords` | Alias for `tags`                                            |
| `aliases`  | Alternative titles for this note, used by `--mention`       |

All metadata are indexed and can be printed in `zk list` output, using the
template variable `{{metadata.<key>}}`, e.g. `{{metadata.description}}`. The
keys are normalized to lower case.
