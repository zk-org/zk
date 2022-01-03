<div align="center">
<h1>zk</h1>
<h4>A plain text note-taking assistant</h4>
<img alt="Screencast" width="95%" src="docs/assets/media/screencast.svg"/>
<p>Looking for a quick usage example? <a href="docs/getting-started.md">Let's get started</a>.</p>
</div>

## Description

`zk` is a command-line tool helping you to maintain a plain text [Zettelkasten](https://zettelkasten.de/introduction/) or [personal wiki](https://en.wikipedia.org/wiki/Personal_wiki).

### Highlights

* [Creating notes from templates](docs/note-creation.md)
* [Advanced search and filtering capabilities](docs/note-filtering.md) including [tags](docs/tags.md), links and mentions
* [Integration with your favorite editors](docs/editors-integration.md):
    * [Any LSP-compatible editor](docs/editors-integration.md)
    * [`zk-nvim`](https://github.com/mickael-menu/zk-nvim) for Neovim 0.5+
    * [`zk-vscode`](https://github.com/mickael-menu/zk-vscode) for Visual Studio Code
    * (*unmaintained*) [`zk.nvim`](https://github.com/megalithic/zk.nvim) for Neovim 0.5+ by [Seth Messer](https://github.com/megalithic)
* [Interactive browser](docs/tool-fzf.md), powered by `fzf`
* [Git-style command aliases](docs/config-alias.md) and [named filters](docs/config-filter.md)
* [Made with automation in mind](docs/automation.md)
* [Notebook housekeeping](docs/notebook-housekeeping.md)
* [Future-proof, thanks to Markdown](docs/future-proof.md)
* Supports most Markdown syntax flavors
    * Links: regular Markdown links, `[[Wikilinks]]` and Neuron's `[[Folgezettel links]]#`.
    * Tags: `#hashtags`, `:colon:separated:tags:`, Bear's `#multi-word tags#`.
    * [YAML frontmatter](docs/note-frontmatter.md)

[See the changelog](CHANGELOG.md) for the list of upcoming features waiting to be released.

### What `zk` is not

* A note editor.
* A tool to serve your notes on the web – for this, you may be interested in [Neuron](docs/neuron.md) or [Gollum](https://github.com/gollum/gollum).

## Install

[Check out the latest release](https://github.com/mickael-menu/zk/releases) for pre-built binaries for macOS and Linux (`zk` was not tested on Windows).

### Homebrew

A [Homebrew tap](https://github.com/mhanberg/homebrew-zk) is maintained by [@mhanberg](https://github.com/mhanberg).

```sh
brew install [--build-from-source] mhanberg/zk/zk
```

### Arch Linux

You can install [the zk package](https://archlinux.org/packages/community/x86_64/zk/) from the official repos.

```sh
sudo pacman -S zk
```

### Build from scratch

Make sure you have a working [Go installation](https://golang.org/), then clone the repository:

```sh
$ git clone https://github.com/mickael-menu/zk.git
$ cd zk
```

#### On macOS

`icu4c` is required to build `zk`, which you can install with [Homebrew](https://brew.sh/).

```
$ brew install icu4c
$ make
$ ./zk -h
```

#### On Linux

`libicu-dev` is required to build `zk`, use your favorite package manager to install it.

```
$ apt-install libicu-dev
$ make
$ ./zk -h
```

## Related projects

* [Neuron](https://github.com/srid/neuron) – a great tool to publish a Zettelkasten on the web
* [sirupsen's zk](https://github.com/sirupsen/zk) – a collection of scripts with a similar purpose
