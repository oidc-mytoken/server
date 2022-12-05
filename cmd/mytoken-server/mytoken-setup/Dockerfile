FROM oidcmytoken/debian-wait-for:latest
WORKDIR /mytoken
COPY mytoken-setup /usr/bin/mytoken-setup
RUN apt-get update && \
    apt-get install mariadb-client -y && \
    apt-get autoremove -y && \
    apt-get clean -y && \
    rm -rf /var/lib/apt/lists/*
USER 1000:1000
ENTRYPOINT ["mytoken-setup"]
