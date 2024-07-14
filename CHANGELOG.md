# Changelog

All notable changes to this project will be documented in this file.

## Unreleased

## Added

* Path in .zk/config.toml for the default note template now accepts UNIX "~/paths" (by @WhyNotHugo)

## Fixed

* LSP ignores magnet links as links to notes (by @billymosis)
* Compilation robustness for Alpine package builds (by @nmeum)

## 0.14.1

### Fixed

* Fixed parsing large notes @khimaros in https://github.com/zk-org/zk/pull/339
* fix day range parsing (zk-org/zk#382) by @tjex in https://github.com/zk-org/zk/pull/384
* accept tripple dash file URIs as valid links by @tjex in https://github.com/zk-org/zk/pull/391
* fix(lsp): fix trigger completion of zk LSP by @Rahlir in https://github.com/zk-org/zk/pull/397
* fix(lsp): ignore diagnostic check within code blocks by @Rahlir in https://github.com/zk-org/zk/pull/399
* allow notebook as hidden dir by @tjex in https://github.com/zk-org/zk/pull/402

## 0.14.0

### Added

* New [`tool.shell`](docs/tool-shell.md) configuration key to set a custom shell (contributed by [@lsvmello](https://github.com/zk-org/zk/pull/302)).
* New [`notebook.dir`](docs/config-notebook.md) configuration key to set the default notebook (contributed by [@lsvmello](https://github.com/zk-org/zk/pull/304)).

### Changed

* The `note.ignore` configuration property was renamed to `note.exclude`, to be more consistent with the CLI flags.

### Fixed

* Fixed LSP positions using UTF-16 offsets (contributed by [@wrvsrx](https://github.com/zk-org/zk/pull/317)).

## 0.13.0

### Added

* LSP:
    * `zk.new` now returns the created note's content in its output (`content`), and has two new options:
        * `dryRun` will prevent `zk.new` from creating the note on the file system.
        * `insertContentAtLocation` can be used to insert the created note's content into an arbitrary location.
    * A new `zk.link` command to insert a link to a given note (contributed by [@psanker](https://github.com/zk-org/zk/pull/284)).

## 0.12.0

### Added

* LSP: Support for external URLs with `documentLink`.
* New `{{date}}` template helper to obtain a date object from natural language (contributed by [@zalegrala](https://github.com/zk-org/zk/pull/262)).
    ```
    Get a relative date using natural language:
    {{date "next week"}}

    Format a date returned by `get-date`:
    {{format-date (date "monday") "timestamp"}}
    ```
* `zk list` now support multiple `--match`/`-m` flags, which allows to search for several tokens appearing in any order in the notes (contributed by [@rktjmp](https://github.com/zk-org/zk/pull/268)).

### Changed

* **Breaking change:** The `{{date}}` template helper was renamed to `{{format-date}}`. You might need to update your configuration and templates.

### Fixed

* [#243](https://github.com/zk-org/zk/issues/243) LSP: Fixed finding backlink references for notes in a folder.
* [#254](https://github.com/zk-org/zk/issues/254) Fixed SQL error when pairing `--link-to` and `--linked-by`.


## 0.11.1

### Changed

* `zk new` now requires the `--interactive`/`-i` flag to read the note body from a pipe or standard input. [See rational](https://github.com/zk-org/zk/pull/242#issuecomment-1182602001).

### Fixed

* [#244](https://github.com/zk-org/zk/issues/244) Fixed `zk new` waiting for `Ctrl-D` to proceed (contributed by [@pkazmier](https://github.com/zk-org/zk/pull/242)).


## 0.11.0

### Added

* Use regular expressions when searching for notes with `--match`.
    ```sh
    # Find notes containing emails.
    $ zk list --match-strategy re --match ".+@.+"
    $ zk list -Mr -m ".+@.+"
    ```

### Changed

* The flags `--exact-match`/`-e` are deprecated in favor of `--match-strategy exact`/`-Me`.

### Deprecated

* The LSP server does not support resolving a wiki link to a note title anymore.
    * For example, `[[Planet]]` can match a note with filename `i4w0 Planet.md` but not `i4w0.md` with a Markdown title `Planet` anymore.
    * This "smart" fallback resolution based on note titles was too fragile and not supported by the `zk` CLI.

### Fixed

* [#233](https://github.com/zk-org/zk/issues/233) Hide index progress in non-interactive shells.
* [#235](https://github.com/zk-org/zk/issues/235) Fix LSP link recognition with unicode (contributed by [@zkbpkp](https://github.com/zk-org/zk/issues/235)).
* [#236](https://github.com/zk-org/zk/issues/236) Fix updating links after creating a new note.
* [#239](https://github.com/zk-org/zk/discussions/239) Support standard input via shell redirection with `zk new`.


## 0.10.1

### Changed

* Removed the dependency on `libicu`.

### Fixed

* Indexed links are now automatically updated when adding a new note, if it is a better match than the previous link target.


## 0.10.0

### Added

* New `--date` flag for `zk new` to set the current date manually.
* New `--id` flag for `zk new` to skip ID generation and use a provided value (contributed by [@skbolton](https://github.com/zk-org/zk/pull/183)).
* [#144](https://github.com/zk-org/zk/issues/144) LSP auto-completion of YAML frontmatter tags.
* [zk-nvim#26](https://github.com/zk-org/zk-nvim/issues/26) The LSP server doesn't use `additionalTextEdits` anymore to remove the trigger characters when completing links.
    * You can customize the default behavior with the [`use-additional-text-edits` configuration key](docs/config-lsp.md).
* [#163](https://github.com/zk-org/zk/issues/163) Use the `ZK_SHELL` environment variable to override the shell for `zk` only.
* [#173](https://github.com/zk-org/zk/issues/173) Support for double star globbing in `note.ignore` config option.
* [#137](https://github.com/zk-org/zk/issues/137) Customize the `fzf` options used by `zk`'s interactive modes with the [`fzf-options`](docs/tool-fzf.md) config option (contributed by [@Nelyah](https://github.com/zk-org/zk/pull/154)).

* [#168](https://github.com/zk-org/zk/discussions/168) Customize the `fzf` key binding to create new notes with the [`fzf-bind-new`](docs/tool-fzf.md) config option.

### Changed

* The default `fzf` key binding to create a new note with `zk edit --interactive` was changed to `Ctrl-E`, to avoid conflict with the default `Ctrl-N` binding.

### Fixed

* [#126](https://github.com/zk-org/zk/issues/126) Embedded image links shown as not found.
* [#152](https://github.com/zk-org/zk/issues/152) Incorrect timezone for natural dates.
* [#170](https://github.com/zk-org/zk/issues/170) Broken wiki links in subdirectories.
* [#185](https://github.com/zk-org/zk/issues/185) Don't parse a Markdown table header as a colon tag.


## 0.9.0

### Added

* New LSP commands:
    * [`zk.list`](docs/editors-integration.md#zklist) to search for notes.
    * [`zk.tag.list`](docs/editors-integration.md#zktaglist) to retrieve the list of tags.
* `--debug` mode which prints a stacktrace on `SIGINT`.

### Fixed

* [#111](https://github.com/zk-org/zk/issues/111) Filenames take precedence over folders when matching a sub-path with wiki links.
* [#118](https://github.com/zk-org/zk/issues/118) Fix infinite loop when parsing a single-character hashtag.
* [#121](https://github.com/zk-org/zk/issues/121) Take into account the `--no-input` flag with `zk init`.
* [#120](https://github.com/zk-org/zk/discussions/120) Support RFC 3339 dates with the time flags (e.g. `--created-before`).


## 0.8.0

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
* You can now filter through the [YAML frontmatter](docs/note-frontmatter.md) with `zk list --interactive`.

### Fixed

* [#89](https://github.com/zk-org/zk/issues/89) Calling `zk index` from outside the notebook (contributed by [@adamreese](https://github.com/zk-org/zk/pull/90)).
* [#98](https://github.com/zk-org/zk/issues/98) Index wiki links using partial paths for `--linked-by` and `--link-to`.
* [#98](https://github.com/zk-org/zk/issues/98) Ignore spaces around the pipe in wiki links for LSP diagnostics.


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
* Use the `{{abs-path}}` template variable when [formatting notes](docs/template-format.md) to print the absolute path to the note (contributed by [@pstuifzand](https://github.com/zk-org/zk/pull/60)).
* A new `{{substring s index length}}` template helper extracts a portion of a given string, e.g.:
    * `{{substring 'A full quote' 2 4}}` outputs `full`
    * `{{substring 'A full quote' -5 5}` outputs `quote`

### Fixed

* UTF-8 handling in the LSP server.
* [#78](https://github.com/zk-org/zk/issues/78) Do not exclude notes containing broken links from the index.
* Allow setting the `--working-dir` and `--notebook-dir` flags before the `zk` subcommand when using aliases, e.g. `zk -W ~/notes my-alias`.
* [#86](https://github.com/zk-org/zk/issues/86) Index encoded Markdown links.


## 0.6.0

### Added

* Use JSON formats with `zk list` for easy post-processing:
    * `--format json` prints a plain JSON array.
    * `--format jsonl` prints one JSON note object per line, according to [JSON Lines](https://jsonlines.org/).
* The new `{{json}}` template helper serializes any template context variable into a valid JSON value, e.g.:
    * `{{json title}}` prints with quotes `"An interesting note"`
    * `{{json .}}` serializes the full template context as a JSON object.
* Use `--header` and `--footer` options with `zk list` to print arbitrary text at the start or end of the list.
* Support for LSP references to browse the backlinks of the link under the caret (contributed by [@pstuifzand](https://github.com/zk-org/zk/pull/58)).
* New [`note.ignore`](docs/config-note.md) configuration option to ignore files matching the given path globs when indexing notes.
    ```yaml
    [note]
    ignore = [
        "log-*.md"
        "drafts/*"
    ]
    ```

### Fixed

* [#16](https://github.com/zk-org/zk/issues/16) Links with section anchors, e.g. `[[filename#section]]`.
* Unicode support in wiki links. If you use accents or ideograms, please run `zk index --force` after upgrading to fix your index.


## 0.5.0

### Added

* [Editor integration through LSP](https://github.com/zk-org/zk/issues/22):
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
* Fix completion with Neovim's built-in LSP client (contributed by [@cormacrelf](https://github.com/zk-org/zk/pull/39)).


## 0.4.0

### Added

* Interactive wizard for the `zk init` command.
* An experimental Language Server for LSP-compatible editors:
    * Auto-complete Markdown links with `[[` (setup wiki links in the [note formats configuration](docs/note-format.md))
    * Auto-complete [hashtags and colon-separated tags](docs/tags.md).
    * Preview the content of a note when hovering a link.
    * Navigate in your notes by following internal links.
    * [And more to come...](https://github.com/zk-org/zk/issues/22)
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

* [#4](https://github.com/zk-org/zk/issues/4) Terminal borked when piping content with Vim


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

