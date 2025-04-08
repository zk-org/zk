# Static Site Solutions

## Emanote

[Emanote](https://github.com/srid/emanote) is Neuron's successor.

For an Emanote-specific zk configuration, see
[https://emanote.srid.ca/start/resources/zk](https://emanote.srid.ca/start/resources/zk).

## Golumn

[Golumn](https://github.com/gollum/gollum) is _"a simple, Git-powered wiki with
a local frontend and support for many kinds of markup and content."_

## Neuron (stale)

Note: At the time of writing (2025-03-25) Neuron is seemingly no longer actively
developed.

[Neuron](https://neuron.zettel.page/) is a command-line app for managing a
plain-text [Zettelkasten](https://zettelkasten.de/introduction/).

While there is some overlap with `zk`'s features, both tools are actually useful
when paired together:

- `zk` has powerful [filtering](../notes/note-filtering.md) and
  [note generation](../notes/note-creation.md) capabilities
- Neuron shines with its static website generation

Close integration with Neuron was thought through from the start when designing
`zk`. For example, Neuron's
[Folgezettel](https://neuron.zettel.page/folgezettel.html) syntax is supported:
`[[[link]]]`, `#[[link]]` and `[[link]]#`.

<!-- TODO: They automatically add a `from` or `to` link relation when used. -->

But you can make your [notebook](../notes/notebook.md) even more tightly
integrated with Neuron by:

- using the [same settings as Neuron](https://neuron.zettel.page/id.html) to
  generate the [note IDs](../notes/note-id.md) in the
  [note configuration](../config/config-note.md)
  ```toml
  [note]
  filename = "{{id}}"
  id-charset = "hex"
  id-length = 8
  id-case = "lower"
  ```
- adding [command aliases](../config/config-alias.md) for your frequently used
  `neuron` commands
  ```toml
  [alias]
  serve = "neuron gen -wS"
  gen = "neuron gen -o public"
  ```
