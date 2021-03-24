# Changelog

All notable changes to this project will be documented in this file.

## Unreleased

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
