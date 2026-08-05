# Tables
ALTER TABLE AccessTokens
    ADD IF NOT EXISTS expires_at DATETIME NULL;
ALTER TABLE AccessTokens
    ADD IF NOT EXISTS token_type VARCHAR(16) NULL;


# Procedures
DELIMITER ;;

CREATE OR REPLACE PROCEDURE AT_Insert(IN AT TEXT, IN IP TEXT, IN COMM TEXT, IN MT VARCHAR(128),
                                      IN EXPIRESAT DATETIME, IN TOKENTYPE VARCHAR(16))
BEGIN
    SET TIME_ZONE = "+0:00";
    CALL CryptStoreAT_Insert(AT, @ATCryptID);
    INSERT INTO AccessTokens (token_crypt, ip_created, comment, MT_id, expires_at, token_type)
        VALUES (@ATCryptID, IP, COMM, MT, EXPIRESAT, TOKENTYPE);
    SELECT LAST_INSERT_ID();
END;;

CREATE OR REPLACE PROCEDURE AT_GetCached(IN MTID VARCHAR(128), IN SCOPES_JSON TEXT, IN AUDS_JSON TEXT)
BEGIN
    SET TIME_ZONE = "+0:00";
    SELECT cs.crypt, cs.created, at.expires_at, at.token_type
        FROM AccessTokens at
                 JOIN CryptStore cs ON at.token_crypt = cs.id
        WHERE at.MT_id = MTID
          AND at.expires_at IS NOT NULL
          AND at.expires_at > CURRENT_TIMESTAMP()
          AND NOT EXISTS (SELECT 1
                              FROM AT_Attributes aata
                                       JOIN Attributes aa ON aata.attribute_id = aa.id
                              WHERE aa.attribute = 'scope'
                                AND aata.AT_id = at.id
                                AND NOT JSON_CONTAINS(SCOPES_JSON, JSON_QUOTE(aata.`attribute`)))
          AND (SELECT COUNT(DISTINCT aata.`attribute`)
                   FROM AT_Attributes aata
                            JOIN Attributes aa ON aata.attribute_id = aa.id
                   WHERE aa.attribute = 'scope'
                     AND aata.AT_id = at.id) = JSON_LENGTH(SCOPES_JSON)
          AND NOT EXISTS (SELECT 1
                              FROM AT_Attributes aata
                                       JOIN Attributes aa ON aata.attribute_id = aa.id
                              WHERE aa.attribute = 'audience'
                                AND aata.AT_id = at.id
                                AND NOT JSON_CONTAINS(AUDS_JSON, JSON_QUOTE(aata.`attribute`)))
          AND (SELECT COUNT(DISTINCT aata.`attribute`)
                   FROM AT_Attributes aata
                            JOIN Attributes aa ON aata.attribute_id = aa.id
                   WHERE aa.attribute = 'audience'
                     AND aata.AT_id = at.id) = JSON_LENGTH(AUDS_JSON)
        ORDER BY at.id DESC
        LIMIT 1;
END;;

CREATE OR REPLACE PROCEDURE Cleanup_AccessTokens()
BEGIN
    SET TIME_ZONE = "+0:00";
    DELETE
        FROM AccessTokens
        WHERE expires_at IS NOT NULL
          AND expires_at < CURRENT_TIMESTAMP();
END;;

CREATE OR REPLACE PROCEDURE Cleanup()
BEGIN
    CALL Cleanup_MTokens();
    CALL Cleanup_AuthInfo();
    CALL Cleanup_ProxyTokens();
    CALL Cleanup_ActionCodes();
    CALL Cleanup_AccessTokens();
END;;

DELIMITER ;
