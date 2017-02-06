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

Installation
------------

The fastest way to get started is to download a [binary][4] for your platform.

Usage
-----

Let say you want to split the `lib/` directory of a repository to its own
branch; from the "master" Git repository (bare or clone), run:

```bash
splitsh-lite --prefix=lib/
```

The *sha1* of the split is displayed at the end of the execution:

```bash
SHA1=`splitsh-lite --prefix=lib/`
```

The sha1 can be used to create a branch or to push the commits to a new
repository.

Automatically create a branch for the split by passing a branch name
via the `--target` option:

```bash
splitsh-lite --prefix=lib/ --target=branch-name
```

If new commits are made on the repository, update the split by running the same
command again. Updates are much faster as **splitsh-lite** keeps a cache of
already split commits. Caching is possible as **splitsh-lite** guarantees that
two splits of the same code always results in the same history and the same
`sha1`s for each commit.

By default, **splitsh-lite** splits the current checkout-ed branch but you can
split a different branch by passing it explicitly via the `--origin` flag
(mandatory when splitting a bare repository):

```bash
splitsh-lite --prefix=lib/ --origin=origin/master
```

You don't even need to run the command from the Git repository directory if you
pass the `--path` option:

```bash
splitsh-lite --prefix=lib/ --origin=origin/1.0 --path=/path/to/repo
```

Available options:

 * `--prefix` is the prefix of the directory to split; you can put the split
   contents in a sub-directory of the target repository by using the
   `--prefix=from:to` syntax; split several directories by passing multiple
   `--prefix` flags;

 * `--path` is the path of the repository to split (current directory by default);

 * `--origin` is the Git reference for the origin (can be any Git reference
   like `HEAD`, `heads/xxx`, `tags/xxx`, `origin/xxx`, or any `refs/xxx`);

 * `--target` creates a reference for the tip of the split (can be any Git
   reference like `HEAD`, `heads/xxx`, `tags/xxx`, `origin/xxx`, or any
   `refs/xxx`);

 * `--progress` displays a progress bar;

 * `--quiet` suppresses all output on stderr (useful when run from an automated
   script);

 * `--scratch` flushes the cache (useful when a branch is force pushed or in
   case of a cache corruption);

Migrating from `git subtree split`
----------------------------------

Migrating from `git subtree split` to `splith-lite` is easy as both tools
generate the same `sha1`s.

However, note that older versions of `git subtree split` used broken
algorithms, and so generated different `sha1`s than the latest version. You can
simulate those version via the `--git` flag. Use `<1.8.2` or `<2.8.0` depending
on which version of `git subtree split` you want to simulate.

Manual Installation
-------------------

If you want to contribute to `splitsh-lite` or use it as a library, you first
need to install `libgit2`:

```bash
go get -d github.com/libgit2/git2go
cd $GOPATH/src/github.com/libgit2/git2go
git checkout next
git submodule update --init
make install
```

Then, compile `splitsh-lite`:

```bash
go get github.com/splitsh/lite
go build -o splitsh-lite github.com/splitsh/lite
```

If everything goes fine, a `splitsh-lite` binary should be available in the
current directory.

Full-Version
------------

The full version of **splitsh** provides more features including a sanity
checker, GitHub integration for real-time splitting, tag management and
synchronization, and more. It has been used by the Symfony project for many
years but the tool is not yet ready for Open-Source. Stay tuned!

If you think that your Open-Source project might benefit from the full version
of splitsh, send me an email and I will consider splitting your project for
free on my servers (like I do for Symfony and Laravel).
