#!/bin/sh

set -e
if [ -z "$CI_COMMIT_TAG" ]; then .gitlab-ci-scripts/set-prerel-version.sh; fi;
docker login -u $DOCKER_USER -p $DOCKER_PASSWORD
.gitlab-ci-scripts/goreleaser.sh
.gitlab-ci-scripts/upload.sh

curl -d "repo=github.com/oidc-mytoken/server" https://goreportcard.com/checks
