FROM oidcmytoken/debian-wait-for:latest
WORKDIR /mytoken
COPY mytoken-setup /usr/bin/mytoken-setup
USER 1000:1000
ENTRYPOINT ["mytoken-setup"]
