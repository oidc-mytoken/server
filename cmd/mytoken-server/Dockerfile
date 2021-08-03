FROM oidcmytoken/debian-wait-for:latest
WORKDIR /mytoken
COPY mytoken-server /usr/bin/mytoken-server
ENTRYPOINT ["/opt/mytoken/scripts/run.sh"]
CMD ["mytoken-server"]
