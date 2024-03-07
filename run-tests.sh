#!/bin/bash

set -euo pipefail

switchAsSammy()
{
    AUTHOR_DATE=$1
    COMMITTER_DATE=$2
    export GIT_AUTHOR_NAME="Sammy Cobol"
    export GIT_AUTHOR_EMAIL="<sammy.cobol@example.com>"
    export GIT_AUTHOR_DATE="${AUTHOR_DATE}"
    export GIT_COMMITTER_NAME="Fred Foobar"
    export GIT_COMMITTER_EMAIL="<fred.foobar@example.com>"
    export GIT_COMMITTER_DATE="${COMMITTER_DATE}"
}

switchAsFred() {
    AUTHOR_DATE=$1
    COMMITTER_DATE=$2
    export GIT_AUTHOR_NAME="Fred Foobar"
    export GIT_AUTHOR_EMAIL="<fred.foobar@example.com>"
    export GIT_AUTHOR_DATE="${AUTHOR_DATE}"
    export GIT_COMMITTER_NAME="Sammy Cobol"
    export GIT_COMMITTER_EMAIL="<sammy.cobol@example.com>"
    export GIT_COMMITTER_DATE="${COMMITTER_DATE}"
}

simpleTest() {
    rm -rf simple
    mkdir simple
    cd simple
    git init > /dev/null

    switchAsSammy "Sat, 24 Nov 1973 19:01:02 +0200" "Sat, 24 Nov 1973 19:11:22 +0200"
    echo "a" > a
    git add a
    git commit -m"added a" > /dev/null

    switchAsFred "Sat, 24 Nov 1973 20:01:02 +0200" "Sat, 24 Nov 1973 20:11:22 +0200"
    mkdir b/
    echo "b" > b/b
    git add b
    git commit -m"added b" > /dev/null

    switchAsFred "Sat, 24 Nov 1973 21:01:02 +0200" "Sat, 24 Nov 1973 21:11:22 +0200"
    echo "aa" > a
    git add a
    git commit -m"updated a" > /dev/null

    switchAsFred "Sat, 24 Nov 1973 22:01:02 +0200" "Sat, 24 Nov 1973 22:11:22 +0200"
    git rm a > /dev/null
    git commit -m"updated a" > /dev/null
    SHA=`git rev-parse HEAD`

    switchAsFred "Sat, 24 Nov 1973 23:01:02 +0200" "Sat, 24 Nov 1973 23:11:22 +0200"
    echo "bb" > b/b
    git add b/
    git commit -m"updated b" > /dev/null

    GIT_SUBTREE_SPLIT_SHA1=`git subtree split --prefix=b/ -q`
    GIT_SPLITSH_SHA1=`$LITE_PATH --prefix=b/ 2>/dev/null`

    if [ "$GIT_SUBTREE_SPLIT_SHA1" == "$GIT_SPLITSH_SHA1" ]; then
        echo "Test #1 - OK ($GIT_SUBTREE_SPLIT_SHA1 == $GIT_SPLITSH_SHA1)"
    else
        echo "Test #1 - NOT OK ($GIT_SUBTREE_SPLIT_SHA1 != $GIT_SPLITSH_SHA1)"
        exit 1
    fi

    GIT_SUBTREE_SPLIT_SHA1=`git subtree split --prefix=b/ -q $SHA`
    GIT_SPLITSH_SHA1=`$LITE_PATH --prefix=b/ --commit=$SHA 2>/dev/null`

    if [ "$GIT_SUBTREE_SPLIT_SHA1" == "$GIT_SPLITSH_SHA1" ]; then
        echo "Test #2 - OK ($GIT_SUBTREE_SPLIT_SHA1 == $GIT_SPLITSH_SHA1)"
    else
        echo "Test #2 - NOT OK ($GIT_SUBTREE_SPLIT_SHA1 != $GIT_SPLITSH_SHA1)"
        exit 1
    fi

    cd ../
}

mergeTest() {
    rm -rf merge
    mkdir -p merge/src
    cd merge
    git init > /dev/null

    switchAsSammy "Sat, 24 Nov 1973 19:01:01 +0200" "Sat, 24 Nov 1973 19:01:01 +0200"
    echo -e "a\n\nb\n\nc\n\n" > src/foo
    git add src/foo
    git commit -m"init" > /dev/null

    git checkout -b branch1 2> /dev/null

    switchAsSammy "Sat, 24 Nov 1973 19:02:02 +0200" "Sat, 24 Nov 1973 19:02:02 +0200"
    echo -e "a\n\nb\nchange 2\nc\n\n" > src/foo
    git commit -a -m"change 2" > /dev/null

    switchAsSammy "Sat, 24 Nov 1973 19:02:02 +0200" "Sat, 24 Nov 1973 19:02:02 +0200"
    echo -e "a\n\nb\nchange 2\nc\nchange 3\n" > src/foo
    git commit -a -m"change 3" > /dev/null

    git checkout main 2> /dev/null
    switchAsSammy "Sat, 24 Nov 1973 19:02:02 +0200" "Sat, 24 Nov 1973 19:02:02 +0200"
    echo -e "a\nchange 1\nb\n\nc\n\n" > src/foo
    git commit -a -m"change 1" > /dev/null

    git checkout -b branch2 2> /dev/null
    switchAsSammy "Sat, 24 Nov 1973 19:02:02 +0200" "Sat, 24 Nov 1973 19:02:02 +0200"
    echo -e "a\n\nb\nchange 2\nc\n\n" > src/foo
    git commit -a -m"change 2" > /dev/null

    git checkout main 2> /dev/null
    git checkout -b branch3 2> /dev/null
    git merge branch1 --no-edit > /dev/null
    git merge branch2 --no-edit -s ours > /dev/null

    GIT_SUBTREE_SPLIT_SHA1_2="a2c4245703f8dac149ab666242a12e1d4b2510d9"
    GIT_SUBTREE_SPLIT_SHA1_3="ba0dab2c4e99d68d11088f2c556af92851e93b14"
    GIT_SPLITSH_SHA1_2=`$LITE_PATH --git="<2.8.0" --prefix=src/ 2>/dev/null`
    GIT_SPLITSH_SHA1_3=`$LITE_PATH --prefix=src/ 2>/dev/null`

    if [ "$GIT_SUBTREE_SPLIT_SHA1_2" == "$GIT_SPLITSH_SHA1_2" ]; then
        echo "Test #3 - OK ($GIT_SUBTREE_SPLIT_SHA1_2 == $GIT_SPLITSH_SHA1_2)"
    else
        echo "Test #3 - NOT OK ($GIT_SUBTREE_SPLIT_SHA1_2 != $GIT_SPLITSH_SHA1_2)"
        exit 1
    fi

    if [ "$GIT_SUBTREE_SPLIT_SHA1_3" == "$GIT_SPLITSH_SHA1_3" ]; then
        echo "Test #4 - OK ($GIT_SUBTREE_SPLIT_SHA1_3 == $GIT_SPLITSH_SHA1_3)"
    else
        echo "Test #4 - NOT OK ($GIT_SUBTREE_SPLIT_SHA1_3 != $GIT_SPLITSH_SHA1_3)"
        exit 1
    fi

    cd ../
}

twigSplitTest() {
    # run on some Open-Source repositories
    if [ ! -d Twig ]; then
        git clone https://github.com/twigphp/Twig > /dev/null
    fi
    GIT_SUBTREE_SPLIT_SHA1="ea449b0f2acba7d489a91f88154687250d2bdf42"
    GIT_SPLITSH_SHA1=`$LITE_PATH --prefix=lib/ --origin=refs/tags/v1.24.1 --path=Twig --scratch 2>/dev/null`

    if [ "$GIT_SUBTREE_SPLIT_SHA1" == "$GIT_SPLITSH_SHA1" ]; then
        echo "Test #5 - OK ($GIT_SUBTREE_SPLIT_SHA1 == $GIT_SPLITSH_SHA1)"
    else
        echo "Test #5 - NOT OK ($GIT_SUBTREE_SPLIT_SHA1 != $GIT_SPLITSH_SHA1)"
        exit 1
    fi

    cd ../
}

LITE_PATH=`pwd`/splitsh-lite
if [ ! -e $LITE_PATH ]; then
    echo "You first need to compile the splitsh-lite binary"
    exit 1
fi

if [ ! -d splitter-lite-tests ]; then
    mkdir splitter-lite-tests
fi
cd splitter-lite-tests

simpleTest
mergeTest
twigSplitTest
