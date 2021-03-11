# Changelog

All notable changes to this project will be documented in this file.

## Unreleased

### Added

* Support for tags.
    * Filter notes by their tags using `--tag "history, europe"`.
        * To match notes associated with either tags, use a pipe `|` or `OR` (all caps), e.g. `--tag "inbox OR todo"`.
        * If you want to exclude notes having a particular tag, prefix it with `-` or `NOT` (all caps), e.g. `--tag "NOT done"`.
        * Use glob patterns to match multiple tags, e.g. `--tag "book-*"`.
    * Many tag flavors are supported: `#hashtags`, `:colon:separated:tags:` ([opt-in](docs/note-format.md)) and even Bear's [`#multi-word tags#`](https://blog.bear.app/2017/11/bear-tips-how-to-create-multi-word-tags/) ([opt-in](docs/note-format.md)). If you prefer to use a YAML frontmatter, list your tags with the key `tags` or `keywords`.
* Print metadata from the YAML frontmatter in `list` output using `{{metadata.<key>}}`, e.g. `{{metadata.description}}`. Keys are normalized to lower case.
* Use the YAML frontmatter key `date` for the note creation date, when provided.
* Access environment variables from note templates with the `env.<key>` template variable, e.g. `{{env.PATH}}`.

### Changed

* Renamed `--linking-to` filtering option to `--link-to`.
* Multiple `--extra` variables are now separated by `,` instead of `;`.
