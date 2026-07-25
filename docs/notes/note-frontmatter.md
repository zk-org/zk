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
| `modified` | Modification date – takes precedence over the file date     |
| `tags`     | List of tags attached to this note                          |
| `keywords` | Alias for `tags`                                            |
| `aliases`  | Alternative titles for this note, used by `--mention`       |

All metadata are indexed and can be printed in `zk list` output, using the
template variable `{{metadata.<key>}}`, e.g. `{{metadata.description}}`. The
keys are normalized to lower case.

## Date Keys

By default, `zk` tracks the creation date of notes via the `date` key. Likewise,
the modification date can be set and tracked. If the `date` or `modified` keys
have no value set, the file system creation and modification values are used.

The keys names themselves can be customised.

```toml
[format.markdown.frontmatter]
creation-date-key = "created" # default is "date"
modification-date-key = "changed" # default is "modified"
```

Tracking the creation and modification dates is useful as the file's own dates
are not always the dates that you actually created or last edited the file. This
can happen if you use git to clone your notebook from an online repository for
example.

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

This feature is best combined with external tools to update the modified dates
programmatically, see [Modification Dates](../tips/modification-date.md).

To automate the creation date simply include in your template:

```
---
date: {{format-date now}} {{format-date now "time"}}
---
```

For further information on creating notes with templates, see
[here](template-creation.md).
