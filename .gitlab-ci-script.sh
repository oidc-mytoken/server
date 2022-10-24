set -x
mkdir ../shared
first=$(grep '^## ' -nm1 CHANGELOG.md | cut -d':' -f1); \
  second=$(grep '^## ' -nm2 CHANGELOG.md | tail -n1 | cut -d':' -f1); \
  tail -n+$first CHANGELOG.md | head -n$(($second-$first)) > ../shared/release.md
export GORELEASER_CONFIG="$(if echo $CI_COMMIT_TAG | grep -q '-'; then echo '.goreleaser.yml'; else echo '.goreleaser-release.yml'; fi)"
BASEDIR=/go/src/github.com/oidc-mytoken/server
ls -l "${PWD}/../shared"
docker run --rm --privileged \
  -v $PWD:$BASEDIR \
  -w  $BASEDIR\
  -v "${PWD}/../shared":/tmp/shared \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -e DOCKER_USERNAME -e DOCKER_PASSWORD \
  -e GITHUB_TOKEN \
  -e GORELEASER_CONFIG \
  goreleaser/goreleaser release -f $GORELEASER_CONFIG --release-notes /tmp/shared/release.md
