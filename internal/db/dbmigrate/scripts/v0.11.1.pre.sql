DELIMITER ;;

CREATE OR REPLACE PROCEDURE Tags_UpdateColor(IN MTID VARCHAR(128), IN p_TAG VARCHAR(64), IN p_COLOR CHAR(6))
BEGIN
    UPDATE Tags
    SET color=TagColor(p_TAG, p_COLOR)
        WHERE uid = (SELECT user_id FROM MTokens WHERE id = MTID)
          AND tag = p_TAG;
END;;

CREATE OR REPLACE PROCEDURE Tags_UpdateName(IN MTID VARCHAR(128), IN OLD_TAG VARCHAR(64), IN NEW_TAG VARCHAR(64))
BEGIN
    UPDATE Tags
    SET tag=NEW_TAG
        WHERE uid = (SELECT user_id FROM MTokens WHERE id = MTID)
          AND tag = OLD_TAG;
END;;

CREATE OR REPLACE PROCEDURE Notifications_GetTagSubscribedMTsForNotification(IN NID BIGINT UNSIGNED)
BEGIN
    -- Get all MTs that are subscribed to this notification via tag matching
    -- (not including directly subscribed MTs)
    SELECT DISTINCT mt.MT_id
        FROM MTTags mt
                 JOIN NotificationTags nt ON mt.tag_id = nt.tag_id
        WHERE nt.notification_id = NID;
END;;

DELIMITER ;
