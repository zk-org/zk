package core

import (
	"fmt"
	"path/filepath"

	"github.com/zk-org/zk/internal/util/errors"
)

// NotebookStore retrieves or creates new notebooks.
type NotebookStore struct {
	config          Config
	notebookFactory NotebookFactory
	templateLoader  TemplateLoader
	fs              FileStorage

	// Cached opened notebooks.
	notebooks map[string]*Notebook
}

type NotebookStorePorts struct {
	NotebookFactory NotebookFactory
	TemplateLoader  TemplateLoader
	FS              FileStorage
}

// NewNotebookStore creates a new NotebookStore instance using the given
// options and port implementations.
func NewNotebookStore(config Config, ports NotebookStorePorts) *NotebookStore {
	return &NotebookStore{
		config:          config,
		notebookFactory: ports.NotebookFactory,
		templateLoader:  ports.TemplateLoader,
		fs:              ports.FS,
		notebooks:       map[string]*Notebook{},
	}
}

// ErrNotebookNotFound is an error returned when a notebook cannot be found at the given path or its parents.
type ErrNotebookNotFound string

func (e ErrNotebookNotFound) Error() string {
	return fmt.Sprintf("no notebook found in %s or a parent directory", string(e))
}

// Open returns a new Notebook instance for the notebook containing the
// given file path.
func (ns *NotebookStore) Open(path string) (*Notebook, error) {
	wrap := errors.Wrapper("failed to open notebook")

	path = ns.fs.Canonical(path)
	nb := ns.cachedNotebookAt(path)
	if nb != nil {
		return nb, nil
	}

	path, err := ns.fs.Abs(path)
	if err != nil {
		return nil, wrap(err)
	}
	path, err = ns.locateNotebook(path)
	if err != nil {
		return nil, wrap(err)
	}

	configPath := filepath.Join(path, ".zk/config.toml")
	config, err := OpenConfig(configPath, ns.config, ns.fs, false)
	if err != nil {
		return nil, wrap(err)
	}

	nb, err = ns.notebookFactory(path, config)
	if err != nil {
		return nil, wrap(err)
	}
	ns.notebooks[path] = nb

	return nb, nil
}

// cachedNotebookAt returns any cached notebook containing the given path.
func (ns *NotebookStore) cachedNotebookAt(path string) *Notebook {
	path, err := ns.fs.Abs(path)
	if err != nil {
		return nil
	}

	for root, nb := range ns.notebooks {
		if isDesc, err := ns.fs.IsDescendantOf(root, path); isDesc && err == nil {
			return nb
		}
	}

	return nil
}

// InitOpts holds the user preferences when creating a new notebook.
type InitOpts struct {
	WikiLinks     bool
	Hashtags      bool
	ColonTags     bool
	MultiwordTags bool
}

// NewDefaultInitOpts creates a new instance of InitOpts with the default values.
func NewDefaultInitOpts() InitOpts {
	return InitOpts{
		WikiLinks:     true,
		Hashtags:      true,
		ColonTags:     false,
		MultiwordTags: false,
	}
}

// Init creates a new notebook at the given file path.
func (ns *NotebookStore) Init(path string, options InitOpts) (*Notebook, error) {
	wrap := errors.Wrapper("init")

	path, err := ns.fs.Abs(path)
	if err != nil {
		return nil, wrap(err)
	}

	if existingPath, err := ns.locateNotebook(path); err == nil {
		return nil, wrap(fmt.Errorf("a notebook already exists in %v", existingPath))
	}

	// Create the default configuration file.
	config, err := ns.generateConfig(options)
	if err != nil {
		return nil, wrap(err)
	}
	err = ns.fs.Write(filepath.Join(path, ".zk/config.toml"), []byte(config))
	if err != nil {
		return nil, wrap(err)
	}

	// Create the default template.
	err = ns.fs.Write(filepath.Join(path, ".zk/templates/default.md"), []byte(defaultTemplate))
	if err != nil {
		return nil, wrap(err)
	}

	return ns.Open(path)
}

// locateNotebook finds the root of the notebook containing the given path.
func (ns *NotebookStore) locateNotebook(path string) (string, error) {
	if !filepath.IsAbs(path) {
		panic("absolute path expected")
	}

	var locate func(string) (string, error)
	locate = func(currentPath string) (string, error) {
		// For Windows, the root dir may end with volume name, e.g. E:\\
		if currentPath == "/" || currentPath == filepath.VolumeName(currentPath)+"\\" || currentPath == "." {
			return "", ErrNotebookNotFound(path)
		}
		exists, err := ns.fs.DirExists(filepath.Join(currentPath, ".zk"))
		switch {
		case err != nil:
			return "", err
		case exists:
			return currentPath, nil
		default:
			return locate(filepath.Dir(currentPath))
		}
	}

	return locate(path)
}

func (ns *NotebookStore) generateConfig(options InitOpts) (string, error) {
	template, err := ns.templateLoader.LoadTemplate(defaultConfig)
	if err != nil {
		return "", err
	}
	return template.Render(options)
}

const defaultConfig = `# zk configuration file
#
# Uncomment the properties you want to customize.

# NOTE SETTINGS
#
# Defines the default options used when generating new notes.
[note]

# Language used when writing notes.
# This is used to generate slugs or with date formats.
#language = "en"

# The default title used for new note, if no ` + "`" + `--title` + "`" + ` flag is provided.
#default-title = "Untitled"

# Template used to generate a note's filename, without extension.
#filename = "\{{id}}"

# The file extension used for the notes.
#extension = "md"

# Template used to generate a note's content.
# If not an absolute path or "~/unix/path", it's relative to .zk/templates/
template = "default.md"

# Path globs ignored while indexing existing notes.
#ignore = [
#    "drafts/*",
#	"log.md"
#]

# Configure random ID generation.

# The charset used for random IDs. You can use:
#   * letters: only letters from a to z.
#   * numbers: 0 to 9
#   * alphanum: letters + numbers
#   * hex: hexadecimal, from a to f and 0 to 9
#   * custom string: will use any character from the provided value
#id-charset = "alphanum"

# Length of the generated IDs.
#id-length = 4

# Letter case for the random IDs, among lower, upper or mixed.
#id-case = "lower"


# EXTRA VARIABLES
#
# A dictionary of variables you can use for any custom values when generating
# new notes. They are accessible in templates with \{{extra.<key>}}
[extra]

#key = "value"


# GROUP OVERRIDES
#
# You can override global settings from [note] and [extra] for a particular
# group of notes by declaring a [group."<name>"] section.
#
# Specify the list of directories which will automatically belong to the group
# with the optional ` + "`" + `paths` + "`" + ` property.
#
# Omitting ` + "`" + `paths` + "`" + ` is equivalent to providing a single path equal to the name of
# the group. This can be useful to quickly declare a group by the name of the
# directory it applies to.

#[group."<NAME>"]
#paths = ["<DIR1>", "<DIR2>"]
#[group."<NAME>".note]
#filename = "\{{format-date now}}"
#[group."<NAME>".extra]
#key = "value"


# MARKDOWN SETTINGS
[format.markdown]

# Format used to generate links between notes.
# Either "wiki", "markdown" or a custom template. Default is "markdown".
{{#if WikiLinks}}
link-format = "wiki"
{{else}}
#link-format = "wiki"
{{/if}}
# Indicates whether a link's path will be percent-encoded.
# Defaults to true for "markdown" format and false for "wiki" format.
#link-encode-path = true
# Indicates whether a link's path file extension will be removed.
# Defaults to true.
#link-drop-extension = true

# Enable support for #hashtags.
{{#if Hashtags}}
hashtags = true
{{else}}
hashtags = false
{{/if}}
# Enable support for :colon:separated:tags:.
{{#if ColonTags}}
colon-tags = true
{{else}}
colon-tags = false
{{/if}}
# Enable support for Bear's #multi-word tags#
# Hashtags must be enabled for multi-word tags to work.
{{#if MultiwordTags}}
multiword-tags = true
{{else}}
multiword-tags = false
{{/if}}


# EXTERNAL TOOLS
[tool]

# Default editor used to open notes. When not set, the EDITOR or VISUAL
# environment variables are used.
#editor = "vim"

# Pager used to scroll through long output. If you want to disable paging
# altogether, set it to an empty string "".
#pager = "less -FIRX"

# Command used to preview a note during interactive fzf mode.
# Set it to an empty string "" to disable preview.

# bat is a great tool to render Markdown document with syntax highlighting.
#https://github.com/sharkdp/bat
#fzf-preview = "bat -p --color always {-1}"


# LSP
#
#   Configure basic editor integration for LSP-compatible editors.
#   See https://github.com/zk-org/zk/blob/main/docs/editors-integration.md
#
[lsp]

[lsp.diagnostics]
# Each diagnostic can have for value: none, hint, info, warning, error

# Report titles of wiki-links as hints.
#wiki-title = "hint"
# Warn for dead links between notes.
dead-link = "error"

[lsp.completion]
# Customize the completion pop-up of your LSP client.

# Show the note title in the completion pop-up, or fallback on its path if empty.
#note-label = "\{{title-or-path}}"
# Filter out the completion pop-up using the note title or its path.
#note-filter-text = "\{{title}} \{{path}}"
# Show the note filename without extension as detail.
#note-detail = "\{{filename-stem}}"


# NAMED FILTERS
#
#    A named filter is a set of note filtering options used frequently together.
#
[filter]

# Matches the notes created the last two weeks. For example:
#    $ zk list recents --limit 15
#    $ zk edit recents --interactive
#recents = "--sort created- --created-after 'last two weeks'"


# COMMAND ALIASES
#
#   Aliases are user commands called with ` + "`" + `zk <alias> [<flags>] [<args>]` + "`" + `.
#
#   The alias will be executed with ` + "`" + `$SHELL -c` + "`" + `, please refer to your shell's
#   man page to see the available syntax. In most shells:
#     * $@ can be used to expand all the provided flags and arguments
#     * you can pipe commands together with the usual | character
#
[alias]
# Here are a few aliases to get you started.

# Shortcut to a command.
#ls = "zk list $@"

# Default flags for an existing command.
#list = "zk list --quiet $@"

# Edit the last modified note.
#editlast = "zk edit --limit 1 --sort modified- $@"

# Edit the notes selected interactively among the notes created the last two weeks.
# This alias doesn't take any argument, so we don't use $@.
#recent = "zk edit --sort created- --created-after 'last two weeks' --interactive"

# Print paths separated with colons for the notes found with the given
# arguments. This can be useful to expand a complex search query into a flag
# taking only paths. For example:
#   zk list --link-to "` + "`" + `zk path -m potatoe` + "`" + `"
#path = "zk list --quiet --format \{{path}} --delimiter , $@"

# Show a random note.
#lucky = "zk list --quiet --format full --sort random --limit 1"

# Returns the Git history for the notes found with the given arguments.
# Note the use of a pipe and the location of $@.
#hist = "zk list --format path --delimiter0 --quiet $@ | xargs -t -0 git log --patch --"

# Edit this configuration file.
#conf = '$EDITOR "$ZK_NOTEBOOK_DIR/.zk/config.toml"'
`

const defaultTemplate = `# {{title}}

{{content}}
`
