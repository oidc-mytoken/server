# Procedures
DELIMITER ;;
CREATE OR REPLACE PROCEDURE MTokens_GetForUser(IN UID BIGINT UNSIGNED)
BEGIN
    SELECT id, parent_id, id AS revocation_id, name, created, ip_created AS ip
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
    SELECT m.id, m.parent_id, m.id AS revocation_id, m.name, m.created, m.ip_created AS ip
        FROM MTokens m
        WHERE m.id IN
              (SELECT id
                   FROM effected_MTIDs);
    DROP TABLE effected_MTIDs;
END;;

CREATE OR REPLACE PROCEDURE MTokens_IsParentOf(IN PARENT VARCHAR(128), IN CHILD VARCHAR(128))
BEGIN
    CREATE TEMPORARY TABLE IF NOT EXISTS parent_MTIDs (id VARCHAR(128));
    TRUNCATE parent_MTIDs;
    INSERT INTO parent_MTIDs
    WITH RECURSIVE parents AS (SELECT id, parent_id
                                   FROM MTokens
                                   WHERE id = CHILD
                               UNION ALL
                               SELECT mt.id, mt.parent_id
                                   FROM MTokens mt
                                            INNER JOIN parents p
                                   WHERE mt.id = p.parent_id)
    SELECT id
        FROM parents;
    SELECT COUNT(1) FROM parent_MTIDs WHERE id = PARENT;
    DROP TABLE parent_MTIDs;
END;;
DELIMITER ;

