# Tables
CREATE TABLE IF NOT EXISTS ProfileTypes
(
    id                    INT UNSIGNED AUTO_INCREMENT
        PRIMARY KEY, type VARCHAR(100) NOT NULL, CONSTRAINT ProfileTypes_UN
        UNIQUE (type)
);

CREATE TABLE IF NOT EXISTS ServerProfiles
(
    id                    VARCHAR(128)                         NOT NULL
        PRIMARY KEY,
    type                  INT UNSIGNED                         NOT NULL,
    `group`               VARCHAR(64)                          NOT NULL,
    name                  VARCHAR(128)                         NOT NULL,
    payload               LONGTEXT COLLATE utf8mb4_bin         NOT NULL,
    created               DATETIME DEFAULT CURRENT_TIMESTAMP() NOT NULL,
    CONSTRAINT ServerProfiles_UN
        UNIQUE (type, `group`, name),
    CONSTRAINT ServerProfiles_FK
        FOREIGN KEY (type) REFERENCES ProfileTypes (id)
            ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT payload
        CHECK (JSON_VALID(`payload`))
);

# Table Data
INSERT IGNORE INTO ProfileTypes (type)
    VALUES ('profile');
INSERT IGNORE INTO ProfileTypes (type)
    VALUES ('restrictions');
INSERT IGNORE INTO ProfileTypes (type)
    VALUES ('capabilities');
INSERT IGNORE INTO ProfileTypes (type)
    VALUES ('rotation');

INSERT IGNORE INTO Events (event)
    VALUES ('revoked_other_token');
INSERT IGNORE INTO Events (event)
    VALUES ('tokeninfo_history_other_token');

# Procedures
DELIMITER ;;
CREATE OR REPLACE PROCEDURE MTokens_GetForUser(IN UID BIGINT UNSIGNED)
BEGIN
    SELECT id, parent_id, id AS mom_id, name, created, ip_created AS ip
        FROM MTokens
        WHERE user_id = UID
        ORDER BY created;
END;;

CREATE OR REPLACE PROCEDURE MTokens_GetSubtokens(IN MTID VARCHAR(128))
BEGIN
    CREATE TEMPORARY TABLE IF NOT EXISTS effected_MTIDs (id VARCHAR(128));
    TRUNCATE effected_MTIDs;
    INSERT INTO effected_MTIDs
    WITH RECURSIVE childs AS (SELECT id, parent_id
                                  FROM MTokens
                                  WHERE id = MTID
                              UNION ALL
                              SELECT mt.id, mt.parent_id
                                  FROM MTokens mt
                                           INNER JOIN childs c
                                  WHERE mt.parent_id = c.id)
    SELECT id
        FROM childs;
    SELECT m.id, m.parent_id, m.id AS mom_id, m.name, m.created, m.ip_created AS ip
        FROM MTokens m
        WHERE m.id IN
              (SELECT id
                   FROM effected_MTIDs);
    DROP TABLE effected_MTIDs;
END;;

CREATE OR REPLACE PROCEDURE Profiles_GetGroups()
BEGIN
    SELECT DISTINCT `group` FROM ServerProfiles;
END;;

CREATE OR REPLACE PROCEDURE GetProfileTemplate(IN TYPE_ VARCHAR(100), IN GROUP_ VARCHAR(64))
BEGIN
    SELECT id, `group`, name, payload
        FROM ServerProfiles sp
        WHERE sp.`group` = GROUP_
          AND sp.`type` = (SELECT id
                               FROM ProfileTypes pt
                               WHERE pt.`type` = TYPE_);
END;;
CREATE OR REPLACE PROCEDURE Profiles_GetProfiles(IN GROUP_ VARCHAR(64))
BEGIN
    CALL GetProfileTemplate('profile', GROUP_);
END;;
CREATE OR REPLACE PROCEDURE Profiles_GetRestrictions(IN GROUP_ VARCHAR(64))
BEGIN
    CALL GetProfileTemplate('restrictions', GROUP_);
END;;
CREATE OR REPLACE PROCEDURE Profiles_GetCapabilities(IN GROUP_ VARCHAR(64))
BEGIN
    CALL GetProfileTemplate('capabilities', GROUP_);
END;;
CREATE OR REPLACE PROCEDURE Profiles_GetRotations(IN GROUP_ VARCHAR(64))
BEGIN
    CALL GetProfileTemplate('rotation', GROUP_);
END;;

CREATE OR REPLACE PROCEDURE DeleteProfileTemplate(IN TYPE_ VARCHAR(100), IN ID_ VARCHAR(128), IN GROUP_ VARCHAR(64))
BEGIN
    DELETE
        FROM ServerProfiles
        WHERE `group` = GROUP_
          AND ID = ID_
          AND `type` = (SELECT pt.id
                            FROM ProfileTypes pt
                            WHERE pt.`type` = TYPE_);
END;;
CREATE OR REPLACE PROCEDURE Profiles_DeleteProfiles(IN ID_ VARCHAR(128), IN GROUP_ VARCHAR(64))
BEGIN
    CALL DeleteProfileTemplate('profile', ID_, GROUP_);
END;;
CREATE OR REPLACE PROCEDURE Profiles_DeleteRestrictions(IN ID_ VARCHAR(128), IN GROUP_ VARCHAR(64))
BEGIN
    CALL DeleteProfileTemplate('restrictions', ID_, GROUP_);
END;;
CREATE OR REPLACE PROCEDURE Profiles_DeleteCapabilities(IN ID_ VARCHAR(128), IN GROUP_ VARCHAR(64))
BEGIN
    CALL DeleteProfileTemplate('capabilities', ID_, GROUP_);
END;;
CREATE OR REPLACE PROCEDURE Profiles_DeleteRotations(IN ID_ VARCHAR(128), IN GROUP_ VARCHAR(64))
BEGIN
    CALL DeleteProfileTemplate('rotation', ID_, GROUP_);
END;;

CREATE OR REPLACE PROCEDURE InsertProfileTemplate(IN TYPE_ VARCHAR(100), IN ID_ VARCHAR(128), IN GROUP_ VARCHAR(64), IN
    NAME_ VARCHAR(128), IN PAYLOAD_ LONGTEXT)
BEGIN
    INSERT INTO ServerProfiles (id, type, `group`, name, payload)
        VALUES (ID_, (SELECT id
                          FROM ProfileTypes pt
                          WHERE pt.`type` = TYPE_), GROUP_, NAME_, PAYLOAD_);
END;;
CREATE OR REPLACE PROCEDURE Profiles_InsertProfiles(IN ID_ VARCHAR(128), IN GROUP_ VARCHAR(64), IN NAME_ VARCHAR(128)
, IN PAYLOAD_ LONGTEXT)
BEGIN
    CALL InsertProfileTemplate('profile', ID_, GROUP_, NAME_, PAYLOAD_);
END;;
CREATE OR REPLACE PROCEDURE Profiles_InsertRestrictions(IN ID_ VARCHAR(128), IN GROUP_ VARCHAR(64)
, IN NAME_ VARCHAR(128)
, IN PAYLOAD_ LONGTEXT)
BEGIN
    CALL InsertProfileTemplate('restrictions', ID_, GROUP_, NAME_, PAYLOAD_);
END;;
CREATE OR REPLACE PROCEDURE Profiles_InsertCapabilities(IN ID_ VARCHAR(128), IN GROUP_ VARCHAR(64)
, IN NAME_ VARCHAR(128)
, IN PAYLOAD_ LONGTEXT)
BEGIN
    CALL InsertProfileTemplate('capabilities', ID_, GROUP_, NAME_, PAYLOAD_);
END;;
CREATE OR REPLACE PROCEDURE Profiles_InsertRotations(IN ID_ VARCHAR(128), IN GROUP_ VARCHAR(64), IN NAME_ VARCHAR(128)
, IN PAYLOAD_ LONGTEXT)
BEGIN
    CALL InsertProfileTemplate('rotation', ID_, GROUP_, NAME_, PAYLOAD_);
END;;

CREATE OR REPLACE PROCEDURE UpdateProfileTemplate(IN TYPE_ VARCHAR(100), IN ID_ VARCHAR(128), IN GROUP_ VARCHAR(64), IN
    NAME_ VARCHAR(128), IN PAYLOAD_ LONGTEXT)
BEGIN
    UPDATE ServerProfiles sp
    SET sp.name   = NAME_,
        sp.payload=PAYLOAD_
        WHERE sp.type = (SELECT id
                             FROM ProfileTypes pt
                             WHERE pt.`type` = TYPE_)
          AND sp.id = ID_
          AND sp.group = GROUP_;
END;;
CREATE OR REPLACE PROCEDURE Profiles_UpdateProfiles(IN ID_ VARCHAR(128), IN GROUP_ VARCHAR(64), IN NAME_ VARCHAR(128)
, IN PAYLOAD_ LONGTEXT)
BEGIN
    CALL UpdateProfileTemplate('profile', ID_, GROUP_, NAME_, PAYLOAD_);
END;;
CREATE OR REPLACE PROCEDURE Profiles_UpdateRestrictions(IN ID_ VARCHAR(128), IN GROUP_ VARCHAR(64)
, IN NAME_ VARCHAR(128)
, IN PAYLOAD_ LONGTEXT)
BEGIN
    CALL UpdateProfileTemplate('restrictions', ID_, GROUP_, NAME_, PAYLOAD_);
END;;
CREATE OR REPLACE PROCEDURE Profiles_UpdateCapabilities(IN ID_ VARCHAR(128), IN GROUP_ VARCHAR(64)
, IN NAME_ VARCHAR(128)
, IN PAYLOAD_ LONGTEXT)
BEGIN
    CALL UpdateProfileTemplate('capabilities', ID_, GROUP_, NAME_, PAYLOAD_);
END;;
CREATE OR REPLACE PROCEDURE Profiles_UpdateRotations(IN ID_ VARCHAR(128), IN GROUP_ VARCHAR(64), IN NAME_ VARCHAR(128)
, IN PAYLOAD_ LONGTEXT)
BEGIN
    CALL UpdateProfileTemplate('rotation', ID_, GROUP_, NAME_, PAYLOAD_);
END;;

CREATE OR REPLACE PROCEDURE Cleanup_MTokens()
BEGIN
    DELETE FROM MTokens WHERE DATE_ADD(expires_at, INTERVAL 1 MONTH) < CURRENT_TIMESTAMP();
END;;
CREATE OR REPLACE PROCEDURE Cleanup_ProxyTokens()
BEGIN
    DELETE
        FROM ProxyTokens
        WHERE id IN (SELECT id
                         FROM TransferCodesAttributes
                         WHERE DATE_ADD(expires_at, INTERVAL 1 MONTH) < CURRENT_TIMESTAMP());
END;;

CREATE OR REPLACE PROCEDURE MTokens_GetForUser(IN UID BIGINT UNSIGNED)
BEGIN
    SELECT id, parent_id, id AS mom_id, name, created, expires_at, ip_created AS ip
        FROM MTokens
        WHERE user_id = UID
        ORDER BY created;
END;;
CREATE OR REPLACE PROCEDURE MTokens_GetSubtokens(IN MTID VARCHAR(128))
BEGIN
    CREATE TEMPORARY TABLE IF NOT EXISTS effected_MTIDs (id VARCHAR(128));
    TRUNCATE effected_MTIDs;
    INSERT INTO effected_MTIDs
    WITH RECURSIVE childs AS (SELECT id, parent_id
                                  FROM MTokens
                                  WHERE id = MTID
                              UNION ALL
                              SELECT mt.id, mt.parent_id
                                  FROM MTokens mt
                                           INNER JOIN childs c
                                  WHERE mt.parent_id = c.id)
    SELECT id
        FROM childs;
    SELECT m.id, m.parent_id, m.id AS mom_id, m.name, m.created, m.expires_at, m.ip_created AS ip
        FROM MTokens m
        WHERE m.id IN
              (SELECT id
                   FROM effected_MTIDs);
    DROP TABLE effected_MTIDs;
END;;

DELIMITER ;

