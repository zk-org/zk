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
* [Interactive browser](docs/tool-fzf), powered by `fzf`
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

`zk` was only tested on macOS and Linux.

### Build from scratch

Make sure you have a working [Go installation](https://golang.org/), with Go's `bin` directory in your `PATH`.

```sh
$ git clone https://github.com/mickael-menu/zk.git
$ cd zk
$ chmod a+x go
$ ./go install
$ zk -h
```

## Related projects

* [Neuron](https://github.com/srid/neuron) – a great tool to publish a Zettelkasten on the web
* [sirupsen's zk](https://github.com/sirupsen/zk) – a collection of scripts with a similar purpose
