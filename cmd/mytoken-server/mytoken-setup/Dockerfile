FROM debian:stable
RUN apt-get update && \
    apt-get install ca-certificates -y && \
    apt-get autoremove -y && \
    apt-get clean -y && \
    rm -rf /var/lib/apt/lists/*
WORKDIR /mytoken
COPY mytoken-setup /usr/bin/mytoken-setup
ENTRYPOINT ["mytoken-setup"]
