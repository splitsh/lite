#!/bin/bash

set -e
set -f

if [ ! -d splitter-lite-tests ]; then
    mkdir splitter-lite-tests
fi
cd splitter-lite-tests

rm -rf simple
mkdir simple
cd simple
git init > /dev/null

export GIT_AUTHOR_NAME="Sammy Cobol"
export GIT_AUTHOR_EMAIL="<sammy.cobol@example.com>"
export GIT_AUTHOR_DATE="Sat, 24 Nov 1973 19:01:02 +0200"
export GIT_COMMITTER_NAME="Fred Foobar"
export GIT_COMMITTER_EMAIL="<fred.foobar@example.com>"
export GIT_COMMITTER_DATE="Sat, 24 Nov 1973 19:11:22 +0200"
echo "a" > a
git add a
git commit -m"added a" > /dev/null

export GIT_AUTHOR_NAME="Fred Foobar"
export GIT_AUTHOR_EMAIL="<fred.foobar@example.com>"
export GIT_AUTHOR_DATE="Sat, 24 Nov 1973 20:01:02 +0200"
export GIT_COMMITTER_NAME="Sammy Cobol"
export GIT_COMMITTER_EMAIL="<sammy.cobol@example.com>"
export GIT_COMMITTER_DATE="Sat, 24 Nov 1973 20:11:22 +0200"
mkdir b/
echo "b" > b/b
git add b
git commit -m"added b" > /dev/null

export GIT_AUTHOR_NAME="Fred Foobar"
export GIT_AUTHOR_EMAIL="<fred.foobar@example.com>"
export GIT_AUTHOR_DATE="Sat, 24 Nov 1973 21:01:02 +0200"
export GIT_COMMITTER_NAME="Sammy Cobol"
export GIT_COMMITTER_EMAIL="<sammy.cobol@example.com>"
export GIT_COMMITTER_DATE="Sat, 24 Nov 1973 21:11:22 +0200"
echo "aa" > a
git add a
git commit -m"updated a" > /dev/null

export GIT_AUTHOR_NAME="Fred Foobar"
export GIT_AUTHOR_EMAIL="<fred.foobar@example.com>"
export GIT_AUTHOR_DATE="Sat, 24 Nov 1973 22:01:02 +0200"
export GIT_COMMITTER_NAME="Sammy Cobol"
export GIT_COMMITTER_EMAIL="<sammy.cobol@example.com>"
export GIT_COMMITTER_DATE="Sat, 24 Nov 1973 22:11:22 +0200"
git rm a > /dev/null
git commit -m"updated a" > /dev/null

export GIT_AUTHOR_NAME="Fred Foobar"
export GIT_AUTHOR_EMAIL="<fred.foobar@example.com>"
export GIT_AUTHOR_DATE="Sat, 24 Nov 1973 23:01:02 +0200"
export GIT_COMMITTER_NAME="Sammy Cobol"
export GIT_COMMITTER_EMAIL="<sammy.cobol@example.com>"
export GIT_COMMITTER_DATE="Sat, 24 Nov 1973 23:11:22 +0200"
echo "bb" > b/b
git add b/
git commit -m"updated b" > /dev/null

GIT_SUBTREE_SPLIT_SHA1=`git subtree split --prefix=b/ -q`
GIT_SPLITSH_SHA1=`$GOPATH/src/github.com/splitsh/lite/lite --prefix=b/ --quiet`

if [ "$GIT_SUBTREE_SPLIT_SHA1" == "$GIT_SUBTREE_SPLIT_SHA1" ]; then
    echo "OK ($GIT_SUBTREE_SPLIT_SHA1 == $GIT_SUBTREE_SPLIT_SHA1)"
else
    echo "OK ($GIT_SUBTREE_SPLIT_SHA1 != $GIT_SUBTREE_SPLIT_SHA1)"
    exit 1
fi

GIT_SUBTREE_SPLIT_SHA1=`git subtree split --prefix=b/ -q bff8cdfaaf78a8842b8d9241ccfd8fb6e026f508...`
GIT_SPLITSH_SHA1=`$GOPATH/src/github.com/splitsh/lite/lite --prefix=b/ --quiet --commit=bff8cdfaaf78a8842b8d9241ccfd8fb6e026f508`

if [ "$GIT_SUBTREE_SPLIT_SHA1" == "$GIT_SUBTREE_SPLIT_SHA1" ]; then
    echo "OK ($GIT_SUBTREE_SPLIT_SHA1 == $GIT_SUBTREE_SPLIT_SHA1)"
else
    echo "OK ($GIT_SUBTREE_SPLIT_SHA1 != $GIT_SUBTREE_SPLIT_SHA1)"
    exit 1
fi

cd ../

# run on some Open-Source repositories
if [ ! -d Twig ]; then
    git clone https://github.com/twigphp/Twig > /dev/null
fi
GIT_SUBTREE_SPLIT_SHA1="ea449b0f2acba7d489a91f88154687250d2bdf42"
GIT_SPLITSH_SHA1=`$GOPATH/src/github.com/splitsh/lite/lite --prefix=lib/ --origin=refs/tags/v1.24.1 --path=Twig --quiet --scratch`

if [ "$GIT_SUBTREE_SPLIT_SHA1" == "$GIT_SUBTREE_SPLIT_SHA1" ]; then
    echo "OK ($GIT_SUBTREE_SPLIT_SHA1 == $GIT_SUBTREE_SPLIT_SHA1)"
else
    echo "OK ($GIT_SUBTREE_SPLIT_SHA1 != $GIT_SUBTREE_SPLIT_SHA1)"
    exit 1
fi

cd ../
