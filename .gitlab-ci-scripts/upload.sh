
REPO_TARGET="/prerel"
if [ -n "$CI_COMMIT_TAG" ] && echo "$CI_COMMIT_TAG" | grep -qv '~'; then
  REPO_TARGET="/preprod"
fi

# ssh-key-script
[ -e /tmp/ssh-private-keys/${REPO_USER} ] && {
  eval $(ssh-agent -s)
  cat /tmp/ssh-private-keys/${REPO_USER} | tr -d '\r' | ssh-add -
  test -d ~/.ssh || mkdir -p ~/.ssh
  chmod 700 ~/.ssh
}
[ -e /tmp/ssh-private-keys/known_hosts ] && {
  test -d ~/.ssh || mkdir -p ~/.ssh
  cp /tmp/ssh-private-keys/known_hosts ~/.ssh/known_hosts
  chmod 644 ~/.ssh/known_hosts
}
ssh-add -l
ssh -o StrictHostKeyChecking=no "${REPO_USER}@${REPO_HOST}" "hostname -f"

# sign-repo function
sign_repos() {
    ssh "${REPO_USER}@${REPO_HOST}" "~/ci-voodoo/ci-tools/sign-all-repos.sh -t ${REPO_TARGET}"
}

upload_files() {
  UPLOAD_DIR=/tmp/package-upload
  ssh "${REPO_USER}@${REPO_HOST}" "rm -rf $UPLOAD_DIR && mkdir -p $UPLOAD_DIR"
  scp results/* "${REPO_USER}@${REPO_HOST}:${UPLOAD_DIR}"
}

distribute_files() {
    ssh "${REPO_USER}@${REPO_HOST}" "~/ci-voodoo/ci-tools/distribute-local-packages.sh -t ${REPO_TARGET} -w mytoken"
}


# upload and sign
upload_files
distribute_files
sign_repos
