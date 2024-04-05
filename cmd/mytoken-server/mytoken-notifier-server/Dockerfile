FROM debian:stable
WORKDIR /mytoken
COPY mytoken-notifier-server /usr/bin/mytoken-notifier-server
USER 1000:1000
ENTRYPOINT ["mytoken-notifier-server"]
