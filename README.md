Monorepo to Manyrepos made easy
===============================

**tl;dr**: **splitsh-lite** is a replacement for the `subtree split` Git
built-in command that is much faster and has more features at the same time.

When starting a new project, do you store all the code in one monolith
repository? Or are you creating many repositories?

Both strategies work well but both have drawbacks as well. **splitsh** helps use
both strategies at the same time by providing tools that automatically
synchronize a mono repository to many repositories.

**splitsh-lite** is a sub-project with the goal of providing a faster replacement
of the `git subtree split` command.

If you want to learn more about monorepo vs manyrepos, watch this 4-minute
lightning talk I gave at dotScale
(or [read the slides](https://speakerdeck.com/fabpot/a-monorepo-vs-manyrepos))...
or watch the longer version from
[DrupalCon](https://www.youtube.com/watch?v=4w3-f6Xhvu8).

The main **splitsh-lite** feature is its ability to create a branch in a repository
from one or many directories.

Installation
------------

Install libgit2:

```bash
go get -d github.com/libgit2/git2go
cd $GOPATH/src/github.com/libgit2/git2go
git checkout next
git submodule update --init
make install
```

Compiling

```bash
go build -o splitsh-lite github.com/splitsh/lite
```

If everything goes fine, a `splitsh-lite` binary should be available in the
current directory.

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
command again. Updates are much faster as **splitsh-lite** keeps a cache of already
split commits. Caching is possible as **splitsh-lite** guarantees that two splits of
the same code always results in the same history and the same `sha1`s for each
commit.

By default, **splitsh-lite** splits the current checkout-ed branch but you can split
a different branch by passing it explicitly with `--origin` (mandatory when
splitting a bare repository):

```bash
splitsh-lite --prefix=lib/ --origin=origin/1.0
```

You don't even need to run the command from the Git repository directory if you
pass the `--path` option:

```bash
splitsh-lite --prefix=lib/ --origin=origin/1.0 --path=/path/to/repo
```

Available options:

 * `--prefix` is the prefix of the directory to split; you can put the split
   contents in a directory by using the `--prefix=from:to` syntax; splitting
   several directories is also possible by passing multiple `--prefix` options;

 * `--path` is the path to the repository to split (current directory by default);

 * `--origin` is the Git reference for the origin (can be any Git reference
   like `HEAD`, `heads/xxx`, `tags/xxx`, `origin/xxx`, or any `refs/xxx`);

 * `--target` creates a reference for the tip of the split (can be any Git reference
   like `HEAD`, `heads/xxx`, `tags/xxx`, `origin/xxx`, or any `refs/xxx`);

 * `--progress` displays a nice progress bar during the split;

 * `--quiet` suppresses all output on stderr (useful when run from an automated
   script).

 * `--scratch` flushes the cache (useful when a branch is force pushed or in
   case of corruption)

 * `--legacy` simulates old versions of `git subtree split` where `sha1`s
   for the split commits were computed differently (useful if you are switching
   from the git command to **splitsh-lite**).

**splitsh** provides more features including a sanity checker, GitHub integration
for real-time splitting, tagging management and synchronization, and more.
It has been used by the Symfony project for many years but the tool is not yet
ready for Open-Source. Stay tuned!

If you think that your Open-Source project might benefit from the full version
of splitsh, send me an email and I will consider splitting your project for free
on my servers (like I do for Symfony and Laravel).
