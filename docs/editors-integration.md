# Editors integration

There are several extensions available to integrate `zk` in your favorite editor:

* [`zk.nvim`](https://github.com/megalithic/zk.nvim) for Neovim 0.5+, maintained by [Seth Messer](https://github.com/megalithic)
* [`zk-vscode`](https://github.com/mickael-menu/zk-vscode) for Visual Studio Code

## Language Server Protocol

`zk` ships with a [Language Server](https://microsoft.github.io/language-server-protocol/overviews/lsp/overview/) to provide basic support for any LSP-compatible editor. The currently supported features are:

* Auto-complete Markdown links with `[[` (setup wiki-links in the [note formats configuration](note-format.md))
* Auto-complete [hashtags and colon-separated tags](tags.md).
* Preview the content of a note when hovering a link.
* Navigate in your notes by following internal links.
* [And more to come...](https://github.com/mickael-menu/zk/issues/22)

To start the Language Server, use the `zk lsp` command. Refer to the following sections for editor-specific examples. [Feel free to share the configuration for your editor](https://github.com/mickael-menu/zk/issues/22).

### Vim and Neovim

#### Vim and Neovim 0.4

With [`coc.nvim`](https://github.com/neoclide/coc.nvim), run `:CocConfig` and add the following in the settings file:

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

#### Neovim 0.5 built-in LSP client

Using [`nvim-lspconfig`](https://github.com/neovim/nvim-lspconfig):

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

### Sublime Text

Install the [Sublime LSP](https://github.com/sublimelsp/LSP) package, then run the **Preferences: LSP Settings** command. Add the following to the settings file:

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

### Visual Studio Code

Install the [`zk-vscode`](https://marketplace.visualstudio.com/items?itemName=mickael-menu.zk-vscode) extension from the Marketplace.
