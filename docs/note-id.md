# Note ID

Each note is uniquely identified by its path relative to the [notebook](notebook.md)'s root. However, in some cases it is more convenient to refer to a "note ID", which is the unique part of its filename. For example, the note ID of the file `200911172034 An interesting concept.md` is `200911172034`. You could have several notes named "An interesting concept", but only one with the ID `200911172034`.

The purpose of using a unique identifier in your note filenames is to create stable links between your notes, which will not break even if you change the title of the linked note. [See this reference for more information](https://zettelkasten.de/introduction/#the-unique-identifier).

There are several flavors of note IDs and `zk` supports most of them. You can set it up in the [note configuration](config-note.md).

## Random ID

A random ID enables short and memorable unique identifiers. By default, `zk` is configured to generate random IDs of four alphanumeric characters. I found this to be the sweet spot between an easily memorable and usable ID and enough candidates. This default setting can generate 1 679 616 unique IDs.

## Timestamp

Another common ID is a timestamp in the `YYYYMMDDHHMM` shape. This is less readable than a short random ID, but has the added advantage of being sortable by creation date. However, I find this not so useful in practice.

## Sequential IDs

Sequential (incremented) IDs are currently not supported by `zk`. They get ugly very quickly when deleting outdated notes and have an irregular shape.

