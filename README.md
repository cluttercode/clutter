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

`cat.c`:

```c
#include <stdio.h>

void cat() { // [# cat lang=c #]
	printf("meow\n")
}

```

`cat.hs`:

```haskell
cat = putStrLn("meow") -- [# cat lang=haskell #]
```

Clutter allows to:

```
$ clutter search cat
cat README.md:2
cat cat.c:3 lang=c
cat cat.go:6 lang=go
cat cat.hs:1 lang=haskell
cat cat.py:2 lang=python

$ clutter search -g cat lang=py\*
cat cat.py:2 lang=python
```

## Syntax

```
[# name attr1 attr2=value2 #]
```

`[# @attr name #]` translates to `[# name attr #]`, which is the same as `[# name attr= #]`.

`[# .name #]` translates to `[# name scope="current filename" #]`.

`[# ./name #]` translates to `[# name scope="current dir" #]`.

## Integrations

- [Vim Plugin](https://github.com/cluttercode/vim-clutter)

- [Sphinx Extension](https://github.com/cluttercode/sphinx-clutter)

## Installation

### Using gobinaries

```shell
curl -sf https://gobinaries.com/cluttercode/clutter/cmd/clutter@latest | sh
```

### Using go

```shell
go install github.com/cluttercode/clutter/cmd/clutter
```

### Prebuilt binaries

See [releases](https://github.com/cluttercode/clutter/releases).

### From source

```shell
make install
```

## TODO

- [ ] Docs: resolve, search
- [ ] More tests
- [ ] Cross repo.

- [ ] Only account for tags in comments.