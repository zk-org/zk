# Alias function for managing project notes.

This is a bash based extension for managing projects using zk.
This does not provide other typical project management tasks, I may extend this in the future for additional zettelkasten behaviors.

## Assumptions:

This assumes a few things. 
- `rg` `ripgrep` is installed. (You can change the `rg` calls to `grep` if you like though.
- Using bash 4.0 or later with the use of `readarray` You can change this line to use read instead if you run an older version of bash.
- You have a folder at `$ZK_NOTEBOOK_DIR/projects`

## Usage:

```bash
zk p "Project Name"
```

You can omit the `"`s if you are using single word names.
If there is a single match it will open immediately.
If there are multiple matches it will bring you into interactive list mode filtered to the original search.
If there is no match a new project will be created in the `projects` dir.
Currently `-M re` is set on all listings.
If you want to use for searching `zk p "ANY REGEX IN HERE WILL WORK" -s`.
The -s flag (currently) has to come after the query and prevents a new file from being created.

## Configuration:

Here is a basic configuration assuming all of the assumptions hold. 

```toml
[group.projects]
paths = ["projects"]

[tool]
shell = "/bin/bash"

[alias]
p = '''
if [ $# -eq 0 ];
  then
    zk list projects -i
else
  query_string="$(zk list $ZK_NOTEBOOK_DIR/projects -q -M re -m "$1" | rg '\S')"
  if [ "$query_string" == "" ]; then
    matches=()
  else
    readarray -t matches <<< $query_string
  fi
  if [ "${#matches[@]}" -eq 1 ]; then
    zk edit "$(echo "${matches[0]}" | xargs -n1 echo | rg '\.md')"
  elif [ "${#matches[@]}" -eq 0 ]; then
    if [ "$2" == "-s" ]; then
      echo "No matches found"
    else
      zk new -W="$ZK_NOTEBOOK_DIR/projects" --title "$1"
    fi
  else
    zk list projects -iq -m "$1" -M re
  fi
fi
'''
```
