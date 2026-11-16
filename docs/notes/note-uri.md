# Note URIs

A `zk://` URI is a portable address for a precise location in a notebook. It is
meant for linking back and forth between `zk` and the systems around it: paste
one into a task tracker, a chat message, a browser bookmark or another
notebook, and `zk open` will bring you back to the right file, line and column.

For linking *between* notes, keep using ordinary
[Markdown or wiki links](note-format.md) instead — `zk` does not resolve
`zk://` URIs found in note bodies.

## The URI format

```
zk://?v=/home/you/notes&f=ref/meeting.md&l=12&c=4
```

| Parameter | Required | Description                                             |
| --------- | -------- | ------------------------------------------------------- |
| `v`       | yes      | Absolute path of the notebook (vault) directory         |
| `f`       | yes      | Path of the file, relative to the notebook root         |
| `l`       | no       | Line to point at, starting at 1                         |
| `c`       | no       | Column to point at, starting at 1                       |

Values are percent-encoded, except for the path separators, which are left
as-is so that the URI stays readable. `c` is a best-effort character column:
editors disagree on how to count one, and some ignore it entirely.

Because `v` is an absolute path, a URI is only meaningful on a machine where the
notebook lives at that path.

## Creating URIs

`zk uri` prints the URI of a single note:

```sh
$ zk uri ref/meeting.md
zk://?v=/home/you/notes&f=ref/meeting.md

$ zk uri ref/meeting.md --line 12 --column 4
zk://?v=/home/you/notes&f=ref/meeting.md&l=12&c=4
```

The argument is a path, or any href the other commands would accept:

```sh
$ zk uri meeting
zk://?v=/home/you/notes&f=ref/meeting.md
```

To print them in bulk, `zk list --link` is a shorthand for the `uri`
[output format](note-format.md), which is also available as the `{{uri}}`
template variable:

```sh
$ zk list --link --tag inbox
$ zk list --format "{{title}}: {{uri}}"
```

### From your editor

The [LSP server](../tips/editors-integration.md) exposes a `zk.uri` command,
which returns the URI of a location in the document being edited. A language
server cannot reach your clipboard, so it hands the URI back to the editor,
which decides what to do with it. In Neovim:

```lua
vim.lsp.buf_request(0, "workspace/executeCommand", {
  command = "zk.uri",
  arguments = {
    vim.fn.getcwd(),
    { location = vim.lsp.util.make_range_params() },
  },
}, function(_, res)
  if res then
    vim.fn.setreg("+", res.uri)
    print(res.uri)
  end
end)
```

## Opening URIs

```sh
$ zk open "zk://?v=/home/you/notes&f=ref/meeting.md&l=12&c=4"
```

`zk open` works from anywhere: it finds the notebook from the URI itself, so it
needs neither a current notebook nor an up-to-date index. A URI arriving from
another program is not taken on trust — `zk` refuses to open it unless `v` is a
directory containing a `.zk` directory, `f` stays inside it, and the file
actually exists.

Your editor is taken from `ZK_EDITOR`, your own
[configuration](../config/tool-editor.md), `VISUAL` or `EDITOR`, as usual. It is
deliberately *not* read from the configuration of the notebook named in the URI,
since `tool.editor` is an arbitrary shell command and the URI comes from
elsewhere.

### Line and column support per editor

There is no common syntax for opening a file at a location, so `zk` picks one
from the name of the editor binary:

| Editor                                              | Syntax used               |
| --------------------------------------------------- | ------------------------- |
| `vi`, `view`, `vim`, `nvim`, `gvim`, `mvim`         | `+<line> <path>`          |
| `nano`                                              | `+<line>,<column> <path>` |
| `emacs`, `emacsclient`, `kak`, `micro`              | `+<line>:<column> <path>` |
| `hx`, `helix`, `subl`, `sublime_text`               | `<path>:<line>:<column>`  |
| `code`, `codium`, `cursor`, `windsurf`, and friends | `--goto <path>:<line>:<column>` |

Any other editor — including a wrapper such as `flatpak run …` — is given the
plain path, and the position is dropped. To get around that, point
`tool.editor` at a small script of your own which puts the arguments where your
editor wants them.

## Registering `zk://` with your desktop

`zk` makes no changes to your system; this is set up by hand. On Linux, write
`~/.local/share/applications/zk-uri.desktop`:

```ini
[Desktop Entry]
Type=Application
Name=zk
Exec=zk open %u
Terminal=true
NoDisplay=true
MimeType=x-scheme-handler/zk;
```

Then register it as the handler for the scheme:

```sh
$ update-desktop-database ~/.local/share/applications
$ xdg-mime default zk-uri.desktop x-scheme-handler/zk
$ xdg-open "zk://?v=$ZK_NOTEBOOK_DIR&f=some-note.md&l=1"
```

A few things worth knowing:

* Use `%u`, not `%U`: `zk open` takes a single URI.
* Set `Terminal=true` for a terminal editor such as Vim, Helix or Kakoune, and
  `Terminal=false` for a graphical one such as VS Code or Sublime Text.
* A desktop session does not inherit your shell environment, so `ZK_EDITOR` is
  most likely unset and `zk` may not even be on its `PATH`. Use an absolute path
  in `Exec`, and set `tool.editor` in your
  [global configuration file](../config/config.md).

macOS needs a wrapper application declaring `CFBundleURLTypes`, and Windows a
`HKEY_CURRENT_USER\Software\Classes\zk` key. Neither is covered here.
