#!/bin/sh

DEVSTRING="pr"
VERSION_FILE=internal/model/version/VERSION

while [ $# -gt 0 ]; do
  case $1 in
    --devstring)
      DEVSTRING="$2"
      shift # past argument
      shift # past value
      ;;
    --version_file)
      VERSION_FILE="$2"
      shift # past argument
      shift # past value
      ;;
    --*|-*)
      echo "Unknown option $1"
      exit 1
      ;;
  esac
done

git config user.email

echo "CI: $CI"

[ "x${CI}" = "xtrue" ] && {
    echo "Setting up git in CI"
    git config --global --add safe.directory "$PWD"
    git config user.email "ci@repo.data.kit.edu"
    git config user.name "cicd"
}

# Get master branch name:
#   use origin if exists
#   else use last found remote
REMOTES=$(git remote show)
for R in $REMOTES; do
    MASTER=master
    MASTER_BRANCH="refs/remotes/${R}/${MASTER}"
    #echo "Master-branch: ${MASTER_BRANCH}"
    [ "x${R}" = "xorigin" ] && break
done

PREREL=$(git rev-list --count HEAD ^"$MASTER_BRANCH")

# use version file:
VERSION=$(cat "$VERSION_FILE")
PR_VERSION="${VERSION}-${DEVSTRING}${PREREL}"
echo "$PR_VERSION" > "$VERSION_FILE"
echo "$PR_VERSION"

echo "$PR_VERSION" > "$VERSION_FILE"
git add "$VERSION_FILE"
git commit -m "dummy prerel version"
git tag "v${PR_VERSION}"
