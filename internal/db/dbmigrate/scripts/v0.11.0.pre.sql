# Tables
ALTER TABLE version
    ADD IF NOT EXISTS go DATETIME DEFAULT NULL NULL;

CREATE OR REPLACE TABLE Tags
(
    id    BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    tag   VARCHAR(64)     NOT NULL,
    uid   BIGINT UNSIGNED NOT NULL,
    color CHAR(6)         NOT NULL,
    CONSTRAINT Tags_UNIQUE
        UNIQUE (tag, uid),
    CONSTRAINT Tags_Users_FK
        FOREIGN KEY (uid) REFERENCES Users (id)
            ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE OR REPLACE TABLE CalendarTags
(
    calendar_id VARCHAR(128) NOT NULL, tag_id BIGINT UNSIGNED NOT NULL, CONSTRAINT
    CalendarTags_UNIQUE
    UNIQUE (calendar_id, tag_id), CONSTRAINT CalendarTags_Calendars_FK
        FOREIGN KEY (calendar_id) REFERENCES Calendars (id)
            ON UPDATE CASCADE ON DELETE CASCADE, CONSTRAINT CalendarTags_Tags_FK
        FOREIGN KEY (tag_id) REFERENCES Tags (id)
            ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE OR REPLACE TABLE MTTags
(
    MT_id            VARCHAR(128)    NOT NULL,
    tag_id           BIGINT UNSIGNED NOT NULL,
    include_children BIT             NOT NULL,
    CONSTRAINT MTTags_UNIQUE
        UNIQUE (MT_id, tag_id),
    CONSTRAINT MTTags_Tags_FK
        FOREIGN KEY (tag_id) REFERENCES Tags (id)
            ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT Tags_MTokens_FK
        FOREIGN KEY (MT_id) REFERENCES MTokens (id)
            ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE OR REPLACE TABLE NotificationTags
(
    notification_id BIGINT UNSIGNED NOT NULL, tag_id BIGINT UNSIGNED NOT NULL, CONSTRAINT NotificationTags_UNIQUE
    UNIQUE (notification_id, tag_id), CONSTRAINT NotificationTags_Notifications_FK
        FOREIGN KEY (notification_id) REFERENCES Notifications (id)
            ON UPDATE CASCADE ON DELETE CASCADE, CONSTRAINT NotificationTags_Tags_FK
        FOREIGN KEY (tag_id) REFERENCES Tags (id)
            ON UPDATE CASCADE ON DELETE CASCADE
);


# Move Calendar names into tag
ALTER TABLE Calendars
    DROP KEY IF EXISTS Calendars_UN_1;
ALTER TABLE Calendars
    DROP KEY IF EXISTS Calendars_UN;
ALTER TABLE Calendars
    DROP COLUMN IF EXISTS ics_path;
ALTER TABLE Calendars
    ADD IF NOT EXISTS description TEXT NULL;


## PROCEDURES
DROP PROCEDURE IF EXISTS Calendar_Get;

DELIMITER ;;

CREATE OR REPLACE PROCEDURE Calendar_ClearTags(IN CID VARCHAR(128))
BEGIN
    DELETE FROM CalendarTags WHERE calendar_id = CID;
END;;

CREATE OR REPLACE PROCEDURE Calendar_IDForSameUserAsMT(IN CID VARCHAR(128), IN MTID VARCHAR(128))
BEGIN
    SELECT COUNT(*) FROM MTokens WHERE id = MTID AND user_id = (SELECT uid FROM Calendars WHERE id = CID);
END;;

CREATE OR REPLACE PROCEDURE Calendar_LinkTag(IN CID VARCHAR(128), IN TAG_ VARCHAR(64))
BEGIN
    DECLARE v_uid BIGINT UNSIGNED;
    DECLARE v_tid BIGINT UNSIGNED;

    SELECT uid INTO v_uid FROM Calendars WHERE id = CID;
    CALL Tags_Assert(v_uid, TAG_, NULL, v_tid);

    INSERT IGNORE INTO CalendarTags (calendar_id, tag_id) VALUES (CID, v_tid);
END;;


CREATE OR REPLACE PROCEDURE Calendar_UnlinkTag(IN CID VARCHAR(128), IN TAG_ VARCHAR(64))
BEGIN
    DECLARE v_uid BIGINT UNSIGNED;
    DECLARE v_tid BIGINT UNSIGNED;

    SELECT uid INTO v_uid FROM Calendars WHERE id = CID;
    SELECT id INTO v_tid FROM Tags WHERE tag = TAG_ AND uid = v_uid;

    DELETE FROM CalendarTags WHERE calendar_id = CID AND tag_id = v_tid;
END;;

CREATE OR REPLACE PROCEDURE Calendar_Delete(IN MTID VARCHAR(128), IN CID VARCHAR(128))
BEGIN
    DELETE FROM Calendars WHERE uid = (SELECT m.user_id FROM MTokens m WHERE m.id = MTID) AND id = CID;
END;;

CREATE OR REPLACE PROCEDURE Calendar_GetByID(IN CID VARCHAR(128))
BEGIN
    SELECT id, ics, description FROM Calendars WHERE id = CID;
END;;

CREATE OR REPLACE PROCEDURE Calendar_GetMTsInCalendar(IN CALID VARCHAR(128))
BEGIN
    SELECT DISTINCT MT_id
        FROM (SELECT cm.MT_id
                  FROM CalendarMapping cm
                  WHERE cm.calendar_id = CALID
              UNION ALL
              SELECT mt.MT_id
                  FROM MTTags mt
                           JOIN CalendarTags ct ON mt.tag_id = ct.tag_id
                  WHERE ct.calendar_id = CALID) t;
END;;

CREATE OR REPLACE PROCEDURE Calendar_Insert(IN MTID VARCHAR(128), IN CID VARCHAR(128), IN DESCR TEXT,
                                            IN ICS_ LONGTEXT)
BEGIN
    SET TIME_ZONE = "+0:00";
    INSERT INTO Calendars (id, uid, description, ics)
        VALUES (CID, (SELECT m.user_id FROM MTokens m WHERE m.id = MTID), DESCR, ICS_);
END;;


CREATE OR REPLACE PROCEDURE Calendar_List(IN MTID VARCHAR(128))
BEGIN
    SELECT id, ics, description
        FROM Calendars
        WHERE uid = (SELECT m.user_id
                         FROM MTokens m
                         WHERE m.id = MTID);
END;;

CREATE OR REPLACE PROCEDURE Calendar_GetTags(IN CID VARCHAR(128))
BEGIN
    SELECT t.tag, t.color FROM Tags t WHERE id IN (SELECT c.tag_id FROM CalendarTags c WHERE c.calendar_id = CID);
END;;

CREATE OR REPLACE PROCEDURE Calendar_ListForMT(IN MTID VARCHAR(128))
BEGIN
    SELECT id, ics, description
        FROM Calendars
        WHERE id IN (SELECT calendar_id FROM CalendarMapping WHERE MT_id = MTID);
END;;

DROP PROCEDURE IF EXISTS Calendar_Update;
DROP PROCEDURE IF EXISTS Calendar_UpdateInternal;

CREATE OR REPLACE PROCEDURE Calendar_UpdateICS(IN MTID VARCHAR(128), IN CID
    VARCHAR(128), IN ICS_ LONGTEXT)
BEGIN
    SET TIME_ZONE = "+0:00";
    UPDATE Calendars
    SET ics=ICS_
        WHERE uid = (SELECT m.user_id FROM MTokens m WHERE m.id = MTID)
          AND id = CID;
END;;

CREATE OR REPLACE PROCEDURE Calendar_UpdateDescription(IN MTID VARCHAR(128), IN
    CID VARCHAR(128), IN DESCR TEXT)
BEGIN
    SET TIME_ZONE = "+0:00";
    UPDATE Calendars
    SET description=DESCR
        WHERE uid = (SELECT m.user_id FROM MTokens m WHERE m.id = MTID)
          AND id = CID;
END;;

CREATE OR REPLACE PROCEDURE Calendar_UpdateICSInternal(IN CID VARCHAR(128), IN ICS_ LONGTEXT)
BEGIN
    SET TIME_ZONE = "+0:00";
    UPDATE Calendars SET ics=ICS_ WHERE id = CID;
END;;


CREATE OR REPLACE FUNCTION TagColor(p_tag VARCHAR(64), p_color CHAR(6)) RETURNS CHAR(6)
    DETERMINISTIC
BEGIN
    -- If a valid color is provided, return it
    IF p_color IS NOT NULL AND p_color <> '' THEN
        RETURN p_color;
    END IF;

    -- Generate and return the computed color
    RETURN LPAD(HEX(CRC32(p_tag)), 6, '0');
END;;

CREATE OR REPLACE PROCEDURE Tags_Assert(IN p_uid BIGINT UNSIGNED, IN p_tag VARCHAR(64), IN p_color CHAR(6),
                                        OUT p_tag_id BIGINT UNSIGNED)
BEGIN

    INSERT INTO Tags (tag, uid, color)
        VALUES (p_tag, p_uid, TagColor(p_tag, p_color))
    ON DUPLICATE KEY UPDATE id = LAST_INSERT_ID(id);

    SET p_tag_id = LAST_INSERT_ID();
END;;

CREATE OR REPLACE PROCEDURE Tags_Create(IN MTID VARCHAR(128), IN NAME VARCHAR(64), IN COLOR_ CHAR(6))
BEGIN
    INSERT IGNORE INTO Tags (uid, tag, color)
        VALUES ((SELECT user_id FROM MTokens WHERE id = MTID), NAME, TagColor(NAME,
                                                                              COLOR_));
END;;


CREATE OR REPLACE PROCEDURE Tags_Delete(IN MTID VARCHAR(128), IN p_TAG VARCHAR(64))
BEGIN
    DELETE FROM Tags WHERE uid = (SELECT user_id FROM MTokens WHERE id = MTID) AND tag = p_TAG;
END;;

CREATE OR REPLACE PROCEDURE Tags_List(IN MTID VARCHAR(128))
BEGIN
    SELECT tag, color FROM Tags WHERE uid = (SELECT user_id FROM MTokens WHERE id = MTID);
END;;

CREATE OR REPLACE PROCEDURE Tags_Update(IN MTID VARCHAR(128), IN OLD_TAG VARCHAR(64), IN NEW_TAG VARCHAR(64),
                                        IN p_COLOR CHAR(6))
BEGIN
    UPDATE Tags
    SET tag=NEW_TAG, color=p_COLOR
        WHERE uid = (SELECT user_id FROM MTokens WHERE id = MTID) AND tag = OLD_TAG;
END;;

CREATE OR REPLACE PROCEDURE MTokens_LinkTag(IN MTID VARCHAR(128), IN TAG_ VARCHAR(64), IN INCLUDE_CHILDREN_ BIT)
BEGIN
    DECLARE v_uid BIGINT UNSIGNED;
    DECLARE v_tid BIGINT UNSIGNED;

    SELECT user_id INTO v_uid FROM MTokens WHERE id = MTID;
    CALL Tags_Assert(v_uid, TAG_, NULL, v_tid);

    INSERT IGNORE INTO MTTags (MT_id, tag_id, include_children)
        VALUES (MTID, v_tid, INCLUDE_CHILDREN_);
END;;

CREATE OR REPLACE PROCEDURE MTokens_UnlinkTag(IN MTID VARCHAR(128), IN TAG_ VARCHAR(64))
BEGIN
    DECLARE v_uid BIGINT UNSIGNED;
    DECLARE v_tid BIGINT UNSIGNED;

    SELECT user_id INTO v_uid FROM MTokens WHERE id = MTID;
    SELECT id INTO v_tid FROM Tags WHERE tag = TAG_ AND uid = v_uid;

    DELETE FROM MTTags WHERE MT_id = MTID AND tag_id = v_tid;
END;;

CREATE OR REPLACE PROCEDURE MTokens_ClearTags(IN MTID VARCHAR(128))
BEGIN
    DELETE FROM MTTags WHERE MT_id = MTID;
END;;

CREATE OR REPLACE PROCEDURE MTokens_GetTags(IN MTID VARCHAR(128))
BEGIN
    SELECT t.id AS tag_id, t.tag, t.color AS tag_color, mt.include_children AS tag_include_children
        FROM Tags t
                 JOIN MTTags mt ON t.id = mt.tag_id
        WHERE mt.MT_id = MTID;
END;;

CREATE OR REPLACE PROCEDURE MTokens_GetChildren(IN MTID VARCHAR(128))
BEGIN
    SELECT id, parent_id, name, created, expires_at, ip_created AS ip
        FROM MTokens
        WHERE parent_id = MTID;
END;;

CREATE OR REPLACE PROCEDURE MTTag_GetIncludeChildren(IN MTID VARCHAR(128), IN TAG_ VARCHAR(64))
BEGIN
    DECLARE v_uid BIGINT UNSIGNED;
    DECLARE v_tid BIGINT UNSIGNED;

    SELECT user_id INTO v_uid FROM MTokens WHERE id = MTID;
    SELECT id INTO v_tid FROM Tags WHERE tag = TAG_ AND uid = v_uid;

    SELECT include_children FROM MTTags WHERE MT_id = MTID AND tag_id = v_tid;
END;;


CREATE OR REPLACE PROCEDURE ActionCodes_AddRemoveFromCalendar(IN MTID VARCHAR(128), IN CALENDARID VARCHAR(128),
                                                              IN CODE_ VARCHAR(128))
BEGIN
    DECLARE aid BIGINT UNSIGNED;
    DECLARE id BIGINT UNSIGNED;
    SET TIME_ZONE = "+0:00";
    SELECT a.id FROM Actions a WHERE a.`action` = 'remove_from_calendar' INTO aid;
    INSERT INTO ActionCodes (action, code) VALUES (aid, CODE_);
    SELECT LAST_INSERT_ID() INTO id;
    INSERT INTO ActionReferencesCalendarEntries (action_id, calendar_mapping_id)
        VALUES (id, (SELECT cm.mapping_id
                         FROM CalendarMapping cm
                         WHERE cm.MT_id = MTID
                           AND cm.calendar_id = CALENDARID));
END;;

CREATE OR REPLACE PROCEDURE Version_Get()
BEGIN
    SELECT version, bef, go, aft FROM version;
END;;

CREATE OR REPLACE PROCEDURE Version_SetGo(IN VERSION TEXT)
BEGIN
    SET TIME_ZONE = "+0:00";
    INSERT INTO version (version, go)
        VALUES (VERSION, CURRENT_TIMESTAMP())
    ON DUPLICATE KEY UPDATE go=CURRENT_TIMESTAMP();
END;;

CREATE OR REPLACE PROCEDURE MTokens_GetForUser(IN UID BIGINT UNSIGNED)
BEGIN
    SELECT m.id,
           m.parent_id,
           m.id         AS mom_id,
           m.name,
           m.created,
           m.expires_at,
           m.ip_created AS ip,
           t.tag,
           t.color      AS
                           tag_color,
           mt.include_children
                        AS tag_include_children
        FROM MTokens m
                 LEFT JOIN MTTags mt ON m.id = mt.MT_id
                 LEFT JOIN Tags t ON mt.tag_id = t.id
        WHERE m.user_id = UID
        ORDER BY created;
END;;

CREATE OR REPLACE PROCEDURE Mtokens_GetInfo(IN MTID VARCHAR(128))
BEGIN
    SELECT m.id,
           m.parent_id,
           m.id         AS mom_id,
           m.name,
           m.created,
           m.expires_at,
           m.ip_created AS ip,
           t.tag,
           t.color      AS
                           tag_color,
           mt.include_children
                        AS tag_include_children
        FROM MTokens m
                 LEFT JOIN MTTags mt ON m.id = mt.MT_id
                 LEFT JOIN Tags t ON mt.tag_id = t.id
        WHERE m.id = MTID;

END;;


DELIMITER ;

# Values

INSERT IGNORE INTO Events (event)
    VALUES ('tags_listed');
INSERT IGNORE INTO Events (event)
    VALUES ('tag_created');
INSERT IGNORE INTO Events (event)
    VALUES ('tag_updated');
INSERT IGNORE INTO Events (event)
    VALUES ('tag_deleted');
INSERT IGNORE INTO Events (event)
    VALUES ('tag_added_to_token');
INSERT IGNORE INTO Events (event)
    VALUES ('tag_added_to_other_token');
INSERT IGNORE INTO Events (event)
    VALUES ('tag_removed_from_token');
INSERT IGNORE INTO Events (event)
    VALUES ('tag_removed_from_other_token');
INSERT IGNORE INTO Events (event)
    VALUES ('calendar_tags_updated');
INSERT IGNORE INTO Events (event)
    VALUES ('notification_tags_updated');
