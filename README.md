Git Subtree Splitter
====================

**splitsh-lite** replaces the `subtree split` Git built-in command to make
**splitting a monolithic repository** to read-only standalone repositories
**easy and fast**.

Why do I need this tool?
------------------------

When starting a project, do you store all the code in one repository? Or are
you creating many standalone repositories?

Both strategies work well and both have drawbacks as well. **splitsh** helps
use both strategies by providing tools that automatically **synchronize a
monolithic repository to standalone repositories** in real-time.

**splitsh-lite** is a sub-project that provides a faster implementation of the
`git subtree split` command, which helps create standalone repositories for one
or more sub-directories of a main repository.

If you want to learn more about monorepo vs manyrepos, watch this [4-minute
lightning talk](http://www.thedotpost.com/2016/05/fabien-potencier-monolithic-repositories-vs-many-repositories)
I gave at dotScale
(or [read the slides](https://speakerdeck.com/fabpot/a-monorepo-vs-manyrepos))...
or watch the longer version from
[DrupalCon](https://www.youtube.com/watch?v=4w3-f6Xhvu8).
["The Monorepo - Storing your source code has never been so much fun"](https://speakerdeck.com/garethr/the-monorepo-storing-your-source-code-has-never-been-so-much-fun)
is also a great resource.

**Note** If you currently have multiple repositories that you want to merge into
a monorepo, use the [tomono](https://github.com/unravelin/tomono) tool.

Installation
------------

Docker (recommended)
--------------------

The recommended way to use the splitter is via the official Docker image:

Manual Installation (not recommended)
-------------------------------------

To build the binary , you first need to install `libgit2`, preferably using
your package manager of choice:

* Via brew:

  ```bash
  brew install libgit2@1.5
  ```

* Via apt:

  ```bash
  apt install libgit2-dev
  ```

Note that the last version of `libgit2` supported (by git2go) is 1.5.

If you get `libgit2` version `1.5`, you're all set and jump to the build step
below. If not, you first need to change the `git2go` version used in the code.
Using the table on the
[libgit2](https://github.com/libgit2/git2go#which-go-version-to-use)
repository, figure out which version of the `git2go` you need based on the
`liggit2` library you installed. Let's say you need version `v31`:

```bash
sed -i -e 's/v34/v31/g' go.mod splitter/*.go
go mod tidy
```

On MacOS, export the following flags:

```bash
export LDFLAGS="-L/opt/homebrew/opt/libgit2@1.5/lib"
export CPPFLAGS="-I/opt/homebrew/opt/libgit2@1.5/include"
export PKG_CONFIG_PATH="/opt/homebrew/opt/libgit2@1.5/lib/pkgconfig"
```

Then, build the `splitsh-lite` binary:

```bash
go build -o splitsh-lite github.com/splitsh/lite
```

If everything goes fine, a `splitsh-lite` binary should be available in the
current directory.

If you want to integrate splitsh with Git, install it like this (and use it via
`git splitsh`):

```bash
cp splitsh-lite "$(git --exec-path)"/git-splitsh
```

Usage
-----

Let's say you want to split the `lib/` directory of a repository stored in the
current directory from the current branch (bare or clone), run:

```bash
# Docker
docker run --rm -v $PWD:/data splitsh-lite --prefix=lib/

# Binary
splitsh-lite --prefix=lib/
```

The command outputs the *sha1* of the split:

```bash
SHA1=`splitsh-lite --prefix=lib/`
```

The sha1 can be used to create a branch or to push the commits to a new
repository.

Automatically create a branch for the split by passing a branch name
via the `--target` option:

```bash
splitsh-lite --prefix=lib/ --target=heads/branch-name
```

If new commits are made to the repository, update the split by running the same
command again. Updates are much faster as **splitsh-lite** keeps a cache of
already split commits. Caching is possible as **splitsh-lite** guarantees that
two splits of the same code always results in the same history and the same
`sha1`s for each commit.

By default, **splitsh-lite** splits the currently checked out branch but you can
split a different branch by passing it explicitly via the `--origin` flag
(mandatory when splitting a bare repository):

```bash
splitsh-lite --prefix=lib/ --origin=origin/master
```

You don't even need to run the command from the Git repository directory:

```bash
# Docker
docker run --rm -v /path/to/repo:/data splitsh-lite --prefix=lib/ --origin=origin/1.0

# Binary
splitsh-lite --prefix=lib/ --origin=origin/1.0 --path=/path/to/repo
```

Available options:

 * `--prefix` is the prefix of the directory to split; you can put the split
   contents in a sub-directory of the target repository by using the
   `--prefix=from:to` syntax; split several directories by passing multiple
   `--prefix` flags;

 * `--path` is the path of the repository to split (current directory by
   default, or use the `-v` option of Docker when using the Docker image);

 * `--origin` is the Git reference for the origin (can be any Git reference
   like `HEAD`, `heads/xxx`, `tags/xxx`, `origin/xxx`, or any `refs/xxx`);

 * `--target` creates a reference for the tip of the split (can be any Git
   reference like `heads/xxx`, `tags/xxx`, `origin/xxx`, or any `refs/xxx`);

 * `--progress` displays a progress bar;

 * `--scratch` flushes the cache (useful when a branch is force pushed or in
   case of a cache corruption).

Migrating from `git subtree split`
----------------------------------

Migrating from `git subtree split` to `splith-lite` is easy as both tools
generate the same `sha1`s.

However, note that older versions of `git subtree split` used broken
algorithms, and so generated different `sha1`s than the latest version. You can
simulate those version via the `--git` flag. Use `<1.8.2` or `<2.8.0` depending
on which version of `git subtree split` you want to simulate.
