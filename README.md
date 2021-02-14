<!-- [# %stop #] -->

# [# Clutter #]

Clutter facilitates an unintrusive way to link textual content in a source tree. Think ctags for comments.

The motivation is to be able to link various concepts in the code base, making implicit links between code segments explicit.

## TL;DR

Given:

`README.md`:

```
The code here include implementations for functions in each language to print a cat vocalization.
See [# cat #] for all implementations.
```

`cat.go`:

```go
package cat

import "fmt"

func cat() { // [# cat lang=go #]
  fmt.Println("meow")
}
```

`cat.py`:

```python
def cat(): # [# cat lang=python #]
  print("meow")
```

`cat.hs`:

```haskell
cat = putStrLn("meow") -- [# cat lang=haskell #]
```

Clutter allows to:

```
$ clutter search cat
cat README.md:2.5-13
cat cat.go:5.17-33 lang=go
cat cat.hs:1.27-48 lang=haskell
cat cat.py:1.14-34 lang=python

$ clutter search -g cat lang=py\*
cat cat.py:1.14-34 lang=python

$ clutter resolve --loc README.md:2.5
cat README.md:2.5-13
cat cat.go:5.17-33 lang=go
cat cat.hs:1.27-48 lang=haskell
cat cat.py:1.14-34 lang=python

$ clutter resolve --loc README.md:2 --next
cat cat.go:5.17-33 lang=go
```

## Tag Syntax

Each tag has a name and key-value attributes. Each attribute key must be unique.

```
[# name attr1 attr2=value2 ... #]
```

- `name` must satisfy the regular expression  `[\w_][\w_\:\-\.\/]`.
- `attr`s (keys) must satisfy the regular expression `[\w_][\w_\:\-]`.

### Special Attributes

- `scope` denotes where to look for matching tags. This can be either a path to a specific file, or to a directory and end with a `/`.

  **NOTE**: this is experimental and might change in the future.

- `search` is used for search tags.  See below.

### Syntactic Sugar

- `[# @attr name #]` translates to `[# name attr #]`, which is the same as `[# name attr= #]`.

- `[# .name #]` translates to `[# name scope="current filename" #]`.

- `[# ./name #]` translates to `[# name scope="current dir" #]`.

### Search Tags

Search tags are tags that instead of declaring a specific place in the code, denote a pattern to search for.

- `[# ?gl * lang=hs #]` means search for any tag name (using glob) that has `lang=hs`.

- `[# ? cat #]` means search for all tags with the name `cat` using exact match.

- `[# ?re "^c.+"  ]` search for all tags that begin with a `c` (using regexp) and has at least one more character in their names.

While these tags are indexed, clutter will never return these as a search/resolve result.

These tags are written as a normal tag to the index, with an added attribute `search` that contain the type of matcher used. For example:

```
* somewhere:1.1-19 lang=hs search=glob
```

## Index

To generate an index, use:

```
$ clutter index
```

This creates an index, which by default written to `.clutter/index`. This might be useful for very large repositories to speed up other commands. The index will be useful in the future for searching a repository for tags without the need to clone it first. 

By default an index is not used. An index can be used by either specifying its filenames using the `-i` option, or a configuration field.

### Structure

Each index entry is of the form:

```
name path:line.startcol-endcol attrs
```

`line`, `startcol` and `endcol` start at 1.  `attrs` are in a `key=value` format and are sorted. The index as a whole is sorted first by the tag name, then its location, then its scope, and last the rest of the sorted attributes. Essentially, `cat .clutter/index | sort` should have the same output as `cat .clutter/index`.

The index is treated as a `csv` file with a single space as a field delimiter. If any other spaces present in any other field, expect it to be properly quoted by clutter.

All other commands try to read from the index first, and if it does not exist - scan the tree instead.

## Search

Clutter provides a CLI command to perform searchs on tags, for example:

```
$ clutter search -g \* loc=foo/\* lang=go
```

will search for all tags under the path `foo/` that has an attribute `lang` set to `go`.

`-g` denotes use of glob matching for all fields. `-e` denotes use of regex. If neither is specified, exact matching is used.

## Resolve

The `resolve` CLI command is built for used by IDEs. For example, it is used by [vim-clutter]() and [vscode-clutter](https://github.com/cluttercode/vim-clutter).

For example:

```
$ clutter resolve --loc README.md:2.5 --loc-from-stdin --next --cyclic
```

will find the next use of the tag located at `README.md:2.5`. If this is the last occurance, the first occurance is returned. `README.md` content is read from stdin, which is useful if that file is not saved yet. If the tag at loc is a search tag, it is treated the same way `search` does, meaning multiple results, if any, will always be returned.

A useful optimization that is implemented here by `resolve` is that if the tag pointed to by `--loc` is local (`.some-tag` or `sometag scope=README.md`), the tree is not scanned as the data in the file at loc is sufficient.

## Lint

**TODO**

## Configuration

By default clutter tries to read the file `.clutter/config.yaml` in the current directory. The full structure of the file is as follows, shown with default values:

```yaml
scanner:
  use-index: false  # ry to read the index first, else or if index does not exist - scan.
  ignore: [".git"]  # .gitignore formatted list of paths to ignore.
  bracket:          # bracket configuration.
    left: "[#"
    right: "#]"
```

## Integrations

- [Vim Plugin](https://github.com/cluttercode/vim-clutter)
- [VSCode Plugin](https://github.com/cluttercode/vscode-clutter)
- [Sphinx Extension](https://github.com/cluttercode/sphinx-clutter)

## Installation

### Using gobinaries

For trustful people.

```shell
curl -sf https://gobinaries.com/cluttercode/clutter/cmd/clutter@latest | sh
```

### Using go

For the already initiated.

```shell
go install github.com/cluttercode/clutter/cmd/clutter
```

### From source

Hardcore!

```shell
git clone github.com/cluttercode/clutter
cd clutter
make install
```

### Prebuilt binaries

See [releases](https://github.com/cluttercode/clutter/releases).


## TODO

- [ ] More tests.
- [ ] Cross repo.
- [ ] Only account for tags in comments.
