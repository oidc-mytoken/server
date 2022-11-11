mkdir ../shared
first=$(grep '^## ' -nm1 CHANGELOG.md | cut -d':' -f1); \
  second=$(grep '^## ' -nm2 CHANGELOG.md | tail -n1 | cut -d':' -f1); \
  tail -n+$first CHANGELOG.md | head -n$(($second-$first)) > ../shared/release.md
GORELEASER_CONFIG=".goreleaser.yml"
if [ -n "$CI_COMMIT_TAG" ] && echo "$CI_COMMIT_TAG" | grep -qv '-'; then
GORELEASER_CONFIG=".goreleaser-release.yml"
fi
BASEDIR=/go/src/github.com/oidc-mytoken/server
docker run --rm --privileged \
  -v "$PWD":"$BASEDIR" \
  -w "$BASEDIR" \
  -v "${PWD}/../shared":/tmp/shared \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -e DOCKER_USERNAME -e DOCKER_PASSWORD \
  -e GITHUB_TOKEN \
  -e GORELEASER_CONFIG \
  goreleaser/goreleaser release -f $GORELEASER_CONFIG --release-notes /tmp/shared/release.md
ls -l results