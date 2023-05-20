# Note configuration

The `[note]` section from the [configuration file](config.md) is used to set the [note creation rules](note-creation.md). The following properties are customizable:

* `language` (string)
    * Two-letters code of the language used when writing notes, e.g. `en`.
    * This is used to generate slugs or with date formats. For now, only English is fully supported.
* `default-title` (string)
    * The default title used for new notes when no `--title` option is provided.
* `filename` (string)
    * [Template](template.md) used to generate the note filename, without its file extension.
* `extension` (string)
    * File extension for the generated note. By default, `md` (Markdown) is used.
* `template` (string)
    * Path to the [template](template.md) used to generate the note content.
    * Either an absolute path, or relative to `.zk/templates/`.
* `exclude` (list of strings)
    * List of [path globs](https://en.wikipedia.org/wiki/Glob_\(programming\)) excluded during note indexing.
* `id-charset` (string)
    * Characters set used to [generate random IDs](note-id.md).
    * You can use:
        * `letters` for characters from `a` to `z`
        * `numbers` for characters from `0` to `9`
        * `alphanum` for `letters` + `numbers`
        * `hex` for characters from `a` to `f` and `0` to `9`
        * a free string for custom characters
* `id-length` (integer)
    * Length of the generated random IDs.
* `id-case` (enum)
    * Letter case for the generated random IDs.
    * Possible values are `lower`, `upper` or `mixed`.

## Common filename templates

Here are some common filename patterns you may want to use:

* `{{id}}` – e.g. `i2hn8.md`
    * Just a [random ID](note-id.md), simple and elegant.
    * To use [Neuron](neuron.md)'s ID format, set:
        ```toml
        [note]
        id-charset = "hex"
        id-length = 8
        id-case = "lower"
        ```
* `{{slug title}}` – e.g. `an-interesting-concept.md`
    * A [slugified](template.md) version of the title given with `--title`.
    * Readable and practical for web servers, but fragile in case of renaming.
* `{{id}}-{{slug title}}` – e.g. `i2hn8-an-interesting-concept.md`
    * The best of both worlds? Readable but if you link only with the prefix ID, you can rename without breaking links.
* `{{format-date now 'timestamp'}}` – e.g. `200911172034.md`
    * Verbose, but sortable by creation date and stable.
* `{{format-date now 'timestamp'}} {{title}}` – e.g. `200911172034 An interesting concept.md`
    * The format of [The Archive](https://zettelkasten.de/the-archive/) and [sirupsen's zk](https://github.com/sirupsen/zk).
* `{{format-date now '%Y-%m-%d'}}` – e.g. `2009-11-17.md`
    * Sortable, human-friendly format for a daily journal.
    * i.e. [Maintaining a daily journal](daily-journal.md).
