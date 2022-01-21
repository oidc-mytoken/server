# noinspection SqlResolveForFile

# Tables
DROP VIEW IF EXISTS MyTokens;

ALTER TABLE MTokens
    DROP FOREIGN KEY Mytokens_FK_1;
ALTER TABLE MTokens
    DROP COLUMN root_id;

ALTER TABLE Users
    DROP COLUMN token_tracing;
ALTER TABLE Users
    DROP COLUMN jwt_pk;

ALTER TABLE TransferCodesAttributes
    ADD ssh_key_fp VARCHAR(128) NULL;

TRUNCATE TABLE AuthInfo;
ALTER TABLE AuthInfo
    ADD request_json JSON NOT NULL;
ALTER TABLE AuthInfo
    DROP COLUMN iss;
ALTER TABLE AuthInfo
    DROP COLUMN restrictions;
ALTER TABLE AuthInfo
    DROP COLUMN capabilities;
ALTER TABLE AuthInfo
    DROP COLUMN name;
ALTER TABLE AuthInfo
    DROP COLUMN subtoken_capabilities;
ALTER TABLE AuthInfo
    DROP COLUMN rotation;
ALTER TABLE AuthInfo
    DROP COLUMN response_type;
ALTER TABLE AuthInfo
    DROP COLUMN max_token_len;

# CryptStore
CREATE TABLE `CryptPayloadTypes`
(
    `id`           INT(10) UNSIGNED NOT NULL AUTO_INCREMENT,
    `payload_type` VARCHAR(128)     NOT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `CryptPayloadTypes_UN` (`payload_type`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 4
  DEFAULT CHARSET = utf8mb4;

RENAME TABLE RefreshTokens TO CryptStore;
ALTER TABLE CryptStore
    CHANGE rt crypt TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL;
ALTER TABLE CryptStore
    ADD COLUMN IF NOT EXISTS payload_type INT UNSIGNED NULL;
ALTER TABLE CryptStore
    ADD CONSTRAINT CryptStore_FK FOREIGN KEY (payload_type) REFERENCES CryptPayloadTypes (id) ON UPDATE CASCADE;
UPDATE CryptStore
SET payload_type=(SELECT id FROM CryptPayloadTypes WHERE payload_type = 'RT')
    WHERE payload_type IS NULL;
ALTER TABLE CryptStore
    MODIFY COLUMN payload_type INT UNSIGNED NOT NULL;

INSERT INTO CryptStore (crypt, payload_type, created, updated)
SELECT *
    FROM (SELECT at.token AS crypt, at.created, at.created AS updated FROM AccessTokens at) jwts
             JOIN (SELECT cpt.id AS payload_type FROM CryptPayloadTypes cpt WHERE cpt.payload_type = 'AT') p;
ALTER TABLE AccessTokens
    DROP COLUMN IF EXISTS created;
ALTER TABLE AccessTokens
    ADD COLUMN IF NOT EXISTS token_crypt BIGINT UNSIGNED NULL;
ALTER TABLE AccessTokens
    ADD CONSTRAINT AccessTokens_FK_1 FOREIGN KEY (token_crypt) REFERENCES CryptStore (id) ON UPDATE CASCADE;
UPDATE AccessTokens a
SET a.token_crypt=(SELECT c.id FROM CryptStore c WHERE c.crypt = a.token)
    WHERE token_crypt IS NULL;
ALTER TABLE AccessTokens
    MODIFY COLUMN token_crypt BIGINT UNSIGNED NOT NULL;
ALTER TABLE AccessTokens
    DROP COLUMN IF EXISTS token;

DELETE
    FROM ProxyTokens
    WHERE jwt = '';
INSERT INTO CryptStore (crypt, payload_type, created, updated)
SELECT *
    FROM (SELECT pt.jwt AS crypt, pt.created, pt.created AS updated FROM TransferCodes pt) jwts
             JOIN (SELECT cpt.id AS payload_type FROM CryptPayloadTypes cpt WHERE cpt.payload_type = 'MT') p;
ALTER TABLE ProxyTokens
    ADD COLUMN IF NOT EXISTS jwt_crypt BIGINT UNSIGNED NULL;
ALTER TABLE ProxyTokens
    ADD CONSTRAINT ProxyTokens_FK_1 FOREIGN KEY (jwt_crypt) REFERENCES CryptStore (id) ON UPDATE CASCADE;
UPDATE ProxyTokens p
SET jwt_crypt=(SELECT id FROM CryptStore c WHERE c.crypt = p.jwt)
    WHERE jwt_crypt IS NULL;
ALTER TABLE ProxyTokens
    MODIFY COLUMN jwt_crypt BIGINT UNSIGNED NOT NULL;
ALTER TABLE ProxyTokens
    DROP COLUMN IF EXISTS jwt;

CREATE OR REPLACE VIEW TransferCodes AS
SELECT `pt`.`id`                AS `id`,
       `cs`.`crypt`             AS `jwt`,
       `tca`.`created`          AS `created`,
       `tca`.`expires_in`       AS `expires_in`,
       `tca`.`expires_at`       AS `expires_at`,
       `tca`.`revoke_MT`        AS `revoke_MT`,
       `tca`.`response_type`    AS `response_type`,
       `tca`.`max_token_len`    AS `max_token_len`,
       `tca`.`consent_declined` AS `consent_declined`,
       `tca`.ssh_key_fp         AS `ssh_key_fp`
    FROM ((`ProxyTokens` `pt` JOIN `CryptStore` `cs` ON (`pt`.`jwt_crypt` = `cs`.`id`))
             JOIN `TransferCodesAttributes` `tca` ON (`pt`.`id` = `tca`.`id`));

-- RTCryptStore source
CREATE OR REPLACE ALGORITHM = UNDEFINED VIEW `RTCryptStore` AS
SELECT `CryptStore`.`id`      AS `id`,
       `CryptStore`.`crypt`   AS `crypt`,
       `CryptStore`.`created` AS `created`,
       `CryptStore`.`updated` AS `updated`
    FROM `CryptStore`
    WHERE `CryptStore`.`payload_type` = (
        SELECT `CryptPayloadTypes`.`id`
            FROM `CryptPayloadTypes`
            WHERE `CryptPayloadTypes`.`payload_type` = 'RT');

-- ATCryptStore source
CREATE OR REPLACE ALGORITHM = UNDEFINED VIEW `ATCryptStore` AS
SELECT `CryptStore`.`id`      AS `id`,
       `CryptStore`.`crypt`   AS `crypt`,
       `CryptStore`.`created` AS `created`,
       `CryptStore`.`updated` AS `updated`
    FROM `CryptStore`
    WHERE `CryptStore`.`payload_type` = (
        SELECT `CryptPayloadTypes`.`id`
            FROM `CryptPayloadTypes`
            WHERE `CryptPayloadTypes`.`payload_type` = 'AT');

-- MTCryptStore source
CREATE OR REPLACE ALGORITHM = UNDEFINED VIEW `MTCryptStore` AS
SELECT `CryptStore`.`id`      AS `id`,
       `CryptStore`.`crypt`   AS `crypt`,
       `CryptStore`.`created` AS `created`,
       `CryptStore`.`updated` AS `updated`
    FROM `CryptStore`
    WHERE `CryptStore`.`payload_type` = (
        SELECT `CryptPayloadTypes`.`id`
            FROM `CryptPayloadTypes`
            WHERE `CryptPayloadTypes`.`payload_type` = 'MT');

-- SSHPublicKeys definition
CREATE TABLE SSHPublicKeys
(
    user          BIGINT UNSIGNED                      NOT NULL,
    ssh_key_fp    VARCHAR(128)                         NOT NULL,
    key_id        BIGINT UNSIGNED AUTO_INCREMENT,
    MT_crypt      BIGINT UNSIGNED                      NOT NULL,
    created       DATETIME DEFAULT CURRENT_TIMESTAMP() NOT NULL,
    last_used     DATETIME                             NULL,
    ssh_user_hash VARCHAR(128)                         NOT NULL,
    name          TEXT                                 NULL,
    MT_id         VARCHAR(128)                         NOT NULL,
    PRIMARY KEY (user, ssh_key_fp),
    CONSTRAINT ssh_pub_keys_UN
        UNIQUE (key_id),
    CONSTRAINT ssh_pub_keys_UN_1
        UNIQUE (ssh_user_hash),
    CONSTRAINT SSHPublicKeys_FK
        FOREIGN KEY (MT_id) REFERENCES MTokens (id)
            ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT ssh_pub_keys_FK
        FOREIGN KEY (user) REFERENCES Users (id)
            ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT ssh_pub_keys_FK_1
        FOREIGN KEY (MT_crypt) REFERENCES CryptStore (id)
            ON UPDATE CASCADE ON DELETE CASCADE
);

# Procedures
DELIMITER ;;
CREATE OR REPLACE PROCEDURE ATAttribute_Insert(IN ATID BIGINT UNSIGNED, IN VALUE TEXT, IN ATTRIBUTE TEXT)
BEGIN
    INSERT INTO AT_Attributes (AT_id, attribute_id, `attribute`)
        VALUES (ATID, (SELECT attr.id FROM Attributes attr WHERE attr.attribute = ATTRIBUTE), VALUE);
END;;

CREATE OR REPLACE PROCEDURE AT_Insert(IN AT TEXT, IN IP TEXT, IN COMM TEXT, IN MT VARCHAR(128))
BEGIN
    CALL CryptStoreAT_Insert(AT, @ATCryptID);
    INSERT INTO AccessTokens (token_crypt, ip_created, comment, MT_id) VALUES (@ATCryptID, IP, COMM, MT);
    SELECT LAST_INSERT_ID();
END;;

CREATE FUNCTION AddToCryptStore(CRYPT_ TEXT, PAYLOAD_T INT UNSIGNED) RETURNS BIGINT UNSIGNED
BEGIN
    INSERT INTO CryptStore (crypt, payload_type) VALUES (CRYPT_, PAYLOAD_T);
    RETURN LAST_INSERT_ID();
END;;

CREATE OR REPLACE PROCEDURE AuthInfo_Delete(IN STATE TEXT)
BEGIN
    DELETE FROM AuthInfo WHERE state_h = STATE;
END;;

CREATE OR REPLACE PROCEDURE AuthInfo_Get(IN STATE TEXT)
BEGIN
    SELECT state_h,
           request_json,
           polling_code,
           code_verifier
        FROM AuthInfo
        WHERE state_h = STATE
          AND expires_at >= CURRENT_TIMESTAMP();
END;;

CREATE OR REPLACE PROCEDURE AuthInfo_Insert(IN STATE_H_ VARCHAR(128), IN REQUEST LONGTEXT,
                                            IN EXPIRES_IN_ INT, IN POLLING_CODE_ BIT)
BEGIN
    INSERT INTO AuthInfo (`state_h`, `request_json`, `expires_in`, `polling_code`)
        VALUES (STATE_H_, REQUEST, EXPIRES_IN_, POLLING_CODE_);
END;;

CREATE OR REPLACE PROCEDURE AuthInfo_SetCodeVerifier(IN STATE TEXT, IN VERIFIER TEXT)
BEGIN
    UPDATE AuthInfo SET code_verifier = VERIFIER WHERE state_h = STATE;
END;;

CREATE OR REPLACE PROCEDURE AuthInfo_Update(IN STATE TEXT, IN REQUEST LONGTEXT)
BEGIN
    UPDATE AuthInfo
    SET request_json = REQUEST
        WHERE state_h = STATE;
END;;

CREATE OR REPLACE PROCEDURE CryptStoreAT_Insert(IN EncryptedAT TEXT, OUT ID BIGINT UNSIGNED)
BEGIN
    SET ID = (SELECT AddToCryptStore(EncryptedAT,
                                     (SELECT cpt.id FROM CryptPayloadTypes cpt WHERE cpt.payload_type = 'AT')));
END;;

CREATE OR REPLACE PROCEDURE CryptStoreMT_Insert(IN EncryptedMT TEXT, OUT ID BIGINT UNSIGNED)
BEGIN
    SET ID = (SELECT AddToCryptStore(EncryptedMT,
                                     (SELECT cpt.id FROM CryptPayloadTypes cpt WHERE cpt.payload_type = 'MT')));
END;;

CREATE OR REPLACE PROCEDURE CryptStoreRT_Get(IN MTID VARCHAR(128))
BEGIN
    SELECT crypt AS refresh_token FROM CryptStore WHERE id = (SELECT m.rt_id FROM MTokens m WHERE m.id = MTID);
END;;

CREATE OR REPLACE PROCEDURE CryptStoreRT_Insert(IN EncryptedRT TEXT, OUT ID BIGINT UNSIGNED)
BEGIN
    SET ID = (SELECT AddToCryptStore(EncryptedRT,
                                     (SELECT cpt.id FROM CryptPayloadTypes cpt WHERE cpt.payload_type = 'RT')));
END;;

CREATE OR REPLACE PROCEDURE CryptStore_Delete(IN CRYPTID BIGINT UNSIGNED)
BEGIN
    DELETE FROM CryptStore WHERE id = CRYPTID;
END;;

CREATE OR REPLACE PROCEDURE CryptStore_Update(IN CRYPTID BIGINT UNSIGNED, IN VALUE TEXT)
BEGIN
    UPDATE CryptStore SET crypt = VALUE WHERE id = CRYPTID;
END;;

CREATE OR REPLACE PROCEDURE EncryptionKeysRT_Insert(IN CRYPTKEY TEXT, IN RTID BIGINT UNSIGNED, IN MTID VARCHAR(128))
BEGIN
    CALL EncryptionKeys_Insert(CRYPTKEY, @KEYID);
    INSERT INTO RT_EncryptionKeys (rt_id, MT_id, key_id) VALUES (RTID, MTID, @KEYID);
END;;

CREATE OR REPLACE PROCEDURE EncryptionKeys_Delete(IN KEYID BIGINT UNSIGNED)
BEGIN
    DELETE FROM EncryptionKeys WHERE id = KEYID;
END;;

CREATE OR REPLACE PROCEDURE EncryptionKeys_Get(IN KEYID BIGINT UNSIGNED)
BEGIN
    SELECT encryption_key FROM EncryptionKeys WHERE id = KEYID;
END;;

CREATE OR REPLACE PROCEDURE EncryptionKeys_GetRTKeyForMT(IN MTID VARCHAR(128))
BEGIN
    DECLARE rtid BIGINT UNSIGNED;
    DECLARE keyid BIGINT UNSIGNED;
    SELECT rt_id FROM MTokens WHERE id = MTID INTO rtid;
    SELECT key_id FROM RT_EncryptionKeys WHERE MT_id = MTID AND rt_id = rtid INTO keyid;
    SELECT *
        FROM (SELECT ek.encryption_key, ek.id AS key_id FROM EncryptionKeys ek WHERE ek.id = keyid) encr
                 JOIN
             (SELECT crypt AS refresh_token, id AS rt_id FROM CryptStore WHERE id = rtid) rt;
END;;

CREATE OR REPLACE PROCEDURE EncryptionKeys_Insert(IN CRYPTKEY TEXT, OUT ID BIGINT UNSIGNED)
BEGIN
    INSERT INTO EncryptionKeys (encryption_key) VALUES (CRYPTKEY);
    SET ID = (SELECT LAST_INSERT_ID());
END;;

CREATE OR REPLACE PROCEDURE EncryptionKeys_Update(IN KEYID BIGINT UNSIGNED, IN NEW_KEY TEXT)
BEGIN
    UPDATE EncryptionKeys SET encryption_key = NEW_KEY WHERE id = KEYID;
END;;

CREATE OR REPLACE PROCEDURE EventHistory_Get(IN MTID VARCHAR(128))
BEGIN
    SELECT MT_id, event, time, comment, ip, user_agent FROM EventHistory WHERE MT_id = MTID;
END;;

CREATE OR REPLACE PROCEDURE Event_Insert(IN MTID VARCHAR(128), IN EVENT TEXT, IN COMMENT_ TEXT, IN IP_ TEXT,
                                         IN USERAGENT TEXT)
BEGIN
    INSERT INTO MT_Events (MT_id, event_id, comment, ip, user_agent)
        VALUES (MTID, (SELECT e.id FROM Events e WHERE e.event = EVENT), COMMENT_, IP_, USERAGENT);
END;;

CREATE OR REPLACE PROCEDURE Grants_CheckEnabled(IN MTID VARCHAR(128), IN GRANTT TEXT)
BEGIN
    SELECT ug.enabled
        FROM UserGrants ug
        WHERE ug.user_id = (SELECT m.user_id FROM MTokens m WHERE m.id = MTID)
          AND ug.grant_id = (SELECT g.id FROM Grants g WHERE g.grant_type = GRANTT);
END;;

CREATE OR REPLACE PROCEDURE Grants_Disable(IN MTID VARCHAR(128), IN GRANT_T TEXT)
BEGIN
    INSERT INTO UserGrants (user_id, grant_id, enabled)
        VALUES ((SELECT m.user_id FROM MTokens m WHERE m.id = MTID),
                (SELECT g.id FROM Grants g WHERE g.grant_type = GRANT_T), 0)
    ON DUPLICATE KEY UPDATE enabled = 0;
END;;

CREATE OR REPLACE PROCEDURE Grants_Enable(IN MTID VARCHAR(128), IN GRANT_T TEXT)
BEGIN
    INSERT INTO UserGrants (user_id, grant_id, enabled)
        VALUES ((SELECT m.user_id FROM MTokens m WHERE m.id = MTID),
                (SELECT g.id FROM Grants g WHERE g.grant_type = GRANT_T), 1)
    ON DUPLICATE KEY UPDATE enabled = 1;
END;;

CREATE OR REPLACE PROCEDURE Grants_Get(IN MTID VARCHAR(128))
BEGIN
    SELECT g.grant_type, gg.enabled
        FROM (SELECT ug.grant_id, ug.enabled
                  FROM UserGrants ug
                  WHERE ug.user_id = (SELECT m.user_id FROM MTokens m WHERE m.id = MTID)) gg
                 JOIN Grants g ON g.id = gg.grant_id;
END;;

CREATE OR REPLACE PROCEDURE MTokens_Check(IN MTID VARCHAR(128), IN SEQ BIGINT UNSIGNED)
BEGIN
    SELECT COUNT(1) FROM MTokens WHERE id = MTID AND seqno = SEQ;
END;;

CREATE OR REPLACE PROCEDURE MTokens_CheckID(IN MTID VARCHAR(128))
BEGIN
    SELECT COUNT(1) FROM MTokens WHERE id = MTID;
END;;

CREATE OR REPLACE PROCEDURE MTokens_CheckRotating(IN MTID VARCHAR(128), IN SEQ BIGINT UNSIGNED, IN LIFETIME INT)
BEGIN
    SELECT COUNT(1)
        FROM MTokens
        WHERE id = MTID
          AND seqno = SEQ
          AND TIMESTAMPADD(SECOND, LIFETIME, last_rotated) >= CURRENT_TIMESTAMP();
END;;

CREATE OR REPLACE PROCEDURE MTokens_Delete(IN MTID VARCHAR(128))
BEGIN
    DELETE FROM MTokens WHERE id = MTID;
END;;

CREATE OR REPLACE PROCEDURE MTokens_GetAllForSameUser(IN MTID VARCHAR(128))
BEGIN
    DECLARE UID BIGINT UNSIGNED;
    SELECT user_id FROM MTokens WHERE id = MTID INTO UID;
    CALL MTokens_GetForUser(UID);
END;;

CREATE OR REPLACE PROCEDURE MTokens_GetForUser(IN UID BIGINT UNSIGNED)
BEGIN
    SELECT id, parent_id, name, created, ip_created AS ip FROM MTokens WHERE user_id = UID ORDER BY created;
END;;

CREATE OR REPLACE PROCEDURE MTokens_GetRTID(IN MTID VARCHAR(128))
BEGIN
    SELECT rt_id FROM MTokens WHERE id = MTID;
END;;

CREATE OR REPLACE PROCEDURE MTokens_GetSubtokens(IN MTID VARCHAR(128))
BEGIN
    CREATE TEMPORARY TABLE IF NOT EXISTS effected_MTIDs (id VARCHAR(128));
    TRUNCATE effected_MTIDs;
    INSERT INTO effected_MTIDs
    WITH RECURSIVE childs AS (
        SELECT id, parent_id
            FROM MTokens
            WHERE id = MTID
        UNION ALL
        SELECT mt.id, mt.parent_id
            FROM MTokens mt
                     INNER JOIN childs c
            WHERE mt.parent_id = c.id
    )
    SELECT id
        FROM childs;
    SELECT m.id, m.parent_id, m.name, m.created, m.ip_created AS ip
        FROM MTokens m
        WHERE m.id IN
              (SELECT id
                   FROM effected_MTIDs);
    DROP TABLE effected_MTIDs;
END;;

CREATE OR REPLACE PROCEDURE MTokens_Insert(IN SUB TEXT, IN ISS TEXT, IN MTID VARCHAR(128), IN SEQNO_ BIGINT UNSIGNED,
                                           IN PARENT VARCHAR(128), IN RTID BIGINT UNSIGNED,
                                           IN NAME_ TEXT, IN IP TEXT)
BEGIN
    CALL Users_GetID(SUB, ISS, @UID);
    INSERT INTO MTokens (id, seqno, parent_id, rt_id, name, ip_created, user_id)
        VALUES (MTID, SEQNO_, PARENT, RTID, NAME_, IP, @UID);
END;;

CREATE OR REPLACE PROCEDURE MTokens_RevokeRec(IN MTID VARCHAR(128))
BEGIN
    CREATE TEMPORARY TABLE IF NOT EXISTS effected_MTIDs (id VARCHAR(128));
    TRUNCATE effected_MTIDs;
    INSERT INTO effected_MTIDs
    WITH RECURSIVE childs AS (
        SELECT id, parent_id
            FROM MTokens
            WHERE id = MTID
        UNION ALL
        SELECT mt.id, mt.parent_id
            FROM MTokens mt
                     INNER JOIN childs c
            WHERE mt.parent_id = c.id
    )
    SELECT id
        FROM childs;
    DELETE
        FROM EncryptionKeys
        WHERE id = ANY (SELECT key_id FROM RT_EncryptionKeys WHERE MT_id IN (SELECT id FROM effected_MTIDs));
    DELETE FROM MTokens WHERE id IN (SELECT id FROM effected_MTIDs);
    DROP TABLE effected_MTIDs;
END;;

CREATE OR REPLACE PROCEDURE MTokens_UpdateSeqNo(IN MTID VARCHAR(128), IN SEQNO_ BIGINT UNSIGNED)
BEGIN
    UPDATE MTokens SET seqno=SEQNO_, last_rotated = CURRENT_TIMESTAMP() WHERE id = MTID;
END;;

CREATE OR REPLACE PROCEDURE MTokens_getName(IN MTID VARCHAR(128))
BEGIN
    SELECT name FROM MTokens WHERE id = MTID;
END;;

CREATE OR REPLACE PROCEDURE ProxyTokens_Delete(IN PTID VARCHAR(128))
BEGIN
    DECLARE jwtID BIGINT UNSIGNED;
    SELECT pt.jwt_crypt FROM ProxyTokens pt WHERE pt.id = PTID INTO jwtID;
    DELETE FROM ProxyTokens WHERE id = PTID;
    DELETE FROM CryptStore WHERE id = jwtID;
END;;

CREATE OR REPLACE PROCEDURE ProxyTokens_GetMT(IN PTID VARCHAR(128))
BEGIN
    SELECT cs.crypt AS jwt, pp.MT_id
        FROM (SELECT pt.jwt_crypt, pt.MT_id FROM ProxyTokens pt WHERE pt.id = PTID) pp
                 JOIN CryptStore cs ON pp.jwt_crypt = cs.id;
END;;

CREATE OR REPLACE PROCEDURE ProxyTokens_Insert(IN PTID VARCHAR(128), IN JWT TEXT, IN MTID VARCHAR(128))
BEGIN
    CALL CryptStoreMT_Insert(JWT, @JWTID);
    INSERT INTO ProxyTokens (id, jwt_crypt, MT_id) VALUES (PTID, @JWTID, MTID);
END;;

CREATE OR REPLACE PROCEDURE ProxyTokens_Update(IN PTID VARCHAR(128), IN JWT TEXT, IN MTID VARCHAR(128))
BEGIN
    DECLARE jwtID BIGINT UNSIGNED;
    SELECT pt.jwt_crypt FROM ProxyTokens pt WHERE pt.id = PTID INTO jwtID;
    IF (jwtID IS NULL) THEN
        CALL CryptStoreMT_Insert(JWT, @JWTID);
        UPDATE ProxyTokens SET MT_id=MTID, jwt_crypt=@JWTID WHERE id = PTID;
    ELSE
        UPDATE CryptStore SET crypt=JWT WHERE id = jwtID;
        UPDATE ProxyTokens SET MT_id=MTID WHERE id = PTID;
    END IF;
END;;

CREATE OR REPLACE PROCEDURE RT_CountLinks(IN RTID BIGINT UNSIGNED)
BEGIN
    SELECT COUNT(1) FROM MTokens WHERE rt_id = RTID;
END;;

CREATE OR REPLACE PROCEDURE SSHInfo_Delete(IN MTID VARCHAR(128), IN FP VARCHAR(128))
BEGIN
    DECLARE uid BIGINT UNSIGNED;
    DECLARE sshMTID VARCHAR(128);
    DECLARE cid BIGINT UNSIGNED;
    DECLARE rid BIGINT UNSIGNED;
    DECLARE rckid BIGINT UNSIGNED;
    DECLARE rtCount INT;

    SELECT m.`user_id` FROM MTokens m WHERE m.id = MTID INTO uid;
    SELECT s.MT_id, s.MT_crypt FROM SSHPublicKeys s WHERE s.ssh_key_fp = FP AND s.user = uid INTO sshMTID, cid;
    SELECT m.`rt_id` FROM MTokens m WHERE m.id = sshMTID INTO rid;
    SELECT k.`key_id` FROM RT_EncryptionKeys k WHERE k.rt_id = rid AND k.MT_id = sshMTID INTO rckid;
    CALL EncryptionKeys_Delete(rckid);
    CALL MTokens_Delete(sshMTID);
    SELECT COUNT(1) FROM MTokens WHERE rt_id = rid INTO rtCount;
    IF (rtCount = 0) THEN
        CALL CryptStore_Delete(rid);
    END IF;
    CALL CryptStore_Delete(cid);
    DELETE FROM SSHPublicKeys WHERE `user` = uid AND ssh_key_fp = FP;
END;;

CREATE OR REPLACE PROCEDURE SSHInfo_Get(IN KeyHash VARCHAR(128), IN UserHash VARCHAR(128))
BEGIN
    SELECT spk.key_id,
           spk.name,
           spk.ssh_key_fp,
           spk.ssh_user_hash,
           spk.created,
           spk.last_used,
           e.enabled,
           ms.crypt AS MT_crypt
        FROM ((SELECT * FROM SSHPublicKeys WHERE ssh_key_fp = KeyHash AND ssh_user_hash = UserHash) spk
            JOIN (SELECT ug.user_id, ug.enabled
                      FROM (UserGrants ug
                               JOIN
                               (SELECT * FROM Grants gg WHERE gg.grant_type = 'ssh') g
                               ON ug.grant_id = g.id)
            ) e
            ON spk.`user` = e.user_id)
                 JOIN MTCryptStore ms
                      ON spk.MT_crypt = ms.id;
END;;

CREATE OR REPLACE PROCEDURE SSHInfo_GetAll(IN MTID VARCHAR(128))
BEGIN
    SELECT s.ssh_key_fp, s.name, s.created, s.last_used
        FROM SSHPublicKeys s
        WHERE s.`user` = (SELECT m.`user_id` FROM MTokens m WHERE m.id = MTID);
END;;

CREATE OR REPLACE PROCEDURE SSHInfo_Insert(IN MTID VARCHAR(128), IN KEY_FP VARCHAR(128), IN SSH_USER_H VARCHAR(128),
                                           IN NAME_ TEXT, IN ENCRYPTED_MT TEXT)
BEGIN
    CALL CryptStoreMT_Insert(ENCRYPTED_MT, @CRYPT_ID);
    INSERT INTO SSHPublicKeys (user, ssh_key_fp, ssh_user_hash, name, MT_crypt, MT_id)
        VALUES ((SELECT m.user_id FROM MTokens m WHERE m.id = MTID), KEY_FP, SSH_USER_H, NAME_, @CRYPT_ID, MTID);
END;;

CREATE OR REPLACE PROCEDURE SSHInfo_UsedKey(IN KEY_FP VARCHAR(128), IN USER_H VARCHAR(128))
BEGIN
    UPDATE SSHPublicKeys SET last_used = CURRENT_TIMESTAMP() WHERE ssh_key_fp = KEY_FP AND ssh_user_hash = USER_H;
END;;

CREATE OR REPLACE PROCEDURE TokenUsages_GetAT(IN MTID VARCHAR(128), IN RHASH CHAR(128))
BEGIN
    SELECT usages_AT FROM TokenUsages WHERE restriction_hash = RHASH AND MT_id = MTID;
END;;

CREATE OR REPLACE PROCEDURE TokenUsages_GetOther(IN MTID VARCHAR(128), IN RHASH CHAR(128))
BEGIN
    SELECT usages_other FROM TokenUsages WHERE restriction_hash = RHASH AND MT_id = MTID;
END;;

CREATE OR REPLACE PROCEDURE TokenUsages_IncrAT(IN MTID VARCHAR(128), IN RESTRICTION_ LONGTEXT, IN RHASH CHAR(128))
BEGIN
    INSERT INTO TokenUsages (MT_id, restriction, restriction_hash, usages_AT)
        VALUES (MTID, RESTRICTION_, RHASH, 1)
    ON DUPLICATE KEY UPDATE usages_AT = usages_AT + 1;
END;;

CREATE OR REPLACE PROCEDURE TokenUsages_IncrOther(IN MTID VARCHAR(128), IN RESTRICTION_ LONGTEXT, IN RHASH CHAR(128))
BEGIN
    INSERT INTO TokenUsages (MT_id, restriction, restriction_hash, usages_other)
        VALUES (MTID, RESTRICTION_, RHASH, 1)
    ON DUPLICATE KEY UPDATE usages_other = usages_other + 1;
END;;

CREATE OR REPLACE PROCEDURE TransferCodeAttributes_DeclineConsent(IN TCID VARCHAR(128))
BEGIN
    UPDATE TransferCodesAttributes SET consent_declined=1 WHERE id = TCID;
END;;

CREATE OR REPLACE PROCEDURE TransferCodeAttributes_GetRevokeJWT(IN TCID VARCHAR(128))
BEGIN
    SELECT revoke_MT FROM TransferCodesAttributes WHERE id = TCID;
END;;

CREATE OR REPLACE PROCEDURE TransferCodeAttributes_Insert(IN TCID VARCHAR(128), IN EXPIRES_IN_ INT, IN REVOKE_MT_ BIT,
                                                          IN RESPONSE_TYPE_ TEXT, IN MAX_TOKEN_LEN_ INT)
BEGIN
    INSERT INTO TransferCodesAttributes (id, expires_in, revoke_MT, response_type, max_token_len)
        VALUES (TCID, EXPIRES_IN_, REVOKE_MT_, RESPONSE_TYPE_, MAX_TOKEN_LEN_);
END;;

CREATE OR REPLACE PROCEDURE TransferCodeAttributes_UpdateSSHKey(IN PCID VARCHAR(128), IN KEY_FP VARCHAR(128))
BEGIN
    UPDATE TransferCodesAttributes SET ssh_key_fp=KEY_FP WHERE id = PCID;
END;;

CREATE OR REPLACE PROCEDURE TransferCodes_GetStatus(IN PCID VARCHAR(128))
BEGIN
    SELECT 1                                AS found,
           CURRENT_TIMESTAMP() > expires_at AS expired,
           response_type,
           consent_declined,
           max_token_len,
           ssh_key_fp
        FROM TransferCodes
        WHERE id = PCID;
END;;

CREATE OR REPLACE PROCEDURE Users_GetID(IN SUB TEXT, IN ISS TEXT, OUT UID BIGINT UNSIGNED)
BEGIN
    SET UID = (SELECT id FROM Users u WHERE u.sub = SUB AND u.iss = ISS);
    IF (UID IS NULL) THEN
        CALL Users_Insert(SUB, ISS, UID);
    END IF;
END;;

CREATE OR REPLACE PROCEDURE Users_Insert(IN SUB_ TEXT, IN ISS_ TEXT, OUT UID BIGINT UNSIGNED)
BEGIN
    INSERT INTO Users (sub, iss) VALUES (SUB_, ISS_);
    SET UID = (SELECT LAST_INSERT_ID());
END;;

CREATE OR REPLACE PROCEDURE Version_Get()
BEGIN
    SELECT version, bef, aft FROM version;
END;;

CREATE OR REPLACE PROCEDURE Version_SetAfter(IN VERSION TEXT)
BEGIN
    INSERT INTO version (version, aft)
        VALUES (VERSION, CURRENT_TIMESTAMP())
    ON DUPLICATE KEY UPDATE aft=CURRENT_TIMESTAMP();
END;;

CREATE OR REPLACE PROCEDURE Version_SetBefore(IN VERSION TEXT)
BEGIN
    INSERT INTO version (version, bef)
        VALUES (VERSION, CURRENT_TIMESTAMP())
    ON DUPLICATE KEY UPDATE bef=CURRENT_TIMESTAMP();
END;;
DELIMITER ;


# Predefined Values
INSERT IGNORE INTO Grants (grant_type)
    VALUES ('ssh');

INSERT IGNORE INTO CryptPayloadTypes (payload_type)
    VALUES ('RT');
INSERT IGNORE INTO CryptPayloadTypes (payload_type)
    VALUES ('AT');
INSERT IGNORE INTO CryptPayloadTypes (payload_type)
    VALUES ('MT');

INSERT IGNORE INTO Events (event)
    VALUES ('settings_grant_enable');
INSERT IGNORE INTO Events (event)
    VALUES ('settings_grant_disable');
INSERT IGNORE INTO Events (event)
    VALUES ('settings_grants_listed');
INSERT IGNORE INTO Events (event)
    VALUES ('ssh_keys_listed');
INSERT IGNORE INTO Events (event)
    VALUES ('ssh_key_added');

DELETE
    FROM Grants
    WHERE grant_type = 'access_token';
DELETE
    FROM Grants
    WHERE grant_type = 'private_key_jwt';

DELETE
    FROM Events
    WHERE event = 'mng_disabled_AT_grant';
DELETE
    FROM Events
    WHERE event = 'mng_disabled_JWT_grant';
DELETE
    FROM Events
    WHERE event = 'mng_disabled_tracing';
DELETE
    FROM Events
    WHERE event = 'mng_enabled_AT_grant';
DELETE
    FROM Events
    WHERE event = 'mng_enabled_JWT_grant';
DELETE
    FROM Events
    WHERE event = 'mng_enabled_tracing';
DELETE
    FROM Events
    WHERE event = 'mng_linked_grant';
DELETE
    FROM Events
    WHERE event = 'mng_unlinked_grant';
