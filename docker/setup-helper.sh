#!/bin/bash

DIR="${MYTOKEN_DOCKER_DIR:-$HOME/mytoken-docker}"
CONFIG="${MYTOKEN_DOCKER_CONFIG:-docker-config.yaml}"

# Create base dir and log dir
mkdir -p $DIR/logs

# Create mount points for db data
mkdir -p $DIR/db/b1 $DIR/db/1 $DIR/db/2 $DIR/db/3
sudo chown -R 1001:1001 $DIR/db

# Install geo-location-db
docker run --rm -v$CONFIG:/etc/mytoken/config.yaml -v$DIR:/root/mytoken/ oidcmytoken/mytoken-setup install geoip-db

# Create signing key for mytokens
docker run --rm -v$CONFIG:/etc/mytoken/config.yaml -v$DIR:/root/mytoken/ oidcmytoken/mytoken-setup signing-key

# Create a self-signed certificate for https
# The key and crt must be in a single .pem file
openssl req -new -x509 -nodes -days 365 -newkey rsa:2048 -subj "/C=DE/ST=/L=/O=/CN=localhost" -keyout $DIR/localhost.key -out $DIR/localhost.pem
cat $DIR/localhost.pem $DIR/localhost.key >$DIR/localhost.crt.pem

# Create password files for different db users
pwgen -nsBc 30 1 >$DIR/db_password
pwgen -nsBc 30 1 >$DIR/db_root_password
pwgen -nsBc 30 1 >$DIR/db_replication_password
pwgen -nsBc 30 1 >$DIR/db_backup_password

# Create ssh host keys
ssh-keygen -f $DIR/ssh_host_rsa_key -t rsa -N ""
ssh-keygen -f $DIR/ssh_host_ecdsa_key -t ecdsa -N ""
ssh-keygen -f $DIR/ssh_host_ed25519_key -t ed25519 -N ""
