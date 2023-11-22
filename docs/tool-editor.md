# Setting your default editor

`zk` is not a text editor. Instead, it is designed to interface with your favorite editor to write your notes.

You can customize which editor to use either from the [configuration file](config.md) or environment variables. In order of precedence, `zk` will use:

1. `ZK_EDITOR` environment variable
2. `editor` configuration property
    ```toml
    [tool]
    editor = "vim"
    ```
3. `VISUAL` environment variable
4. `EDITOR` environment variable

When invoking the editor, `zk` will provide the following environment to the editor:

`ZK_QUERY`: the query interactively provided to `fzf` or otherwise the value passed with the `-m` flag

## Advanced editor usage

It is also possible to create a custom script for your editor, for example to jump to a specific instance of a search term.

Consider the following script:

```bash
#!/bin/bash

set -eou pipefail

grep -nEv '^[[:space:]]*$' "$1" \
    | fzf \
        --tiebreak=begin \
        --exact \
        --tabstop=4 \
        --height=100% \
        --layout=reverse \
        --no-hscroll \
        --color=hl:-1,hl+:-1 \
        --preview-window=wrap \
        --delimiter=':' \
        --with-nth=2.. \
        --query="${ZK_QUERY}" \
    | sed 's/:.*$//' \
    | xargs -o -I {} vim +{} "$1"
```

This script could then be configured as the `editor` in `.zk/config.yaml` to open a second `fzf` instance prepopulated with the query which was previously entered into `zk`.
