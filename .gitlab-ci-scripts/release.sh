#!/bin/sh

set -e
if [ -z "$CI_COMMIT_TAG" ]; then .gitlab-ci-scripts/set-prerel-version.sh; fi;
echo $DOCKER_USERNAME
echo $DOCKER_PASSWORD | docker login -u $DOCKER_USER --password-stdin $DOCKER_REGISTRY
.gitlab-ci-scripts/goreleaser.sh
.gitlab-ci-scripts/upload.sh

curl -d "repo=github.com/oidc-mytoken/server" https://goreportcard.com/checks
