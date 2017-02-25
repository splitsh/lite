#!/usr/bin/env bash

set -e

if [ $# -eq 0 ]; then
	set -- -h
fi
OPTS_SPEC="\
git splitsh init    url
git splitsh publish splits --heads=<heads> --tags=<tags> --splits=<splits>
git splitsh update
--
h,help        show the help
q             quiet
debug         show plenty of debug output
n,dry-run     do everything except actually send the updates
work-dir      directory that contains the working directory

 options for 'publish'
heads=        only publish for listed heads instead of all heads
no-heads      do not publish any heads
tags=         only publish for listed tags instead of all tags
no-tags       do not publish any tags
update        fetch updates from repository before publishing
rebuild-tags  rebuild all tags (as opposed to skipping tags that are already synced)
"
eval "$(echo "$OPTS_SPEC" | git rev-parse --parseopt -- "$@" || echo exit $?)"

DEBUG="  :DEBUG >"
PATH=$PATH:$(git --exec-path)

die()
{
    printf >&2 '%s\n' "$*"
    exit 1
}

if [ "$(hash splitsh-lite &>/dev/null && echo OK)" = "" ]; then
	die "Git subplit needs git splitsh-lite in the \$PATH; install it (https://github.com/splitsh/lite)"
fi

QUIET=
COMMAND=
SPLITS=
REPO_URL=
WORK_DIR="${PWD}"
HEADS=
NO_HEADS=
TAGS=
NO_TAGS=
REBUILD_TAGS=
DRY_RUN=
VERBOSE=

splitsh_main()
{
	while [ $# -gt 0 ]; do
		opt="$1"
		shift
		case "$opt" in
			-q) QUIET=1 ;;
			--debug) VERBOSE=1 ;;
			--work-dir) WORK_DIR="$1"; shift ;;
			--heads) HEADS="$1"; shift ;;
			--no-heads) NO_HEADS=1 ;;
			--tags) TAGS="$1"; shift ;;
			--no-tags) NO_TAGS=1 ;;
			--update) UPDATE=1 ;;
			-n) DRY_RUN="--dry-run" ;;
			--dry-run) DRY_RUN="--dry-run" ;;
			--rebuild-tags) REBUILD_TAGS=1 ;;
			--) break ;;
			*) die "Unexpected option: $opt" ;;
		esac
	done

	COMMAND="$1"
	shift

	case "$COMMAND" in
		init)
			if [ $# -lt 1 ]; then die "init command requires URL to be passed as first argument"; fi
			REPO_URL="$1"
			shift
			splitsh_init
			;;
		publish)
			if [ $# -lt 1 ]; then die "publish command requires splits to be passed as first argument"; fi
			SPLITS="$1"
			shift
			splitsh_publish
			;;
		update)
			splitsh_update
			;;
		*) die "Unknown command '$COMMAND'" ;;
	esac
}

say()
{
	if [ -z "$QUIET" ]; then
		echo "$@" >&2
	fi
}

execute()
{
	if [ -n "$VERBOSE" ]; then
		echo "${DEBUG} $@"
	fi

	`$@`
}

splitsh_init()
{
	say "Initializing splitsh from origin (${REPO_URL})"
	execute git clone --bare -q "$REPO_URL" "$WORK_DIR" || die "Could not clone repository"
}

splitsh_publish()
{
	if [ -n "$UPDATE" ]; then
		splitsh_update
	fi

	if [ -z "$HEADS" ] && [ -z "$NO_HEADS" ]; then
		# If heads are not specified and we want heads, discover them.
		HEADS="$(git ls-remote origin 2>/dev/null | grep "refs/heads/" | cut -f3- -d/)"

		if [ -n "$VERBOSE" ]; then
			echo "${DEBUG} HEADS=\"${HEADS}\""
		fi
	fi

	if [ -z "$TAGS" ] && [ -z "$NO_TAGS" ]; then
		# If tags are not specified and we want tags, discover them.
		TAGS="$(git ls-remote origin 2>/dev/null | grep -v "\^{}" | grep "refs/tags/" | cut -f3 -d/)"

		if [ -n "$VERBOSE" ]; then
			echo "${DEBUG} TAGS=\"${TAGS}\""
		fi
	fi

	for SPLIT in $SPLITS; do
		SUBPATH=$(echo "$SPLIT" | cut -f1 -d:)
		REMOTE_URL=$(echo "$SPLIT" | cut -f2- -d:)
		REMOTE_NAME=$(echo "$SPLIT" | git hash-object --stdin)

		if [ -n "$VERBOSE" ]; then
			echo "${DEBUG} SUBPATH=${SUBPATH}"
			echo "${DEBUG} REMOTE_URL=${REMOTE_URL}"
			echo "${DEBUG} REMOTE_NAME=${REMOTE_NAME}"
		fi

		if ! git remote | grep "^${REMOTE_NAME}$" >/dev/null; then
			execute git remote add "$REMOTE_NAME" "$REMOTE_URL"
		fi

		say "Syncing ${SUBPATH} -> ${REMOTE_URL}"

		for HEAD in $HEADS; do
			if ! execute git show-ref --quiet --verify -- "refs/heads/${HEAD}"; then
				say " - skipping head '${HEAD}' (does not exist)"
				continue
			fi

			say " - syncing branch '${HEAD}'"

            SPLIT_CMD="splitsh-lite --quiet --prefix="$SUBPATH" --origin="heads/${HEAD}" >/dev/null"
			if [ -n "$VERBOSE" ]; then
				echo "${DEBUG} $SPLIT_CMD"
			fi

			SHA1=`$SPLIT_CMD`
			if [ $? -eq 0 ]; then
				PUSH_CMD="git push -q ${DRY_RUN} --force $REMOTE_NAME ${SHA1}:refs/heads/${HEAD}"

				if [ -n "$VERBOSE" ]; then
					echo "${DEBUG} $PUSH_CMD"
				fi

				if [ -n "$DRY_RUN" ]; then
					echo \# $PUSH_CMD
                else
    				$PUSH_CMD
				fi
			fi
		done

        REMOTE_TAGS="$(git ls-remote $REMOTE_NAME 2>/dev/null | grep -v "\^{}" | grep "refs/tags/" | cut -f3 -d/)"
		for TAG in $TAGS; do
			if ! execute git show-ref --quiet --verify -- "refs/tags/${TAG}"; then
				say " - skipping tag '${TAG}' (does not exist)"
				continue
			fi

			if [[ $REMOTE_TAGS == *"$TAG"* ]] && [ -z "$REBUILD_TAGS" ]; then
				say " - skipping tag '${TAG}' (already synced)"
				continue
			fi

			say " - syncing tag '${TAG}'"

            SPLIT_CMD="splitsh-lite --quiet --prefix="$SUBPATH" --origin="tags/${TAG}" >/dev/null"
			if [ -n "$VERBOSE" ]; then
				echo "${DEBUG} $SPLIT_CMD"
			fi

			SHA1=`$SPLIT_CMD`
			if [ $? -eq 0 ]; then
				PUSH_CMD="git push -q ${DRY_RUN} --force ${REMOTE_NAME} ${SHA1}:refs/tags/${TAG}"

				if [ -n "$VERBOSE" ]; then
					echo "${DEBUG} PUSH_CMD=\"${PUSH_CMD}\""
				fi

				if [ -n "$DRY_RUN" ]; then
					echo \# $PUSH_CMD
                else
    				$PUSH_CMD
				fi
			fi
		done
	done
}

splitsh_update()
{
	say "Updating splitsh from origin"
	execute "git fetch -q -t origin"
}

splitsh_main "$@"
