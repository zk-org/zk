# Searching and filtering notes

A few commands are built upon `zk`'s powerful note filtering capabilities, such
as `edit` and `list`. They accept any option described here. You may also
declare [named filters](../config/config-filter.md) in the
[configuration file](../config/config.md) for the same set of options you use
frequently.

## Filter by path

All filtering commands take for unique positional argument a list of paths. When
set, only the notes matching the given paths will be returned.

You can use it to find all the notes in a directory.

```sh
$ zk list journal/daily journal/weekly
```

Or specific notes.

```sh
$ zk edit 200911172034-an-interesting-concept.md
```

It works fine with only a path prefix as well. This is useful when you have a
[note ID](note-id.md) prefix, but not the full file path.

```sh
$ zk edit 200911172034
```

These rules apply to all the following options, when they expect a `<path>`
parameter.

```sh
$ zk list --link-to 200911172034
```

You can also use a nested `zk` command to pre-filter paths to feed to an option
with a `<path>` argument.
[See the `inline` command alias example](../config/config-alias.md) for more
explanation.

```sh
# List the notes which have at least one link pointing to them (i.e. not orphans).
$ zk list --exclude "`zk inline --orphan`"

# List the notes which are linked by at least one note from the journal/ directory.
$ zk list --linked-by "`zk inline journal`"
```

## Search the title or body

Use `--match <query>` (or `-m`) to search through the title and body of notes.

The search is powered by different strategies to answer various use cases:

- `fts` (default) uses a
  [full-text search](https://en.wikipedia.org/wiki/Full-text_search) database to
  offer near-instant results and advanced search operators.
- `exact` is useful if you need to find patterns containing special characters.
- `re` enables regular expression for advanced use cases.

Change the currently used strategy with `--match-strategy <strategy>` (or `-M`).
To set the default strategy, you can declare a [custom alias](../config/config-alias.md):

```toml
[alias]
list = "zk list --match-strategy re $@"
```

The `--match` option may be given multiple times, where each argument will be
combined with a boolean AND.

For example,

```sh
$ zk list --tag "recipe" --match "pizza -pineapple" --match "mushrooms"
```

Is equivalent to,

```sh
$ zk list --tag "recipe" --match "(pizza -pineapple) AND (mushrooms)"
```

### Full-text search (`fts`)

The default match strategy is powered by a
[full-text search](https://en.wikipedia.org/wiki/Full-text_search) database
enabling near-instant results. Queries are not case-sensitive and terms are
tokenized, which means that searching for `create` will also match `created` and
`creating`.

A syntax similar to Google Search is available for advanced search queries.

```sh
# FTS is the default match strategy
$ zk list --match "tesla OR edison"

# ...but you can enable it explicitly.
$ zk list --match-strategy fts --match "tesla OR edison"
$ zk list -Mf -m "tesla OR edison"
```

#### Combining terms

By default, the search engine will find the notes containing all the terms in
the query, in any order.

```
"tesla edison"
```

If you want to find the notes containing any or both of the terms, put `OR` (all
caps) or a pipe `|` between them.

```
"tesla OR edison"
"tesla | edison"
```

Search for an exact phrase by surrounding it with double quotes. In this case,
you will need to single quote the full query if you do not want to escape the
double quotes.

```
'tesla "alternating current"'
```

To construct more complex queries, you can group sub-queries with parentheses.

```
"current (tesla OR edison)"
```

Finally, you can filter out results by excluding a term with `NOT` (all caps) or
a `-` prefix.

```
"tesla NOT car"
"tesla -car"
```

#### Search in specific fields

If you want to search only in the title or body of notes, prefix a query with
`title:` or `body:`.

```
"title: tesla"
"body: (tesla OR edison)"
```

#### Prefix terms

Match any term beginning with the given prefix with a wildcard `*`.

```
"edi*"
```

Prefixing a query with `^` will match notes whose title or body start with the
following term.

```
"title: ^journal"
```

### Exact matches (`exact`)

If you need to find patterns containing special characters, such as an
`email@addre.ss` or a `[[wiki-link]]`, use the `exact` match strategy. The
search will be case-insensitive.

```sh
$ zk list --match-strategy exact --match "[[link]]"
$ zk list -Me -m "[[link]]"
```

### Regular expressions (`re`)

For advanced use cases, you can use the `re` match strategy to search the
notebook using regular expressions. The supported syntax is similar to the one
used by Python or Perl.
[See the full reference](https://golang.org/s/re2syntax).

:warning: Make sure to use quotes to prevent your shell from expanding
wildcards.

```sh
# Find notes containing emails.
$ zk list --match-strategy re --match ".+@.+"
$ zk list -Mr -m ".+@.+"
```

## Filter by tags

You can filter your notes by their [tags](tags.md) using `--tags` (or `-t`).

Find the notes having several tags by separating them with a comma.

```sh
$ zk list --tag "history, europe"
```

To match notes having either or both tags, use a pipe `|` or `OR` (all caps).

```sh
$ zk list --tag "inbox OR todo"
```

If you want to exclude notes having a particular tag instead, prefix it with `-`
or `NOT` (all caps).

```sh
$ zk list --tag "NOT done"
```

Your shell might give you some trouble using the `-` prefix. You can quote it
and add an extra space as a workaround, e.g. `--tag " -done"`.

You can use glob patterns to match multiple tags. This is particularly useful if
you use a separator (e.g. `/`) to group multiple tags under a parent tag.

```sh
$ zk list --tag "year/201*"
```

A useful [notebook housekeeping](../tips/notebook-housekeeping.md) feature is to find
tags which _do not_ have tags.

```sh
$ zk list --tagless
```

## Filter by creation or modification date

To find notes created or modified on a specific day, use `--created <date>` and
`--modified <date>`. They accept a human-friendly date for argument.

```
--created yesterday
--created "last tuesday"
--modified "Feb 3"
```

You can filter by range instead, using `--created-before`, `--created-after`,
`--modified-before` and `--modified-after`.

```
--created-before 10am
--modified-after 2021
--created-after "last monday" --created-before yesterday
```

## Explore links

You can use the following options to explore the web of links spanning your
[notebook](notebook.md).

`--linked-by <path>` (or `-L`) finds the notes linked by the given one, while
`--link-to <path>` (or `-l`) searches the notes having a link to it (also known
as _backlinks_).

```
--linked-by 200911172034
--link-to 200911172034
```

These options stop at the first level by default. But you can explore the whole
web by adding the `--recursive` (or `-r`) option to find all the notes leading
to (or from) a given note. If you feel overwhelmed, limit the distance between
two notes with `--max-distance <count>`.

```
--linked-by 200911172034 --recursive --max-distance 3
```

Finally, it can be useful to see which notes have no links pointing to them at
all. You can use the `--orphan` option for this.

## Find related notes

Part of writing a great notebook is to establish links between related notes.
The `--related <path>` option can help by listing results having a linked note
in common, but not yet connected to the note.

```
--related 200911172034
```

## Locate mentions of other notes

Another great way to look for potential new links is to find every mention of
other notes in the note you are currently working on.

```
--mentioned-by 200911172034
```

This option will find every note whose title is mentioned in the given note. To
refer to a note using several names, you can use the
[YAML frontmatter](note-frontmatter.md) to declare additional aliases. For
example, a note titled "Artificial Intelligence" might have for aliases "AI" and
"robot". This method is compatible with
[Obsidian](https://publish.obsidian.md/help/How+to/Add+aliases+to+note).

```
---
title: Artificial Intelligence
aliases: [AI, robot]
---
```

Alternatively, find every note mentioning the given note with `--mention`.

```
--mention 200911172034
```

To find only unlinked mentions, pair the `--mentioned-by` and `--mentions`
options with `--no-linked-by` (resp. `--no-link-to`) to remove notes which are
already linked from the results.

```
--mentioned-by 200911172034 --no-linked-by 200911172034
--mention 200911172034 --no-link-to 200911172034
```

## Exclude notes from the results

To prevent certain notes from polluting the results, you can explicitly exclude
them with `--exclude <path>` (or `-x`). This is particularly useful when you
have a whole directory of notes to be ignored.

```
-x journal
```

## Limit the number of results

If you are only interested into the first few notes, limit the number of results
with `--limit <count>` (or `-n`).

```
--limit 20
```

Using `-n1` is particularly common when you are expecting only a single result.

## Interactive filtering

A common search flow is to reduce the search scope using `zk`'s filtering
options, before selecting manually the notes to process among them. This is
especially useful with `zk edit` to avoid opening many unwanted notes with your
editor.

Use `--interactive` (or `-i`) to select filtered notes manually. The interactive
selection is handled by [`fzf`](../config/tool-fzf.md) which brings a powerful fuzzy
matching search into the mix.

## Sort the results

After finding matching notes, it might be useful to sort them before processing.
The `--sort <criteria>` (or `-s`) option is made for that.

You can add a `+` (ascending) or `-` (descending) suffix to a sort criterion to
customize the order. Each criterion has a sensible intrinsic order by default.

```
--sort path
--sort created+
-st- (eq. --sort title-)
```

| Criterion    | Shortcut | Order | Description                        |
| ------------ | -------- | ----- | ---------------------------------- |
| `created`    | `c`      | `-`   | Creation date                      |
| `modified`   | `m`      | `-`   | Modification date                  |
| `path`       | `p`      | `+`   | File path relative to the notebook |
| `title`      | `t`      | `+`   | Note title                         |
| `random`     | `r`      | `+`   | Order notes randomly               |
| `word-count` | `wc`     | `+`   | Word count in the note             |
