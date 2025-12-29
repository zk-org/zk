# Modification Dates

When synchronizing notes across devices or using version control, the file's
modification time may not reflect when you actually last edited the note. Other
external tools also might be touching the file while not changing the contents.
`zk` offers custom [Date Keys](../notes/note-frontmatter.md#date-keys) to allow
overriding the modification date in the frontmatter of notes.

## Configuration

First, configure `zk` to use the `changed` frontmatter key to obtain the
modification date in `.zk/config.toml`:

```toml
[format.markdown.frontmatter]
modification-date-key = "changed"
```

Then update your notes' frontmatter to include the key with a date value:

```markdown
---
changed: 2025-12-29 09:48
---
```

## Automatically Updating Dates

Since manually changing the modification date is tedious, external tools can be
automated to automatically update the date in the `changed` key of the
frontmatter.

### Neovim with Autocommands

Add this to your Neovim configuration to update the `changed` field
whenever you save a Markdown file:

```lua
vim.api.nvim_create_autocmd("BufWritePre", {
  pattern = "*.md",
  callback = function(args)
    local lines = vim.api.nvim_buf_get_lines(args.buf, 0, -1, false)
    local in_frontmatter = false

    for i, line in ipairs(lines) do
      if line:match("^%-%-%-$") then
        if not in_frontmatter then
          in_frontmatter = true
        else
          break
        end
      elseif in_frontmatter and line:match("^changed:%s") then
        local new_line = "changed: " .. os.date("%Y-%m-%d %H:%M")
        vim.api.nvim_buf_set_lines(args.buf, i - 1, i, false, { new_line })
        return
      end
    end
  end,
})
```

### Git Pre-commit Hook

Create a Git pre-commit hook to automatically update modification dates before committing.
Save this as `.git/hooks/pre-commit` in your notebook repository:

```bash
#!/bin/bash
# Update modification date in frontmatter for all staged . md files

for file in $(git diff --cached --name-only -- '*.md'); do
  if [ -f "$file" ] && grep -q '^changed:' "$file"; then
    sed -i. bak "s/^changed:.*/changed: $(date +'%Y-%m-%d %H:%M:%S')/" "$file"
    rm -f "${file}.bak"
    git add "$file"
  fi
done

#!/usr/bin/env bash

git diff --cached --name-status | egrep -i "^(A|M).*\.(md)$" | while read a file; do
    sed --in-place "/---.*/,/---.*/s/^changed: [0-9]\{4\}-[0-9]\{2\}-[0-9]\{2\} [0-9]\{2\}:[0-9]\{2\}$/changed: $(date "+%Y-%m-%d %H:%M")/" $file
    git add $file
done
```
