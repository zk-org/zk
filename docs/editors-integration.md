# Editors integration

There are several extensions available to integrate `zk` in your favorite editor:

* [`zk-nvim`](https://github.com/zk-org/zk.nvim) for Neovim 0.5+
* [`zk-vscode`](https://github.com/zk-org/zk-vscode) for Visual Studio Code

## Language Server Protocol

`zk` ships with a [Language Server](https://microsoft.github.io/language-server-protocol/overviews/lsp/overview/) to provide basic support for any LSP-compatible editor. The currently supported features are:

* Auto-complete Markdown links with `[[` (setup wiki-links in the [note formats configuration](note-format.md))
* Auto-complete [hashtags and colon-separated tags](tags.md).
* Preview the content of a note when hovering a link.
* Navigate in your notes by following internal links.
* Create a new note using the current selection as title.
* Diagnostics for dead links and wiki-links titles.
* [And more to come...](https://github.com/zk-org/zk/issues/22)
  
You can configure some of these features in your notebook's [configuration file](config-lsp.md).

### Editor LSP configurations

To start the Language Server, use the `zk lsp` command. Refer to the following sections for editor-specific examples. [Feel free to share the configuration for your editor](https://github.com/zk-org/zk/issues/22).

#### Vim and Neovim

##### Vim and Neovim 0.4

With [`coc.nvim`](https://github.com/neoclide/coc.nvim), run `:CocConfig` and add the following in the settings file:

<details><summary><tt>coc-settings.json</tt></summary>

```jsonc
{
  // Important, otherwise link completion containing spaces and other special characters won't work.
  "suggest.invalidInsertCharacters": [],

  "languageserver": {
    "zk": {
      "command": "zk",
      "args": ["lsp"],
      "trace.server": "messages",
      "filetypes": ["markdown"]
    },
  }
}
```
</details>

Here are some additional useful key bindings and custom commands:

<details><summary><tt>~/.config/nvim/init.vim</tt></summary>

```viml
" User command to index the current notebook.
"
" zk.index expects a notebook path as first argument, so we provide the current
" buffer path with expand("%:p").
command! -nargs=0 ZkIndex :call CocAction("runCommand", "zk.index", expand("%:p"))
nnoremap <leader>zi :ZkIndex<CR>

" User command to create and open a new note, to be called like this:
" :ZkNew {"title": "An interesting subject", "dir": "inbox", ...}
"
" Note the concatenation with the "edit" command to open the note right away.
command! -nargs=? ZkNew :exec "edit ".CocAction("runCommand", "zk.new", expand("%:p"), <args>).path

" Create a new note after prompting for its title.
nnoremap <leader>zn :ZkNew {"title": input("Title: ")}<CR>
" Create a new note in the directory journal/daily.
nnoremap <leader>zj :ZkNew {"dir": "journal/daily"}<CR>
```
</details>

##### Neovim 0.5 built-in LSP client

Using [`nvim-lspconfig`](https://github.com/neovim/nvim-lspconfig):

<details><summary><tt>~/.config/nvim/init.lua</tt></summary>

```lua
local lspconfig = require('lspconfig')
local configs = require('lspconfig/configs')

configs.zk = {
  default_config = {
    cmd = {'zk', 'lsp'},
    filetypes = {'markdown'},
    root_dir = function()
      return vim.loop.cwd()
    end,
    settings = {}
  };
}

lspconfig.zk.setup({ on_attach = function(client, buffer) 
  -- Add keybindings here, see https://github.com/neovim/nvim-lspconfig#keybindings-and-completion
end })
```
</details>

#### Sublime Text

Install the [Sublime LSP](https://github.com/sublimelsp/LSP) package, then run the **Preferences: LSP Settings** command. Add the following to the settings file:

<details><summary><tt>LSP.sublime-settings</tt></summary>

```jsonc
{
  "clients": {
    "zk": {
      "enabled": true,
      "command": ["zk", "lsp"],
      "languageId": "markdown",
      "scopes": [ "source.markdown" ],
      "syntaxes": [ "Packages/MarkdownEditing/Markdown.sublime-syntax" ]
    }
  }
}
```
</details>

#### Visual Studio Code

Install the [`zk-vscode`](https://marketplace.visualstudio.com/items?itemName=mickael-menu.zk-vscode) extension from the Marketplace.

### Custom commands

Using `zk`'s LSP custom commands, you can call `zk` commands right from your editor. Please refer to your editor's documentation on how to bind keyboard shortcuts to custom LSP commands.

#### `zk.index`

This LSP command calls `zk index` to refresh your notebook's index. It can be useful to make sure that the auto-completion is up-to-date. `zk.index` takes two arguments:

1. A path to a file or directory in the notebook to index.
2. <details><summary>(Optional) A dictionary of additional options (click to expand)</summary>
    
    | Key     | Type    | Description                       |
    |---------|---------|-----------------------------------|
    | `force` | boolean | Reindexes all the notes when true |
    </details>

`zk.index` returns a dictionary of indexing statistics.

#### `zk.new`

This LSP command calls `zk new` to create a new note. It can be useful to quickly create a new note with a key binding. `zk.new` takes two arguments:

1. A path to any file or directory in the notebook, to locate it.
2. <details><summary>(Optional) A dictionary of additional options (click to expand)</summary>
    
    | Key                       | Type                 | Description                                                                                                          |
    |---------------------------|----------------------|----------------------------------------------------------------------------------------------------------------------|
    | `title`                   | string               | Title of the new note                                                                                                |
    | `content`                 | string               | Initial content of the note                                                                                          |
    | `dir`                     | string               | Parent directory, relative to the root of the notebook                                                               |
    | `group`                   | string               | [Note configuration group](config-group.md)                                                                          |
    | `template`                | string               | [Custom template used to render the note](template-creation.md)                                                      |
    | `extra`                   | dictionary           | A dictionary of extra variables to expand in the template                                                            |
    | `date`                    | string               | A date of creation for the note in natural language, e.g. "tomorrow"                                                 |
    | `edit`                    | boolean              | When true, the editor will open the newly created note (**not supported by all editors**)                            |
    | `dryRun`                  | boolean              | When true, `zk` will not actually create the note on the file system, but will return its generated content and path |
    | `insertLinkAtLocation`    | location<sup>1</sup> | A location in another note where a link to the new note will be inserted                                             |
    | `insertContentAtLocation` | location<sup>1</sup> | A location in another note where the content of the new note will be inserted                                        |

    1. The `location` type is an [LSP Location object](https://microsoft.github.io/language-server-protocol/specification#location), for example:

    ```json
    {
        "uri":"file:///Users/mickael/notes/9se3.md",
        "range": {
            "end":{"line": 5, "character":149},
            "start":{"line": 5, "character":137}
        }
    }
    ```
    </details>

`zk.new` returns a dictionary with two properties:

* `path` containing the absolute path to the created note.
* `content` containing the raw content of the created note.

#### `zk.link`

This LSP command allows editors to tap into the note linking mechanism. It takes three arguments:

1. A `path` to any file in the notebook that will be linked to
2. An LSP `location` object that points to where the link will be inserted
3. An optional title of the link. If `title` is not provided, the title of the note will be inserted instead

`zk.link` returns a JSON object with the path to the linked note, if the linking was successful.

**Note**: This command is _not_ exposed in the command line. This command is targeted at editor / plugin authors to extend zk functionality.

#### `zk.list`

This LSP command calls `zk list` to search a notebook. It takes two arguments:

1. A path to any file or directory in the notebook, to locate it.
2. <details><summary>A dictionary of additional options (click to expand)</summary>
    
  | Key                | Type           | Required?   | Description                                                                                               |
  | ------------------ | -------------- | ----------- | -------------------------------------------------------------------------                                 |
  | `select`           | string array   | Yes         | List of note fields to return<sup>1</sup>                                                                 |
  | `hrefs`            | string array   | No          | Find notes matching the given path, including its descendants                                             |
  | `limit`            | integer        | No          | Limit the number of notes found                                                                           |
  | `match`            | string array   | No          | Terms to search for in the notes                                                                          |
  | `exactMatch`       | boolean        | No          | (deprecated: use `matchStrategy`) Search for exact occurrences of the `match` argument (case insensitive) |
  | `matchStrategy`    | string         | No          | Specify match strategy, which may be "fts" (default), "exact" or "re"                                     |
  | `excludeHrefs`     | string array   | No          | Ignore notes matching the given path, including its descendants                                           |
  | `tags`             | string array   | No          | Find notes tagged with the given tags                                                                     |
  | `mention`          | string array   | No          | Find notes mentioning the title of the given ones                                                         |
  | `mentionedBy`      | string array   | No          | Find notes whose title is mentioned in the given ones                                                     |
  | `linkTo`           | string array   | No          | Find notes which are linking to the given ones                                                            |
  | `linkedBy`         | string array   | No          | Find notes which are linked by the given ones                                                             |
  | `orphan`           | boolean        | No          | Find notes which are not linked by any other note                                                         |
  | `related`          | string array   | No          | Find notes which might be related to the given ones                                                       |
  | `maxDistance`      | integer        | No          | Maximum distance between two linked notes                                                                 |
  | `recursive`        | boolean        | No          | Follow links recursively                                                                                  |
  | `created`          | string         | No          | Find notes created on the given date                                                                      |
  | `createdBefore`    | string         | No          | Find notes created before the given date                                                                  |
  | `createdAfter`     | string         | No          | Find notes created after the given date                                                                   |
  | `modified`         | string         | No          | Find notes modified on the given date                                                                     |
  | `modifiedBefore`   | string         | No          | Find notes modified before the given date                                                                 |
  | `modifiedAfter`    | string         | No          | Find notes modified after the given date                                                                  |
  | `sort`             | string array   | No          | Order the notes by the given criterion                                                                    |

    1. As the output of this command might be very verbose and put a heavy load on the LSP client, you need to explicitly set which note fields you want to receive with the `select` option. The following fields are available: `filename`, `filenameStem`, `path`, `absPath`, `title`, `lead`, `body`, `snippets`, `rawContent`, `wordCount`, `tags`, `metadata`, `created`, `modified` and `checksum`.

    </details>

`zk.list` returns the found notes as a JSON array.

#### `zk.tag.list`

This LSP command calls `zk tag list` to return the list of tags in a notebook. It takes two arguments:

1. A path to any file or directory in the notebook, to locate it.
2. <details><summary>(Optional) A dictionary of additional options (click to expand)</summary>
    
    | Key    | Type         | Required? | Description                                      |
    |--------|--------------|-----------|--------------------------------------------------|
	| `sort` | string array | No        | Order the tags by the given criteria<sup>1</sup> |

    1. The available sort criteria are `name` and `note-count`. You can change the order by appending `-` or `+` to the criterion.

    </details>

`zk.tag.list` returns the tags as a JSON array.
