<div align="center">
<h1>zk</h1>
<h4>A plain text note-taking assistant</h4>
<img alt="Screencast" width="95%" src="docs/assets/media/screencast.svg"/>
<p>Looking for a quick usage example? <a href="docs/getting-started.md">Let's get started</a>.</p>
</div>

## News: We Are In Maintenance Mode


> [!IMPORTANT] 
> As of January 2nd, the original brain behind zk, [Mickaël
> Menu](https://github.com/mickael-menu), made the difficult decision to retire
> from zk and the suite of programs supporting it. He put out a [call for
> maintainers](https://github.com/zk-org/zk/discussions/371), which has
> garnerned enough response to enable the project to continue! So zk is
> definitely still here for you.
>
> During this transition phase, we are placing the project into a maintenance
> mode, which means we are going to address existing issues and any teething
> problems with transferring the code bases to the new
> [zk-org](https://github.com/zk-org) organisation, which is where you can now
> find all the related projects. It also gives us new maintainers the space to
> get up to speed with the code base, which will help us address new issues and
> feature requests when they come.
>
> So for now, feel free to lodge new issues, but please withold on feature
> requests until we are out of maintenance mode. This will help keep our issues
> boards concise and pr's easier to manage.
>
> The [call to maintainers](https://github.com/zk-org/zk/discussions/371) is
> still open. Please comment there if you feel commited enough to come onboard!
> PR's, ideas, discussions and conversations are still and always will be
> warmly welcomed, with or without 'maintainer' status ❤️

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

```sh
brew install zk
```

Or, if you want to the latest changes:

```sh
brew install --HEAD zk
```

### Nix

```sh
# Run zk from Nix store without installing it:
nix run nixpkgs#zk
# Or, to install it permanently:
nix-env -iA zk
```

### Arch Linux

You can install [the zk package](https://archlinux.org/packages/extra/x86_64/zk/) from the official repos.

```sh
sudo pacman -S zk
```

### Build from scratch

Make sure you have a working [Go 1.18+ installation](https://golang.org/), then clone the repository:

```sh
$ git clone https://github.com/mickael-menu/zk.git
$ cd zk
```

#### On macOS

```
$ make
$ ./zk -h
```

#### On Linux

```
$ make
$ ./zk -h
```

## Related projects

* [Neuron](https://github.com/srid/neuron) – a great tool to publish a Zettelkasten on the web
* [Emanote](https://emanote.srid.ca/) – an improved successor to Neuron
* [sirupsen's zk](https://github.com/sirupsen/zk) – a collection of scripts with a similar purpose
* [zk-spaced](https://github.com/matze/zk-spaced) – spaced repetition plugin for zk
