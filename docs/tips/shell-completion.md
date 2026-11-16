# Shell completion

`zk` can teach your shell to complete its commands and flags, and — inside a
notebook — the names of your notes and tags.

```sh
$ zk list --tag <TAB>
biography  biology  book  dystopia  feminism  fiction  history
```

`zk completion` prints the script for a given shell. Where to put it depends on
your shell; run `zk completion <shell> --help` for the details.

## Bash

```sh
# Try it out in the current shell:
source <(zk completion bash)

# Install it for every new shell:
zk completion bash > /etc/bash_completion.d/zk          # Linux
zk completion bash > $(brew --prefix)/etc/bash_completion.d/zk   # macOS
```

Bash completion relies on the
[bash-completion](https://github.com/scop/bash-completion) package, which your
distribution most likely packages already.

## Zsh

```sh
# Install it somewhere on your $fpath:
zk completion zsh > "${fpath[1]}/_zk"
```

If completion is not enabled in your shell yet, add this to `~/.zshrc`:

```sh
autoload -Uz compinit && compinit
```

## Fish

```sh
# Try it out in the current shell:
zk completion fish | source

# Install it for every new shell:
zk completion fish > ~/.config/fish/completions/zk.fish
```

## PowerShell

```powershell
# Try it out in the current session:
zk completion powershell | Out-String | Invoke-Expression

# Install it by adding the line above to your PowerShell profile.
```

## What gets completed

Besides the commands and their flags, `zk` suggests:

* the notes of the current notebook, for `<path>` arguments and for the flags
  taking one, such as `--exclude` or `--link-to`;
* the tags of the current notebook, for `--tag`;
* the note groups declared in the notebook's configuration, for `--group`;
* the predefined formats, sorting terms and matching strategies, for
  `--format`, `--sort` and `--match-strategy`.

The notes and tags are read from the notebook's index, which `zk` brings up to
date as usual before answering. Completion outside a notebook falls back to the
commands and flags alone.

Aliases defined in your configuration are *not* completed: `zk` resolves them
before the command line is parsed, so it never sees them as commands. Your
shell will still complete the alias name itself if you define a shell alias for
it.
