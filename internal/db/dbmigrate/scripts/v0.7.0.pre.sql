# Predefined values
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

