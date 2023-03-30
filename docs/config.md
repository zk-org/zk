# Configuration file

Each [notebook](notebook.md) contains a configuration file used to customize your experience with `zk`. This file is located at `.zk/config.toml` and uses the [TOML format](https://github.com/toml-lang/toml). It is composed of several optional sections:

* `[notebook]` configures the [default notebook](config-notebook.md)
* `[note]` sets the [note creation rules](config-note.md)
* `[extra]` contains free [user variables](config-extra.md) which can be expanded in templates
* `[group]` defines [note groups](config-group.md) with custom rules
* `[format]` configures the [note format settings](note-format.md), such as Markdown options
* `[tool]` customizes interaction with external programs such as:
    * [your default editor](tool-editor.md)
    * [your default shell](tool-shell.md)
    * [your default pager](tool-pager.md)
    * [`fzf`](tool-fzf.md)
* `[lsp]` setups the [Language Server Protocol settings](config-lsp.md) for [editors integration](editors-integration.md)
* `[filter]` declares your [named filters](config-filter.md)
* `[alias]` holds your [command aliases](config-alias.md)

## Global configuration file

You can also create a global configuration file to share aliases and settings across several notebooks. The global configuration is by default located at `~/.config/zk/config.toml`, but you can customize its location with the [`XDG_CONFIG_HOME`](https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html) environment variable.

Notebook configuration files will inherit the settings defined in the global configuration file. You can also share templates by storing them under `~/.config/zk/templates/`.

## Complete example

Here's an example of a complete configuration file:

```toml
# NOTEBOOK SETTINGS
[notebook]
dir = "~/notebook"

# NOTE SETTINGS
[note]

# Language used when writing notes.
# This is used to generate slugs or with date formats.
language = "en"

# The default title used for new note, if no `--title` flag is provided.
default-title = "Untitled"

# Template used to generate a note's filename, without extension.
filename = "{{id}}-{{slug title}}"

# The file extension used for the notes.
extension = "md"

# Template used to generate a note's content.
# If not an absolute path, it is relative to .zk/templates/
template = "default.md"

# Configure random ID generation.

# The charset used for random IDs.
id-charset = "alphanum"

# Length of the generated IDs.
id-length = 4

# Letter case for the random IDs.
id-case = "lower"


# EXTRA VARIABLES
[extra]
author = "Mickaël"


# GROUP OVERRIDES
[group.journal]
paths = ["journal/weekly", "journal/daily"]

[group.journal.note]
filename = "{{format-date now}}"


# MARKDOWN SETTINGS
[format.markdown]
# Enable support for #hashtags
hashtags = true
# Enable support for :colon:separated:tags:
colon-tags = true


# EXTERNAL TOOLS
[tool]

# Default editor used to open notes.
editor = "nvim"

# Default shell used by aliases and commands.
shell = "/bin/bash"

# Pager used to scroll through long output.
pager = "less -FIRX"

# Command used to preview a note during interactive fzf mode.
fzf-preview = "bat -p --color always {-1}"

# NAMED FILTERS
[filter]
recents = "--sort created- --created-after 'last two weeks'"

# COMMAND ALIASES
[alias]

# Edit the last modified note.
edlast = "zk edit --limit 1 --sort modified- $@"

# Edit the notes selected interactively among the notes created the last two weeks.
recent = "zk edit --sort created- --created-after 'last two weeks' --interactive"

# Show a random note.
lucky = "zk list --quiet --format full --sort random --limit 1"

# LSP (EDITOR INTEGRATION)
[lsp]

[lsp.diagnostics]
# Report titles of wiki-links as hints.
wiki-title = "hint"
# Warn for dead links between notes.
dead-link = "error"
```
