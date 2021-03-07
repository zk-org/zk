# Changelog

All notable changes to this project will be documented in this file.

## Unreleased

### Added

* Support for tags.
    * Filter notes by their tags using `--tag "history, europe"`. To match notes associated with either tags, use a pipe `|` or `OR` (all caps), e.g. `--tag "inbox OR todo"`.
    * Many tag flavors are supported: `#hashtags`, `:colon:separated:tags:` and even Bear's [`#multi-word tags#`](https://blog.bear.app/2017/11/bear-tips-how-to-create-multi-word-tags/). If you prefer to use a YAML frontmatter, list your tags with the key `tags` or `keywords`.

### Changed

* Multiple `--extra` variables are now separated by `,` instead of `;`.
