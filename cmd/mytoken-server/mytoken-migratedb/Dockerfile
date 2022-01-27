FROM oidcmytoken/debian-wait-for:latest
WORKDIR /mytoken
COPY mytoken-migratedb /usr/bin/mytoken-migratedb
RUN apt-get update && \
    apt-get install mariadb-client -y && \
    apt-get autoremove -y && \
    apt-get clean -y && \
    rm -rf /var/lib/apt/lists/*
ENTRYPOINT ["/opt/mytoken/scripts/run.sh"]
CMD ["mytoken-migratedb"]
