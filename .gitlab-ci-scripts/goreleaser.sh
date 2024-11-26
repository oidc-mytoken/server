#!/bin/bash

mkdir ../shared
first=$(grep '^## ' -nm1 CHANGELOG.md | cut -d':' -f1); \
  second=$(grep '^## ' -nm2 CHANGELOG.md | tail -n1 | cut -d':' -f1); \
  tail -n+$first CHANGELOG.md | head -n$(($second-$first)) > ../shared/release.md
GORELEASER_CONFIG=".goreleaser.yml"
if [ -n "$CI_COMMIT_TAG" ] && echo "$CI_COMMIT_TAG" | grep -qv '~'; then
GORELEASER_CONFIG=".goreleaser-release.yml"
fi
GORELEASER_OPTIONS=""
[[ "${CI_COMMIT_BRANCH}" != "${CI_DEFAULT_BRANCH}" ]] && {
    [[ "${CI_COMMIT_BRANCH}" != "${PREREL_BRANCH_NAME}" ]] && {
        # we're on devel
        GORELEASER_OPTIONS="--skip docker"
    }
}

goreleaser release -f $GORELEASER_CONFIG --release-notes ../shared/release.md --verbose ${GORELEASER_OPTIONS}
ls -l results