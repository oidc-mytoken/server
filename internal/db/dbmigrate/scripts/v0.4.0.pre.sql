# Tables
DROP VIEW IF EXISTS MyTokens;

ALTER TABLE Users
    DROP COLUMN token_tracing;
ALTER TABLE Users
    DROP COLUMN jwt_pk;

# CryptStore
CREATE TABLE `CryptPayloadTypes`
(
    `id`           int(10) unsigned NOT NULL AUTO_INCREMENT,
    `payload_type` varchar(128)     NOT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `CryptPayloadTypes_UN` (`payload_type`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 4
  DEFAULT CHARSET = utf8mb4;

RENAME TABLE RefreshTokens TO CryptStore;
ALTER TABLE CryptStore
    CHANGE rt crypt text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL;
ALTER TABLE CryptStore
    ADD COLUMN IF NOT EXISTS payload_type INT UNSIGNED NULL;
ALTER TABLE CryptStore
    ADD CONSTRAINT CryptStore_FK FOREIGN KEY (payload_type) REFERENCES CryptPayloadTypes (id) ON UPDATE CASCADE;
UPDATE CryptStore
SET payload_type=(SELECT id FROM CryptPayloadTypes WHERE payload_type = 'RT')
WHERE payload_type IS NULL;
ALTER TABLE CryptStore
    MODIFY COLUMN payload_type INT UNSIGNED NOT NULL;

ALTER TABLE AccessTokens
    DROP COLUMN IF EXISTS created;
ALTER TABLE AccessTokens
    ADD COLUMN IF NOT EXISTS token_crypt BIGINT UNSIGNED NULL;
ALTER TABLE AccessTokens
    ADD CONSTRAINT AccessTokens_FK_1 FOREIGN KEY (token_crypt) REFERENCES CryptStore (id) ON UPDATE CASCADE;
UPDATE AccessTokens a
SET token_crypt=(SELECT id FROM CryptStore c WHERE c.crypt = a.token)
WHERE token_crypt IS NULL;
ALTER TABLE AccessTokens
    MODIFY COLUMN token_crypt BIGINT UNSIGNED NOT NULL;
ALTER TABLE AccessTokens
    DROP COLUMN IF EXISTS token;

CREATE OR REPLACE ALGORITHM = UNDEFINED VIEW `MyTokens` AS
SELECT `mt`.`id`               AS `id`,
       `mt`.`seqno`            AS `seqno`,
       `mt`.`parent_id`        AS `parent_id`,
       `mt`.`root_id`          AS `root_id`,
       `mt`.`name`             AS `name`,
       `mt`.`created`          AS `created`,
       `mt`.`ip_created`       AS `ip_created`,
       `mt`.`user_id`          AS `user_id`,
       `mt`.`rt_id`            AS `rt_id`,
       `cyp`.`crypt`           AS `refresh_token`,
       `cyp`.`updated`         AS `rt_updated`,
       `keys`.`encryption_key` AS `encryption_key`
FROM (((`MTokens` `mt`
    JOIN `CryptStore` `cyp` ON
        (`mt`.`rt_id` = `cyp`.`id`))
    JOIN `RT_EncryptionKeys` `rkeys` ON
        (`mt`.`id` = `rkeys`.`MT_id`
            AND `mt`.`rt_id` = `rkeys`.`rt_id`))
         JOIN `EncryptionKeys` `keys` ON
    (`rkeys`.`key_id` = `keys`.`id`));


-- RTCryptStore source
CREATE OR REPLACE ALGORITHM = UNDEFINED VIEW `RTCryptStore` AS
select `CryptStore`.`id`      AS `id`,
       `CryptStore`.`crypt`   AS `crypt`,
       `CryptStore`.`created` AS `created`,
       `CryptStore`.`updated` AS `updated`
from `CryptStore`
where `CryptStore`.`payload_type` = (
    select `CryptPayloadTypes`.`id`
    from `CryptPayloadTypes`
    where `CryptPayloadTypes`.`payload_type` = 'RT');

-- ATCryptStore source
CREATE OR REPLACE ALGORITHM = UNDEFINED VIEW `ATCryptStore` AS
select `CryptStore`.`id`      AS `id`,
       `CryptStore`.`crypt`   AS `crypt`,
       `CryptStore`.`created` AS `created`,
       `CryptStore`.`updated` AS `updated`
from `CryptStore`
where `CryptStore`.`payload_type` = (
    select `CryptPayloadTypes`.`id`
    from `CryptPayloadTypes`
    where `CryptPayloadTypes`.`payload_type` = 'AT');

-- MTCryptStore source
CREATE OR REPLACE ALGORITHM = UNDEFINED VIEW `MTCryptStore` AS
select `CryptStore`.`id`      AS `id`,
       `CryptStore`.`crypt`   AS `crypt`,
       `CryptStore`.`created` AS `created`,
       `CryptStore`.`updated` AS `updated`
from `CryptStore`
where `CryptStore`.`payload_type` = (
    select `CryptPayloadTypes`.`id`
    from `CryptPayloadTypes`
    where `CryptPayloadTypes`.`payload_type` = 'MT');


# Procedures

create or replace procedure ATAttribute_Insert(IN ATID bigint unsigned, IN VALUE text, IN ATTRIBUTE text)
BEGIN
    INSERT INTO AT_Attributes (AT_id, attribute_id, `attribute`)
    VALUES (ATID, (SELECT attr.id FROM Attributes attr WHERE attr.attribute = ATTRIBUTE), VALUE);
END;

create or replace procedure AT_Insert(IN AT text, IN IP text, IN COMM text, IN MTID varchar(128))
BEGIN
    CALL CryptStoreAT_Insert(AT, @ATCryptID);
    INSERT INTO AccessTokens (token_crypt, ip_created, comment, MT_id) VALUES (@ATCryptID, IP, COMM, MTID);
    SELECT LAST_INSERT_ID();
END;

create function AddToCryptStore(CRYPT_VALUE text, PTYPE int unsigned) returns bigint unsigned
BEGIN
    INSERT INTO CryptStore (crypt, payload_type) VALUES (CRYPT_VALUE, PTYPE);
    RETURN LAST_INSERT_ID();
END;

create or replace procedure AuthInfo_Delete(IN STATE text)
BEGIN
    DELETE FROM AuthInfo WHERE state_h = STATE;
END;

create or replace procedure AuthInfo_Get(IN STATE text)
BEGIN
    SELECT state_h,
           iss,
           restrictions,
           capabilities,
           subtoken_capabilities,
           name,
           polling_code,
           rotation,
           response_type,
           max_token_len,
           code_verifier
    FROM AuthInfo
    WHERE state_h = STATE
      AND expires_at >= CURRENT_TIMESTAMP();
END;

create or replace procedure AuthInfo_Insert(IN STATE_H varchar(128), IN ISS text, IN RESTRICTIONS longtext,
                                            IN CAPABILITIES longtext, IN SUBTOKEN_CAPABILITIES longtext, IN NAME text,
                                            IN EXPIRES_IN int, IN POLLING_CODE bit, IN ROTATION longtext,
                                            IN RESPONSE_TYPE varchar(128), IN MAX_TOKEN_LEN int)
BEGIN
    INSERT INTO AuthInfo (`state_h`, `iss`, `restrictions`, `capabilities`, `subtoken_capabilities`, `name`,
                          `expires_in`, `polling_code`, `rotation`, `response_type`, `max_token_len`)
    VALUES (STATE_H, ISS, RESTRICTIONS, CAPABILITIES, SUBTOKEN_CAPABILITIES, NAME, EXPIRES_IN, POLLING_CODE, ROTATION,
            RESPONSE_TYPE, MAX_TOKEN_LEN);
END;

create or replace procedure AuthInfo_SetCodeVerifier(IN STATE text, IN VERIFIER text)
BEGIN
    UPDATE AuthInfo SET code_verifier = VERIFIER WHERE state_h = STATE;
END;

create or replace procedure AuthInfo_Update(IN STATE text, IN RESTRICTIONS longtext, IN CAPABILITIES longtext,
                                            IN SUBTOKEN_CAPABILITIES longtext, IN ROTATION longtext, IN NAME text)
BEGIN
    UPDATE AuthInfo
    SET restrictions          = RESTRICTIONS,
        capabilities          = CAPABILITIES,
        subtoken_capabilities = SUBTOKEN_CAPABILITIES,
        rotation              = ROTATION,
        name                  = NAME
    WHERE state_h = STATE;
END;

create or replace procedure CryptStoreAT_Insert(IN EncryptedAT text, OUT ID bigint unsigned)
BEGIN
    SET ID = (SELECT AddToCryptStore(EncryptedAT,
                                     (SELECT cpt.id FROM CryptPayloadTypes cpt WHERE cpt.payload_type = 'AT')));
END;

create or replace procedure CryptStoreRT_Get(IN MTID varchar(128))
BEGIN
    SELECT crypt AS refresh_token FROM CryptStore WHERE id = (SELECT m.rt_id FROM MTokens m WHERE m.id = MTID);
END;

create or replace procedure CryptStoreRT_Insert(IN EncryptedRT text, OUT ID bigint unsigned)
BEGIN
    SET ID = (SELECT AddToCryptStore(EncryptedRT,
                                     (SELECT cpt.id FROM CryptPayloadTypes cpt WHERE cpt.payload_type = 'RT')));
END;

create or replace procedure CryptStore_Delete(IN CRYPTID bigint unsigned)
BEGIN
    DELETE FROM CryptStore WHERE id = CRYPTID;
END;

create or replace procedure CryptStore_Update(IN CRYPTID bigint unsigned, IN VALUE text)
BEGIN
    UPDATE CryptStore SET crypt = VALUE WHERE id = CRYPTID;
END;

create or replace procedure EncryptionKeysRT_Insert(IN CRYPTKEY text, IN RTID bigint unsigned, IN MTID varchar(128))
BEGIN
    CALL EncryptionKeys_Insert(CRYPTKEY, @KEYID);
    INSERT INTO RT_EncryptionKeys (rt_id, MT_id, key_id) VALUES (RTID, MTID, @KEYID);
END;

create or replace procedure EncryptionKeys_Delete(IN KEYID bigint unsigned)
BEGIN
    DELETE FROM EncryptionKeys WHERE id = KEYID;
END;

create or replace procedure EncryptionKeys_Get(IN KEYID bigint unsigned)
BEGIN
    SELECT encryption_key FROM EncryptionKeys WHERE id = KEYID;
END;

create or replace procedure EncryptionKeys_GetRTKeyForMT(IN MTID varchar(128))
BEGIN
    DECLARE rtid bigint unsigned;
    DECLARE keyid bigint unsigned;
    SELECT rt_id FROM MTokens WHERE id = MTID INTO rtid;
    SELECT key_id FROM RT_EncryptionKeys WHERE MT_id = MTID AND rt_id = rtid INTO keyid;
    SELECT *
    FROM (SELECT ek.encryption_key, ek.id AS key_id FROM EncryptionKeys ek WHERE ek.id = keyid) encr
             JOIN
         (SELECT crypt AS refresh_token, id AS rt_id FROM CryptStore WHERE id = rtid) rt;
END;

create or replace procedure EncryptionKeys_Insert(IN CRYPTKEY text, OUT ID bigint unsigned)
BEGIN
    INSERT INTO EncryptionKeys (encryption_key) VALUES (CRYPTKEY);
    SET ID = (SELECT LAST_INSERT_ID());
END;

create or replace procedure EncryptionKeys_Update(IN KEYID bigint unsigned, IN NEW_KEY text)
BEGIN
    UPDATE EncryptionKeys SET encryption_key = NEW_KEY WHERE id = KEYID;
END;

create or replace procedure EventHistory_Get(IN MTID varchar(128))
BEGIN
    SELECT MT_id, event, time, comment, ip, user_agent FROM EventHistory WHERE MT_id = MTID;
END;

create or replace procedure Event_Insert(IN MTID varchar(128), IN EVENT text, IN COMMENT text, IN IP text,
                                         IN USERAGENT text)
BEGIN
    INSERT INTO MT_Events (MT_id, event_id, comment, ip, user_agent)
    VALUES (MTID, (SELECT e.id FROM Events e WHERE e.event = EVENT), COMMENT, IP, USERAGENT);
END;

create or replace procedure MTokens_Check(IN MTID varchar(128), IN SEQ bigint unsigned)
BEGIN
    SELECT COUNT(1) FROM MTokens WHERE id = MTID AND seqno = SEQ;
END;

create or replace procedure MTokens_CheckID(IN MTID varchar(128))
BEGIN
    SELECT COUNT(1) FROM MTokens WHERE id = MTID;
END;

create or replace procedure MTokens_CheckRotating(IN MTID varchar(128), IN SEQ bigint unsigned, IN LIFETIME int)
BEGIN
    SELECT COUNT(1)
    FROM MTokens
    WHERE id = MTID
      AND seqno = SEQ
      AND TIMESTAMPADD(SECOND, LIFETIME, last_rotated) >= CURRENT_TIMESTAMP();
END;

create or replace procedure MTokens_Delete(IN MTID varchar(128))
BEGIN
    DELETE FROM MTokens WHERE id = MTID;
END;

create or replace procedure MTokens_GetAllForSameUser(IN MTID varchar(128))
BEGIN
    DECLARE UID BIGINT UNSIGNED;
    SELECT user_id FROM MTokens WHERE id = MTID INTO UID;
    CALL MTokens_GetForUser(UID);
END;

create or replace procedure MTokens_GetForUser(IN UID bigint unsigned)
BEGIN
    SELECT id, parent_id, root_id, name, created, ip_created AS ip FROM MTokens WHERE user_id = UID ORDER BY created;
END;

create or replace procedure MTokens_GetRTID(IN MTID varchar(128))
BEGIN
    SELECT rt_id FROM MTokens WHERE id = MTID;
END;

create or replace procedure MTokens_GetRoot(IN MTID varchar(128))
BEGIN
    SELECT root_id FROM MTokens WHERE id = MTID;
END;

create or replace procedure MTokens_GetSubtokens(IN MTID varchar(128))
BEGIN
    DECLARE ROOTID varchar(128);
    SELECT root_id FROM MTokens m WHERE m.id = MTID INTO ROOTID;
    IF (ROOTID IS NULL) THEN
        SET ROOTID = MTID;
    END IF;
    SELECT m.id, m.parent_id, m.root_id, m.name, m.created, m.ip_created AS ip
    FROM MTokens m
    WHERE m.id = MTID
       OR m.root_id = ROOTID;
END;

create or replace procedure MTokens_Insert(IN SUB text, IN ISS text, IN MTID varchar(128), IN SEQNO bigint unsigned,
                                           IN PARENT varchar(128), IN ROOT varchar(128), IN RTID bigint unsigned,
                                           IN NAME text, IN IP text)
BEGIN
    CALL Users_GetID(SUB, ISS, @UID);
    INSERT INTO MTokens (id, seqno, parent_id, root_id, rt_id, name, ip_created, user_id)
    VALUES (MTID, SEQNO, PARENT, ROOT, RTID, NAME, IP, @UID);
END;

create or replace procedure MTokens_RevokeRec(IN MTID varchar(128))
BEGIN
    CREATE TEMPORARY TABLE IF NOT EXISTS effected_MTIDs
    (
        id VARCHAR(128)
    );
    TRUNCATE effected_MTIDs;
    INSERT INTO effected_MTIDs
    WITH Recursive childs AS (
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
END;

create or replace procedure MTokens_UpdateSeqNo(IN MTID varchar(128), IN SEQNO bigint unsigned)
BEGIN
    UPDATE MTokens SET seqno=SEQNO, last_rotated = CURRENT_TIMESTAMP() WHERE id = MTID;
END;

create or replace procedure ProxyTokens_Delete(IN PTID varchar(128))
BEGIN
    DELETE FROM ProxyTokens WHERE id = PTID;
END;

create or replace procedure ProxyTokens_GetMT(IN PTID varchar(128))
BEGIN
    SELECT jwt, MT_id FROM ProxyTokens WHERE id = PTID;
END;

create or replace procedure ProxyTokens_Insert(IN PTID varchar(128), IN JWT text, IN MTID varchar(128))
BEGIN
    INSERT INTO ProxyTokens (id, jwt, MT_id) VALUES (PTID, JWT, MTID);
END;

create or replace procedure ProxyTokens_Update(IN PTID varchar(128), IN JWT text, IN MTID varchar(128))
BEGIN
    UPDATE ProxyTokens SET jwt = JWT, MT_id = MTID WHERE id = PTID;
END;

create or replace procedure RT_CountLinks(IN RTID bigint unsigned)
BEGIN
    SELECT COUNT(1) FROM MTokens WHERE rt_id = RTID;
END;

create or replace procedure SSHInfo_Get(IN KeyHash varchar(128), IN UserHash varchar(128))
BEGIN
    SELECT spk.key_id,
           spk.name,
           spk.ssh_key_hash,
           spk.ssh_user_hash,
           spk.created,
           spk.last_used,
           e.enabled,
           ms.crypt AS MT_crypt,
           ek.encryption_key
    FROM ((SELECT * FROM ssh_pub_keys WHERE ssh_key_hash = KeyHash AND ssh_user_hash = UserHash) spk
        JOIN (SELECT ug.user_id, ug.enabled
              FROM (UserGrants ug
                       JOIN
                       (SELECT * FROM Grants gg WHERE gg.grant_type = 'ssh') g
                       ON ug.grant_id = g.id)
        ) e
        ON spk.`user` = e.user_id)
             JOIN MTCryptStore ms
                  ON spk.MT_crypt = ms.id
             JOIN EncryptionKeys ek
                  ON spk.encryption_key = ek.id;
END;

create or replace procedure TokenUsages_GetAT(IN MTID varchar(128), IN RHASH char(128))
BEGIN
    SELECT usages_AT FROM TokenUsages WHERE restriction_hash = RHASH AND MT_id = MTID;
END;

create or replace procedure TokenUsages_GetOther(IN MTID varchar(128), IN RHASH char(128))
BEGIN
    SELECT usages_other FROM TokenUsages WHERE restriction_hash = RHASH AND MT_id = MTID;
END;

create or replace procedure TokenUsages_IncrAT(IN MTID varchar(128), IN RESTRICTION longtext, IN RHASH char(128))
BEGIN
    INSERT INTO TokenUsages (MT_id, restriction, restriction_hash, usages_AT)
    VALUES (MTID, RESTRICTION, RHASH, 1)
    ON DUPLICATE KEY UPDATE usages_AT = usages_AT + 1;
END;

create or replace procedure TokenUsages_IncrOther(IN MTID varchar(128), IN RESTRICTION longtext, IN RHASH char(128))
BEGIN
    INSERT INTO TokenUsages (MT_id, restriction, restriction_hash, usages_other)
    VALUES (MTID, RESTRICTION, RHASH, 1)
    ON DUPLICATE KEY UPDATE usages_other = usages_other + 1;
END;

create or replace procedure TransferCodeAttributes_DeclineConsent(IN TCID varchar(128))
BEGIN
    UPDATE TransferCodesAttributes SET consent_declined=1 WHERE id = TCID;
END;

create or replace procedure TransferCodeAttributes_GetRevokeJWT(IN TCID varchar(128))
BEGIN
    SELECT revoke_MT FROM TransferCodesAttributes WHERE id = TCID;
END;

create or replace procedure TransferCodeAttributes_Insert(IN TCID varchar(128), IN EXPIRES_IN int, IN REVOKE_MT bit,
                                                          IN RESPONSE_TYPE text, IN MAX_TOKEN_LEN int)
BEGIN
    INSERT INTO TransferCodesAttributes (id, expires_in, revoke_MT, response_type, max_token_len)
    VALUES (TCID, EXPIRES_IN, REVOKE_MT, RESPONSE_TYPE, MAX_TOKEN_LEN);
END;

create or replace procedure TransferCodes_GetStatus(IN PCID varchar(128))
BEGIN
    SELECT 1 as found, CURRENT_TIMESTAMP() > expires_at AS expired, response_type, consent_declined, max_token_len
    FROM TransferCodes
    WHERE id = PCID;
END;

create or replace procedure Users_GetID(IN SUB text, IN ISS text, OUT UID bigint unsigned)
BEGIN
    SET UID = (SELECT id FROM Users u WHERE u.sub = SUB AND u.iss = ISS);
    IF (UID IS NULL) THEN
        CALL Users_Insert(SUB, ISS, UID);
    END IF;
END;

create or replace procedure Users_Insert(IN SUB text, IN ISS text, OUT UID bigint unsigned)
BEGIN
    INSERT INTO Users (sub, iss) VALUES (SUB, ISS);
    SET UID = (SELECT LAST_INSERT_ID());
END;

create or replace procedure Version_Get()
BEGIN
    SELECT version, bef, aft FROM version;
END;

create or replace procedure Version_SetAfter(IN VERSION text)
BEGIN
    INSERT INTO version (version, aft)
    VALUES (VERSION, current_timestamp())
    ON DUPLICATE KEY UPDATE aft=current_timestamp();
END;

create or replace procedure Version_SetBefore(IN VERSION text)
BEGIN
    INSERT INTO version (version, bef)
    VALUES (VERSION, current_timestamp())
    ON DUPLICATE KEY UPDATE bef=current_timestamp();
END;


