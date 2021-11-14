# Changelog

All notable changes to this project will be documented in this file.

## Unreleased

### Added

* New `zk graph --format json` command which produces a JSON graph of the notes matching the given criteria.
* New template variables `filename` and `filename-stem` when formatting notes (e.g. with `zk list --format`) and for the [`fzf-line`](docs/tool-fzf.md) config key.
* Customize how LSP completion items appear in your editor when auto-completing links with the [`[lsp.completion]` configuration section](docs/config-lsp.md).
    ```toml
    [lsp.completion]
    # Show the note title in the completion pop-up, or fallback on its path if empty.
    note-label = "{{title-or-path}}"
    # Filter out the completion pop-up using the note title or its path.
    note-filter-text = "{{title}} {{path}}"
    # Show the note filename without extension as detail.
    note-detail = "{{filename-stem}}"
    ```
* New `--dry-run` flag for `zk new` which prints out the path and content of the generated note instead of saving it to the file system.
* New `--verbose` flag for `zk index` which prints detailed information about the indexing process.

### Fixed

* [#89](https://github.com/mickael-menu/zk/issues/89) Calling `zk index` from outside the notebook (contributed by [@adamreese](https://github.com/mickael-menu/zk/pull/90)).
* [#98](https://github.com/mickael-menu/zk/issues/98) Index wiki links using partial paths for `--linked-by` and `--link-to`.
* [#98](https://github.com/mickael-menu/zk/issues/98) Ignore spaces around the pipe in wiki links for LSP diagnostics.


## 0.7.0

### Added

* List the tags found in your notebook with `zk tag list`.
    * Many options are available to customize the output, including JSON serialization. See `zk tag list --help`.
* Support for LSP references to browse the backlinks of the current note, if the caret is not over a link.
* New template variables are available when [generating custom Markdown links with `link-format`](docs/note-format.md).
    * `filename`, `path`, `abs-path` and `rel-path` for many path flavors.
    * `metadata` to use information (e.g. `id`) from the YAML frontmatter.
* The LSP server is now matching wiki links to any part of a note's path or its title.
    * Given the note `book/z5mj Information Graphics.md` with the title "Book Review of Information Graphics", the following wiki links would work from a note located under `journal/2020-09-25.md`:
        ```markdown
        [[../book/z5mj]]
        [[book/z5mj]]
        [[z5mj]]
        [[book review information]]
        [[Information Graphics]]
        ```
* Use the `{{abs-path}}` template variable when [formatting notes](docs/template-format.md) to print the absolute path to the note (contributed by [@pstuifzand](https://github.com/mickael-menu/zk/pull/60)).
* A new `{{substring s index length}}` template helper extracts a portion of a given string, e.g.:
    * `{{substring 'A full quote' 2 4}}` outputs `full`
    * `{{substring 'A full quote' -5 5}` outputs `quote`

### Fixed

* UTF-8 handling in the LSP server.
* [#78](https://github.com/mickael-menu/zk/issues/78) Do not exclude notes containing broken links from the index.
* Allow setting the `--working-dir` and `--notebook-dir` flags before the `zk` subcommand when using aliases, e.g. `zk -W ~/notes my-alias`.
* [#86](https://github.com/mickael-menu/zk/issues/86) Index encoded Markdown links.


## 0.6.0

### Added

* Use JSON formats with `zk list` for easy post-processing:
    * `--format json` prints a plain JSON array.
    * `--format jsonl` prints one JSON note object per line, according to [JSON Lines](https://jsonlines.org/).
* The new `{{json}}` template helper serializes any template context variable into a valid JSON value, e.g.:
    * `{{json title}}` prints with quotes `"An interesting note"`
    * `{{json .}}` serializes the full template context as a JSON object.
* Use `--header` and `--footer` options with `zk list` to print arbitrary text at the start or end of the list.
* Support for LSP references to browse the backlinks of the link under the caret (contributed by [@pstuifzand](https://github.com/mickael-menu/zk/pull/58)).
* New [`note.ignore`](docs/config-note.md) configuration option to ignore files matching the given path globs when indexing notes.
    ```yaml
    [note]
    ignore = [
        "log-*.md"
        "drafts/*"
    ]
    ```

### Fixed

* [#16](https://github.com/mickael-menu/zk/issues/16) Links with section anchors, e.g. `[[filename#section]]`.
* Unicode support in wiki links. If you use accents or ideograms, please run `zk index --force` after upgrading to fix your index.


## 0.5.0

### Added

* [Editor integration through LSP](https://github.com/mickael-menu/zk/issues/22):
    * New code actions to create a note using the current selection as title.
    * Custom commands to [run `new` and `index` from your editor](docs/editors-integration.md#custom-commands).
    * Diagnostics to [report dead links or wiki link titles](docs/config-lsp.md).
    * Auto-complete only the path of a Markdown link by typing `[custom title]((`.
* Customize the format of `fzf`'s lines [with your own template](docs/tool-fzf.md).
    ```toml
    [tool]
    fzf-line = "{{style 'green' path}}{{#each tags}} #{{this}}{{/each}} {{style 'black' body}}"
    ```

### Changed

* Automatically index the notebook when saving a note with an LSP-enabled editor.
    * This ensures that tags and notes auto-completion lists are up-to-date.

### Fixed

* Creating a new note from `fzf` in a directory containing spaces.
* Fix completion with Neovim's built-in LSP client (contributed by [@cormacrelf](https://github.com/mickael-menu/zk/pull/39)).


## 0.4.0

### Added

* Interactive wizard for the `zk init` command.
* An experimental Language Server for LSP-compatible editors:
    * Auto-complete Markdown links with `[[` (setup wiki links in the [note formats configuration](docs/note-format.md))
    * Auto-complete [hashtags and colon-separated tags](docs/tags.md).
    * Preview the content of a note when hovering a link.
    * Navigate in your notes by following internal links.
    * [And more to come...](https://github.com/mickael-menu/zk/issues/22)
    * See [the documentation](docs/editors-integration.md) for configuration samples.
* Pair `--match` with `--exact-match` / `-e` to search for (case insensitive) exact occurrences in your notes.
    * This can be useful when looking for terms including special characters, such as `[[name]]`.
* Generating links to notes.
    * Use the `{{link}}` template variable when [formatting notes](docs/template-format.md) to print a link to the note, relative to the working directory.
    * Use the `{{format-link path title}}` template helper to render a custom link.
    * Customize the link format from the [note formats settings](docs/note-format.md). You can for example choose regular Markdown links, wiki links or a custom format.

### Changed

* The local configuration file (`.zk/config.toml`) is not required anymore in a notebook's `.zk` directory.
* `--notebook-dir` does not change the working directory anymore, instead it sets manually the current notebook and disable auto-discovery. Use the new `--working-dir`/`-W` flag to run `zk` as if it was started from this path instead of the current working directory.
    * For convenience, `ZK_NOTEBOOK_DIR` behaves like setting a `--working-dir` fallback, instead of `--notebook-dir`. This way, paths will be relative to the root of the notebook.
    * A practical use case is to use `zk list -W .` when outside a notebook. This will list the notes in `ZK_NOTEBOOK_DIR` but print paths relative to the current directory, making them actionable from your terminal emulator.


## 0.3.0

### Added

* Global `zk` configuration at `~/.config/zk/config.toml`.
    * Useful to share aliases or default settings across several [notebooks](docs/notebook.md).
    * This is the same format as a notebook [configuration file](docs/config.md).
    * Shared templates can be stored in `~/.config/zk/templates/`.
    * `XDG_CONFIG_HOME` is taken into account.
* Use `--notebook-dir` or set `ZK_NOTEBOOK_DIR` to run `zk` as if it was started from this path instead of the current working directory.
    * This allows running `zk` without being in a notebook.
    * By setting `ZK_NOTEBOOK_DIR` in your shell configuration file (e.g. `~/.profile`), you are declaring a default global notebook which will be used when `zk` is not in a notebook.
    * When the notebook directory is set explicitly, any path given as argument will be relative to it instead of the actual working directory.
* Find every note whose title is mentioned in the note you are working on with `--mentioned-by file.md`.
    * To refer to a note using several names, you can use the [YAML frontmatter key `aliases`](https://publish.obsidian.md/help/How+to/Add+aliases+to+note). For example the note titled "Artificial Intelligence" might have: `aliases: [AI, robot]`
    * To find only unlinked mentions, pair it with `--no-linked-by`, e.g. `--mentioned-by file.md --no-linked-by file.md`.
* Declare [named filters](docs/config-filter.md) in the configuration file to reuse [note filtering options](docs/note-filtering.md) used frequently together, for example:
    ```toml
    [filter]
    recents = "--sort created- --created-after 'last two weeks'"
    ```
    ```sh
    $ zk list recents --limit 10
    $ zk edit recents --interactive
    ```

### Fixed

* [#4](https://github.com/mickael-menu/zk/issues/4) Terminal borked when piping content with Vim


## 0.2.1

### Fixed

* Looking for mentions of a note with a title containing double quotes.
* Crash when parsing certain link snippets.


## 0.2.0

### Added

* Support for tags.
    * Filter notes by their tags using `--tag "history, europe"`.
        * To match notes associated with either tags, use a pipe `|` or `OR` (all caps), e.g. `--tag "inbox OR todo"`.
        * If you want to exclude notes having a particular tag, prefix it with `-` or `NOT` (all caps), e.g. `--tag "NOT done"`.
        * Use glob patterns to match multiple tags, e.g. `--tag "book-*"`.
    * Many tag flavors are supported: `#hashtags`, `:colon:separated:tags:` ([opt-in](docs/note-format.md)) and even Bear's [`#multi-word tags#`](https://blog.bear.app/2017/11/bear-tips-how-to-create-multi-word-tags/) ([opt-in](docs/note-format.md)). If you prefer to use a YAML frontmatter, list your tags with the key `tags` or `keywords`.
* Find every mention of a note in your notebook with `--mention file.md`.
    * This will look for occurrences of the note's title in other notes.
    * To refer to a note using several names, you can use the [YAML frontmatter key `aliases`](https://publish.obsidian.md/help/How+to/Add+aliases+to+note). For example the note titled "Artificial Intelligence" might have: `aliases: [AI, robot]`
    * To find only unlinked mentions, pair it with `--no-link-to`, e.g. `--mention file.md --no-link-to file.md`.
* Print metadata from the [YAML frontmatter](docs/note-frontmatter.md) in `list` output using `{{metadata.<key>}}`, e.g. `{{metadata.description}}`. Keys are normalized to lower case.
* Use the YAML frontmatter key `date` for the note creation date, when provided.
* Access environment variables from note templates with the `env.<key>` template variable, e.g. `{{env.PATH}}`.

### Changed

* Renamed `--linking-to` filtering option to `--link-to`.
* Multiple `--extra` variables are now separated by `,` instead of `;`.
