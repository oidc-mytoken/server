#!/bin/bash

MYTOKEN_SERVER=https://mytoken.data.kit.edu
GROUP_USER=_
DATA_DIR="$(dirname "$0")"

while [ $# -gt 0 ]; do
    case "$1" in
    -s|--server|--mytoken-server)   MYTOKEN_SERVER=$2;           shift;;
    -u|--user)                      GROUP_USER=$2;               shift;;
    -p|--pw|--password)             GROUP_PW=$2;                 shift;;
    -d|--data|--data-dir|--dir)     DATA_DIR=$2;                 shift;;
    esac
    shift
done

[ -z "${GROUP_PW}" ] && {
    echo "ERROR: GROUP_PW is not set."
    exit 1
}

echo "DATA_DIR: ${DATA_DIR}"
echo "MYTOKEN_SERVER: ${MYTOKEN_SERVER}"
echo "GROUP_USER: ${GROUP_USER}"

PROFILE_ENDPOINT=$(curl "${MYTOKEN_SERVER}/.well-known/mytoken-configuration" 2>/dev/null| jq -r .profiles_endpoint)
GROUP_ENDPOINT="${PROFILE_ENDPOINT}/${GROUP_USER}"

echo "PROFILE_ENDPOINT: ${PROFILE_ENDPOINT}"
echo "GROUP_ENDPOINT: ${GROUP_ENDPOINT}"
echo "-----"
echo

for DIR in $(find "${DATA_DIR}" -type d | tail -n+2); do
  DIR=$(basename "$DIR")
  res=$(curl "${GROUP_ENDPOINT}/${DIR}" 2>/dev/null)
  declare -A mapping
  for row in $(echo "${res}" | jq -r '.[] | @base64'); do
    _jq() {
      echo ${row} | base64 -d | jq -r ${1}
      }
    name=$(_jq '.name')
    id=$(_jq '.id')
    mapping[$name]=$id
  done
  for f in "${DATA_DIR}/${DIR}"/*; do
    f=$(basename "$f")
    if [ -z "${mapping[$f]}" ]; then
      echo "$DIR/$f does not exist - create it"
      curl -X POST -u "${GROUP_USER}:${GROUP_PW}" "${GROUP_ENDPOINT}/${DIR}" -H "Content-Type: application/json" \
        -d"{\"name\":\"$f\",\"payload\":$(cat "${DATA_DIR}/${DIR}/${f}")}" &>/dev/null
    else
      echo "$DIR/$f already exists - update it"
      curl -X PUT -u "${GROUP_USER}:${GROUP_PW}" "${GROUP_ENDPOINT}/${DIR}/${mapping[$f]}" \
        -H "Content-Type: application/json" \
        -d"{\"name\":\"$f\",\"payload\":$(cat "${DATA_DIR}/${DIR}/${f}")}" &>/dev/null
    fi
  done
done
