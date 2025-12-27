# Extra user variables

`zk` is opened for template context extension which can be useful when
[creating new notes](../notes/note-creation.md), for example:

- expanding custom metadata (author, subject, etc.)
- modifying a [template](../notes/template.md)'s output dynamically depending on the
  value of an extra variable

## Static extra variables

You can declare static extra variables in the [configuration file](config.md)'s
`[extra]` section. Each [note group](config-group.md) can have its own `[extra]`
section, which may override values from the root section.

```toml
[extra]
visibility = "public"
author = "Mickaël"
authors = ["Thomas", "Aristotle"]

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

`--extra-json` can also be used to pass a JSON object instead of command-line
arguments. The following commands are the JSON equivalent to the above ones:

```sh
$ zk new --extra-json '{"author": "Thomas"}'
$ zk new --extra-json '{"show-header": "1", "author": "Thomas"}'
```

While the value passed to `--extra-json` must be a JSON object, its values need
not be:

```sh
$ zk new --extra-json '{"show-header": 1, "authors": ["Thomas", "Aristotle"]}'
```

## Using extra variables in templates

After declaring extra variables, you can expand them inside the
[template used when creating new notes](../notes/template-creation.md), using the usual
[Handlebars syntax](../notes/template.md).

```markdown
# {{title}}

Written by {{extra.author}}.

{{#if extra.show-header}} Behold, the mighty dynamic header! {{/if}}
{{#if extra.authors}}
    {{#each}}
        {{.}}
    {{/each}}
{{/if}}
```
