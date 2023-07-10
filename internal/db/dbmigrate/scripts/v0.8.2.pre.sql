# Procedures
DELIMITER ;;

CREATE OR REPLACE PROCEDURE ATAttribute_Insert(IN ATID BIGINT UNSIGNED, IN VALUE TEXT, IN ATTRIBUTE TEXT)
BEGIN
    SET TIME_ZONE = "+0:00";
    INSERT INTO AT_Attributes (AT_id, attribute_id, `attribute`)
        VALUES (ATID, (SELECT attr.id FROM Attributes attr WHERE attr.attribute = ATTRIBUTE), VALUE);
END;;

CREATE OR REPLACE PROCEDURE AT_Insert(IN AT TEXT, IN IP TEXT, IN COMMENT TEXT, IN MT VARCHAR(128))
BEGIN
    SET TIME_ZONE = "+0:00";
    CALL CryptStoreAT_Insert(AT, @ATCryptID);
    INSERT INTO AccessTokens (token_crypt, ip_created, comment, MT_id) VALUES (@ATCryptID, IP, COMMENT, MT);
    SELECT LAST_INSERT_ID();
END;;

CREATE OR REPLACE PROCEDURE AddToCryptStore(IN crypt TEXT, IN payload_type INT UNSIGNED, OUT ID BIGINT UNSIGNED)
BEGIN
    SET TIME_ZONE = "+0:00";
    INSERT INTO CryptStore (crypt, payload_type) VALUES (crypt, payload_type);
    SET ID = LAST_INSERT_ID();
END;;

CREATE OR REPLACE PROCEDURE AuthInfo_Get(IN STATE TEXT)
BEGIN
    SET TIME_ZONE = "+0:00";
    SELECT state_h, request_json, polling_code, code_verifier
        FROM AuthInfo
        WHERE state_h = STATE
          AND expires_at >= CURRENT_TIMESTAMP();
END;;

CREATE OR REPLACE PROCEDURE AuthInfo_Insert(IN STATE_H_ VARCHAR(128), IN REQUEST LONGTEXT,
                                            IN EXPIRES_IN_ INT, IN POLLING_CODE_ BIT)
BEGIN
    SET TIME_ZONE = "+0:00";
    INSERT INTO AuthInfo (`state_h`, `request_json`, `expires_in`, `polling_code`)
        VALUES (STATE_H_, REQUEST, EXPIRES_IN_, POLLING_CODE_);
END;;

CREATE OR REPLACE PROCEDURE AuthInfo_SetCodeVerifier(IN STATE TEXT, IN VERIFIER TEXT)
BEGIN
    SET TIME_ZONE = "+0:00";
    UPDATE AuthInfo SET code_verifier = VERIFIER WHERE state_h = STATE;
END;;

CREATE OR REPLACE PROCEDURE AuthInfo_Update(IN STATE TEXT, IN REQUEST LONGTEXT)
BEGIN
    SET TIME_ZONE = "+0:00";
    UPDATE AuthInfo SET request_json = REQUEST WHERE state_h = STATE;
END;;

CREATE OR REPLACE PROCEDURE Cleanup_AuthInfo()
BEGIN
    SET TIME_ZONE = "+0:00";
    DELETE FROM AuthInfo WHERE expires_at < CURRENT_TIMESTAMP();
END;;

CREATE OR REPLACE PROCEDURE Cleanup_MTokens()
BEGIN
    SET TIME_ZONE = "+0:00";
    DELETE FROM MTokens WHERE DATE_ADD(expires_at, INTERVAL 1 MONTH) < CURRENT_TIMESTAMP();
END;;

CREATE OR REPLACE PROCEDURE Cleanup_ProxyTokens()
BEGIN
    SET TIME_ZONE = "+0:00";
    DELETE
        FROM ProxyTokens
        WHERE id IN (SELECT id
                         FROM TransferCodesAttributes
                         WHERE DATE_ADD(expires_at, INTERVAL 1 MONTH) < CURRENT_TIMESTAMP());
END;;

CREATE OR REPLACE PROCEDURE CryptStore_Update(IN CRYPTID BIGINT UNSIGNED, IN VALUE TEXT)
BEGIN
    SET TIME_ZONE = "+0:00";
    UPDATE CryptStore SET crypt = VALUE WHERE id = CRYPTID;
END;;

CREATE OR REPLACE PROCEDURE EncryptionKeysRT_Insert(IN CRYPTKEY TEXT, IN RTID BIGINT UNSIGNED, IN MTID VARCHAR(128))
BEGIN
    SET TIME_ZONE = "+0:00";
    CALL EncryptionKeys_Insert(CRYPTKEY, @KEYID);
    INSERT INTO RT_EncryptionKeys (rt_id, MT_id, key_id) VALUES (RTID, MTID, @KEYID);
END;;

CREATE OR REPLACE PROCEDURE EncryptionKeys_Insert(IN CRYPTKEY TEXT, OUT ID BIGINT UNSIGNED)
BEGIN
    SET TIME_ZONE = "+0:00";
    INSERT INTO EncryptionKeys (encryption_key) VALUES (CRYPTKEY);
    SET ID = (SELECT LAST_INSERT_ID());
END;;

CREATE OR REPLACE PROCEDURE EncryptionKeys_Update(IN KEYID BIGINT UNSIGNED, IN NEW_KEY TEXT)
BEGIN
    SET TIME_ZONE = "+0:00";
    UPDATE EncryptionKeys SET encryption_key = NEW_KEY WHERE id = KEYID;
END;;

CREATE OR REPLACE PROCEDURE Event_Insert(IN MTID VARCHAR(128), IN EVENT TEXT, IN COMMENT TEXT,
                                         IN IP TEXT, IN USERAGENT TEXT)
BEGIN
    SET TIME_ZONE = "+0:00";
    INSERT INTO MT_Events (MT_id, event_id, comment, ip, user_agent)
        VALUES (MTID, (SELECT e.id FROM Events e WHERE e.event = EVENT), COMMENT, IP, USERAGENT);
END;;

CREATE OR REPLACE PROCEDURE InsertProfileTemplate(IN TYPE_ VARCHAR(100), IN ID_ VARCHAR(128),
                                                  IN GROUP_ VARCHAR(64), IN NAME_ VARCHAR(128),
                                                  IN PAYLOAD_ LONGTEXT)
BEGIN
    SET TIME_ZONE = "+0:00";
    INSERT INTO ServerProfiles (id, type, `group`, name, payload)
        VALUES (ID_, (SELECT id
                          FROM ProfileTypes pt
                          WHERE pt.`type` = TYPE_), GROUP_, NAME_, PAYLOAD_);
END;;

CREATE OR REPLACE PROCEDURE MTokens_CheckRotating(IN MTID VARCHAR(128), IN SEQ BIGINT UNSIGNED, IN LIFETIME INT)
BEGIN
    SET TIME_ZONE = "+0:00";
    SELECT COUNT(1)
        FROM MTokens
        WHERE id = MTID
          AND seqno = SEQ
          AND TIMESTAMPADD(SECOND, LIFETIME, last_rotated) >= CURRENT_TIMESTAMP();
END;;

CREATE OR REPLACE PROCEDURE MTokens_Insert(IN SUB TEXT, IN ISS TEXT, IN MTID VARCHAR(128),
                                           IN SEQNO_ BIGINT UNSIGNED, IN PARENT VARCHAR(128),
                                           IN RTID BIGINT UNSIGNED, IN NAME_ TEXT, IN IP TEXT,
                                           IN EXPIRES_AT_ DATETIME)
BEGIN
    SET TIME_ZONE = "+0:00";
    CALL Users_GetID(SUB, ISS, @UID);
    INSERT INTO MTokens (id, seqno, parent_id, rt_id, name, ip_created, user_id, expires_at)
        VALUES (MTID, SEQNO_, PARENT, RTID, NAME_, IP, @UID, EXPIRES_AT_);
END;;

CREATE OR REPLACE PROCEDURE MTokens_UpdateSeqNo(IN MTID VARCHAR(128), IN SEQNO BIGINT UNSIGNED)
BEGIN
    SET TIME_ZONE = "+0:00";
    UPDATE MTokens SET seqno=SEQNO, last_rotated = CURRENT_TIMESTAMP() WHERE id = MTID;
END;;

CREATE OR REPLACE PROCEDURE ProxyTokens_Insert(IN PTID VARCHAR(128), IN JWT TEXT, IN MTID VARCHAR(128))
BEGIN
    SET TIME_ZONE = "+0:00";
    CALL CryptStoreMT_Insert(JWT, @JWTID);
    INSERT INTO ProxyTokens (id, jwt_crypt, MT_id) VALUES (PTID, @JWTID, MTID);
END;;

CREATE OR REPLACE PROCEDURE ProxyTokens_Update(IN PTID VARCHAR(128), IN JWT TEXT, IN MTID VARCHAR(128))
BEGIN
    DECLARE jwtID BIGINT UNSIGNED;
    SET TIME_ZONE = "+0:00";
    SELECT pt.jwt_crypt FROM ProxyTokens pt WHERE pt.id = PTID INTO jwtID;
    IF (jwtID IS NULL) THEN
        CALL CryptStoreMT_Insert(JWT, @JWTID);
        UPDATE ProxyTokens SET MT_id=MTID, jwt_crypt=@JWTID WHERE id = PTID;
    ELSE
        UPDATE CryptStore SET crypt=JWT WHERE id = jwtID;
        UPDATE ProxyTokens SET MT_id=MTID WHERE id = PTID;
    END IF;
END;;

CREATE OR REPLACE PROCEDURE SSHInfo_Insert(IN MTID VARCHAR(128), IN KEY_FP VARCHAR(128),
                                           IN SSH_USER_H VARCHAR(128), IN NAME TEXT,
                                           IN ENCRYPTED_MT TEXT)
BEGIN
    SET TIME_ZONE = "+0:00";
    CALL CryptStoreMT_Insert(ENCRYPTED_MT, @CRYPT_ID);
    INSERT INTO SSHPublicKeys (user, ssh_key_fp, ssh_user_hash, name, MT_crypt, MT_id)
        VALUES ((SELECT m.user_id FROM MTokens m WHERE m.id = MTID), KEY_FP, SSH_USER_H, NAME, @CRYPT_ID, MTID);
END;;

CREATE OR REPLACE PROCEDURE SSHInfo_UsedKey(IN KEY_FP VARCHAR(128), IN USER_H VARCHAR(128))
BEGIN
    SET TIME_ZONE = "+0:00";
    UPDATE SSHPublicKeys SET last_used = CURRENT_TIMESTAMP() WHERE ssh_key_fp = KEY_FP AND ssh_user_hash = USER_H;
END;;

CREATE OR REPLACE PROCEDURE TransferCodeAttributes_Insert(IN TCID VARCHAR(128), IN EXPIRES_IN INT,
                                                          IN REVOKE_MT BIT, IN RESPONSE_TYPE TEXT,
                                                          IN MAX_TOKEN_LEN INT)
BEGIN
    SET TIME_ZONE = "+0:00";
    INSERT INTO TransferCodesAttributes (id, expires_in, revoke_MT, response_type, max_token_len)
        VALUES (TCID, EXPIRES_IN, REVOKE_MT, RESPONSE_TYPE, MAX_TOKEN_LEN);
END;;

CREATE OR REPLACE PROCEDURE TransferCodeAttributes_UpdateSSHKey(IN PCID VARCHAR(128), IN KEY_FP VARCHAR(128))
BEGIN
    SET TIME_ZONE = "+0:00";
    UPDATE TransferCodesAttributes SET ssh_key_fp=KEY_FP WHERE id = PCID;
END;;

CREATE OR REPLACE PROCEDURE TransferCodes_GetStatus(IN PCID VARCHAR(128))
BEGIN
    SET TIME_ZONE = "+0:00";
    SELECT 1                                AS found,
           CURRENT_TIMESTAMP() > expires_at AS expired,
           response_type,
           consent_declined,
           max_token_len,
           ssh_key_fp
        FROM TransferCodes
        WHERE id = PCID;
END;;

CREATE OR REPLACE PROCEDURE UpdateProfileTemplate(IN TYPE_ VARCHAR(100), IN ID_ VARCHAR(128),
                                                  IN GROUP_ VARCHAR(64), IN NAME_ VARCHAR(128),
                                                  IN PAYLOAD_ LONGTEXT)
BEGIN
    SET TIME_ZONE = "+0:00";
    UPDATE ServerProfiles sp
    SET sp.name   = NAME_,
        sp.payload=PAYLOAD_
        WHERE sp.type = (SELECT id
                             FROM ProfileTypes pt
                             WHERE pt.`type` = TYPE_)
          AND sp.id = ID_
          AND sp.group = GROUP_;
END;;

CREATE OR REPLACE PROCEDURE Users_Insert(IN SUB TEXT, IN ISS TEXT, OUT UID BIGINT UNSIGNED)
BEGIN
    SET TIME_ZONE = "+0:00";
    INSERT INTO Users (sub, iss) VALUES (SUB, ISS);
    SET UID = (SELECT LAST_INSERT_ID());
END;;

CREATE OR REPLACE PROCEDURE Version_SetAfter(IN VERSION TEXT)
BEGIN
    SET TIME_ZONE = "+0:00";
    INSERT INTO version (version, aft)
        VALUES (VERSION, CURRENT_TIMESTAMP())
    ON DUPLICATE KEY UPDATE aft=CURRENT_TIMESTAMP();
END;;

CREATE OR REPLACE PROCEDURE Version_SetBefore(IN VERSION TEXT)
BEGIN
    SET TIME_ZONE = "+0:00";
    INSERT INTO version (version, bef)
        VALUES (VERSION, CURRENT_TIMESTAMP())
    ON DUPLICATE KEY UPDATE bef=CURRENT_TIMESTAMP();
END;;

DELIMITER ;