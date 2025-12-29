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

## Date Keys

By default, `zk` tracks the creation date of notes via the `date` key (or the
file creation date if it is missing). It is possible to use a different key
name by specifying the following in the configuration:

```toml
[format.markdown.frontmatter]
creation-date-key = "created"
```

In addition to the creation date, `zk` can track the modification date of the
notes via a key in the frontmatter. To do this, add the following to your
configuration:

```toml
[format.markdown.frontmatter]
modification-date-key = "changed"
```

When a value for this key is present in the frontmatter, it is parsed as a date
and takes precedence over the file's modification date (for example when
running a command like `zk list --sort modified`). This is useful when the
modification date cannot be used, since it is modified by external programs.
For example when the notebook is stored in a git repository and the
modification date is not tracked by git. Or when the synchronization mechanism
does not update file attributes.

Your notes then can use these custom keys:

```markdown
---
title: higher complexity does not always result in more order
created: 2025-12-28 09:22
changed: 2025-12-29 10:28
id: u6nh
tags: []
aliases:
---
```

This feature is best combined with external tools changing the frontmatter on
modification, see [Modification Dates](../tips/modification-date.md).
