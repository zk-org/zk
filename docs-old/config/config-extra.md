# Extra user variables

`zk` is opened for template context extension which can be useful when
[creating new notes](../notes/note-creation.md), for example:

- expanding custom metadata (author, subject, etc.)
- modifying a [template](../notes/template.md)'s output dynamically depending on
  the value of an extra variable

## Static extra variables

You can declare static extra variables in the [configuration file](config.md)'s
`[extra]` section. Each [note group](config-group.md) can have its own `[extra]`
section, which may override values from the root section.

```toml
[extra]
visibility = "public"
author = "Mickaël"

[group.journal.extra]
visibility = "private" # overrides
```

## Dynamic extra variables

Maybe more useful, you can provide additional extra variables dynamically to
`zk new` from the command-line with `--extra`. Multiple variables can be
separated by a comma `,`.

```sh
$ zk new --extra author=Thomas
$ zk new --extra show-header=1,author=Thomas
```

## Using extra variables in templates

After declaring extra variables, you can expand them inside the
[template used when creating new notes](../notes/template-creation.md), using
the usual [Handlebars syntax](../notes/template.md).

```markdown
# {{title}}

Written by {{extra.author}}.

{{#if extra.show-header}} Behold, the mighty dynamic header! {{/if}}
```

## Listing extras

You can list all the extras found in your configuration file using
`zk config --list extras`.
