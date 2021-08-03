FROM oidcmytoken/debian-wait-for:latest
WORKDIR /mytoken
COPY mytoken-migratedb /usr/bin/mytoken-migratedb
ENTRYPOINT ["/opt/mytoken/scripts/run.sh"]
CMD ["mytoken-migratedb"]
