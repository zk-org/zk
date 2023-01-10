# Neuron

[Neuron](https://neuron.zettel.page/) is a command-line app for managing a plain-text [Zettelkasten](https://zettelkasten.de/introduction/).

While there is some overlap with `zk`'s features, both tools are actually useful when paired together:

* `zk` has powerful [filtering](note-filtering.md) and [note generation](note-creation.md) capabilities
* Neuron shines with its static website generation

Close integration with Neuron was thought through from the start when designing `zk`. For example, Neuron's [Folgezettel](https://neuron.zettel.page/folgezettel.html) syntax is supported: `[[[link]]]`, `#[[link]]` and `[[link]]#`.

<!-- TODO: They automatically add a `from` or `to` link relation when used. -->

But you can make your [notebook](notebook.md) even more tightly integrated with Neuron by:

* using the [same settings as Neuron](https://neuron.zettel.page/id.html) to generate the [note IDs](note-id.md) in the [note configuration](config-note.md)
    ```toml
    [note]
    filename = "{{id}}"
    id-charset = "hex"
    id-length = 8
    id-case = "lower"
    ```
* adding [command aliases](config-alias.md) for your frequently used `neuron` commands
    ```toml
    [alias]
    serve = "neuron gen -wS"
    gen = "neuron gen -o public"
    ```

## Emanote

Emanote is neuron's successor. For Emanote-specific configuration, see https://emanote.srid.ca/start/resources/zk.
