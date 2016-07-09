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


rm -rf merge
mkdir -p merge/src
cd merge
git init > /dev/null

export GIT_AUTHOR_NAME="Sammy Cobol"
export GIT_AUTHOR_EMAIL="<sammy.cobol@example.com>"
export GIT_AUTHOR_DATE="Sat, 24 Nov 1973 19:01:01 +0200"
export GIT_COMMITTER_NAME="Fred Foobar"
export GIT_COMMITTER_EMAIL="<fred.foobar@example.com>"
export GIT_COMMITTER_DATE="Sat, 24 Nov 1973 19:01:01 +0200"
echo -e "a\n\nb\n\nc\n\n" > src/foo
git add src/foo
git commit -m"init" > /dev/null

git co -b branch1 2> /dev/null
export GIT_AUTHOR_DATE="Sat, 24 Nov 1973 19:02:02 +0200"
export GIT_COMMITTER_DATE="Sat, 24 Nov 1973 19:02:02 +0200"
echo -e "a\n\nb\nchange 2\nc\n\n" > src/foo
git commit -a -m"change 2" > /dev/null

export GIT_AUTHOR_DATE="Sat, 24 Nov 1973 19:02:02 +0200"
export GIT_COMMITTER_DATE="Sat, 24 Nov 1973 19:02:02 +0200"
echo -e "a\n\nb\nchange 2\nc\nchange 3\n" > src/foo
git commit -a -m"change 3" > /dev/null

git co master 2> /dev/null
export GIT_AUTHOR_DATE="Sat, 24 Nov 1973 19:02:02 +0200"
export GIT_COMMITTER_DATE="Sat, 24 Nov 1973 19:02:02 +0200"
echo -e "a\nchange 1\nb\n\nc\n\n" > src/foo
git commit -a -m"change 1" > /dev/null

git co -b branch2 2> /dev/null
export GIT_AUTHOR_DATE="Sat, 24 Nov 1973 19:02:02 +0200"
export GIT_COMMITTER_DATE="Sat, 24 Nov 1973 19:02:02 +0200"
echo -e "a\n\nb\nchange 2\nc\n\n" > src/foo
git commit -a -m"change 2" > /dev/null

git co master 2> /dev/null
git co -b branch3 2> /dev/null
git merge branch1 --no-edit > /dev/null
git merge branch2 --no-edit -s ours > /dev/null

GIT_SUBTREE_SPLIT_SHA1_2="a2c4245703f8dac149ab666242a12e1d4b2510d9"
GIT_SUBTREE_SPLIT_SHA1_3="ba0dab2c4e99d68d11088f2c556af92851e93b14"
GIT_SPLITSH_SHA1_2=`$GOPATH/src/github.com/splitsh/lite/lite --git="<2.8.0" --prefix=src/ --quiet`
GIT_SPLITSH_SHA1_3=`$GOPATH/src/github.com/splitsh/lite/lite --prefix=src/ --quiet`

if [ "$GIT_SUBTREE_SPLIT_SHA1_2" == "$GIT_SUBTREE_SPLIT_SHA1_2" ]; then
    echo "OK ($GIT_SUBTREE_SPLIT_SHA1_2 == $GIT_SUBTREE_SPLIT_SHA1_2)"
else
    echo "OK ($GIT_SUBTREE_SPLIT_SHA1_2 != $GIT_SUBTREE_SPLIT_SHA1_2)"
    exit 1
fi

if [ "$GIT_SUBTREE_SPLIT_SHA1_3" == "$GIT_SUBTREE_SPLIT_SHA1_3" ]; then
    echo "OK ($GIT_SUBTREE_SPLIT_SHA1_3 == $GIT_SUBTREE_SPLIT_SHA1_3)"
else
    echo "OK ($GIT_SUBTREE_SPLIT_SHA1_3 != $GIT_SUBTREE_SPLIT_SHA1_3)"
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
