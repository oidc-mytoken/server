### Tables

CREATE TABLE IF NOT EXISTS Actions
(
    id                      INT UNSIGNED AUTO_INCREMENT
        PRIMARY KEY, action VARCHAR(128) NOT NULL, CONSTRAINT Actions_UN
        UNIQUE (action)
);

CREATE TABLE IF NOT EXISTS ActionCodes
(
    id                      BIGINT UNSIGNED AUTO_INCREMENT
        PRIMARY KEY,
    action                  INT UNSIGNED NOT NULL,
    code                    VARCHAR(128) NOT NULL,
    expires_at              DATETIME     NULL,
    CONSTRAINT ActionCodes_UN
        UNIQUE (code),
    CONSTRAINT ActionCodes_FK
        FOREIGN KEY (action) REFERENCES Actions (id)
);

CREATE INDEX IF NOT EXISTS AuthInfo_FK
    ON AuthInfo (polling_code);

ALTER TABLE Users
    ADD email TEXT NULL;
ALTER TABLE Users
    ADD email_verified BOOL DEFAULT 0 NOT NULL;
ALTER TABLE Users
    ADD prefer_html_mail BOOL DEFAULT 1 NOT NULL;

CREATE TABLE IF NOT EXISTS ActionReferencesUser
(
    action_id BIGINT UNSIGNED NOT NULL, uid BIGINT UNSIGNED NOT NULL, CONSTRAINT ActionReferencesUser_FK
    FOREIGN KEY (uid) REFERENCES Users (id)
        ON UPDATE CASCADE ON DELETE CASCADE, CONSTRAINT ActionReferencesUser_FK_1
        FOREIGN KEY (action_id) REFERENCES ActionCodes (id)
            ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS Calendars
(
    id                    VARCHAR(128)    NOT NULL
        PRIMARY KEY, name                                    VARCHAR(128) NOT NULL,
    uid                   BIGINT UNSIGNED NOT NULL, ics_path VARCHAR(128) NOT NULL, ics LONGTEXT NOT NULL,
    CONSTRAINT Calendars_UN
        UNIQUE (ics_path),
    CONSTRAINT Calendars_UN_1
        UNIQUE (name, uid),
    CONSTRAINT Calendars_FK
        FOREIGN KEY (uid) REFERENCES Users (id)
            ON UPDATE CASCADE ON DELETE CASCADE
);

ALTER TABLE MTokens
    ADD capabilities JSON NULL;
ALTER TABLE MTokens
    ADD rotation JSON NULL;
ALTER TABLE MTokens
    ADD restrictions JSON NULL;

CREATE TABLE IF NOT EXISTS ActionReferencesMytokens
(
    action_id BIGINT UNSIGNED NOT NULL, MT_id VARCHAR(128) NOT NULL, CONSTRAINT ActionReferencesMytokens_FK
    FOREIGN KEY (action_id) REFERENCES ActionCodes (id)
        ON UPDATE CASCADE ON DELETE CASCADE, CONSTRAINT ActionReferencesMytokens_FK_1
        FOREIGN KEY (MT_id) REFERENCES MTokens (id)
            ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS CalendarMapping
(
    calendar_id VARCHAR(128) NOT NULL, MT_id VARCHAR(128) NOT NULL, mapping_id BIGINT UNSIGNED AUTO_INCREMENT
    PRIMARY KEY, CONSTRAINT CalendarMapping_FK
        FOREIGN KEY (MT_id) REFERENCES MTokens (id)
            ON UPDATE CASCADE ON DELETE CASCADE, CONSTRAINT CalendarMapping_FK_1
        FOREIGN KEY (calendar_id) REFERENCES Calendars (id)
            ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS ActionReferencesCalendarEntries
(
    action_id           BIGINT UNSIGNED NOT NULL,
    calendar_mapping_id BIGINT UNSIGNED NOT NULL,
    CONSTRAINT ActionReferencesCalendarEntries_FK
        FOREIGN KEY (action_id) REFERENCES ActionCodes (id)
            ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT ActionReferencesCalendarEntries_FK_1
        FOREIGN KEY (calendar_mapping_id) REFERENCES CalendarMapping (mapping_id)
            ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS Notifications
(
    id                    BIGINT UNSIGNED AUTO_INCREMENT
        PRIMARY KEY,
    type                  VARCHAR(32)          NOT NULL,
    management_code       VARCHAR(128)         NOT NULL,
    ws                    VARCHAR(128)         NULL,
    user_wide             TINYINT(1) DEFAULT 0 NOT NULL,
    uid                   BIGINT UNSIGNED      NOT NULL,
    CONSTRAINT Notifications_pk2
        UNIQUE (management_code),
    CONSTRAINT Notifications_FK
        FOREIGN KEY (uid) REFERENCES Users (id)
            ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS MTNotificationsMapping
(
    MT_id            VARCHAR(128)         NOT NULL,
    notification_id  BIGINT UNSIGNED      NOT NULL,
    include_children TINYINT(1) DEFAULT 1 NOT NULL,
    CONSTRAINT MTNotificationsMapping_pk
        UNIQUE (notification_id, MT_id),
    CONSTRAINT MTNotificationsMapping_MTokens_id_fk
        FOREIGN KEY (MT_id) REFERENCES MTokens (id),
    CONSTRAINT MTNotificationsMapping_Notifications_id_fk
        FOREIGN KEY (notification_id) REFERENCES Notifications (id)
);

CREATE TABLE IF NOT EXISTS SubscribedNotificationClasses
(
    notificaton_id BIGINT UNSIGNED NOT NULL, class VARCHAR(128) NOT NULL, CONSTRAINT SubscribedNotificationClasses_pk
    UNIQUE (notificaton_id, class), CONSTRAINT SubscribedNotificationClasses_Notifications_id_fk
        FOREIGN KEY (notificaton_id) REFERENCES Notifications (id)
);


### Views

CREATE OR REPLACE VIEW CalendarRemoveCodes AS
SELECT `fa`.`id`                    AS `id`,
       `fa`.`action`                AS `action`,
       `fa`.`code`                  AS `code`,
       `fa`.`expires_at`            AS `expires_at`,
       `arce`.`calendar_mapping_id` AS `calendar_mapping_id`,
       `cm`.`MT_id`                 AS `MT_id`,
       `c`.`id`                     AS `calendar_id`,
       `c`.`ics`                    AS `ics`
    FROM ((((SELECT `ac`.`id`         AS `id`,
                    `ac`.`action`     AS `action`,
                    `ac`.`code`       AS `code`,
                    `ac`.`expires_at` AS `expires_at`
                 FROM `ActionCodes` `ac`
                 WHERE `ac`.`action` = (SELECT `a`.`id`
                                            FROM `Actions` `a`
                                            WHERE `a`.`action` = 'remove_from_calendar')) `fa` JOIN `ActionReferencesCalendarEntries` `arce`
            ON (`arce`.`action_id` = `fa`.`id`)) JOIN `CalendarMapping` `cm`
           ON (`arce`.`calendar_mapping_id` = `cm`.`mapping_id`)) JOIN `Calendars` `c`
          ON (`cm`.`calendar_id` = `c`.`id`));

CREATE OR REPLACE VIEW EventHistory AS
SELECT `me`.`time`       AS `time`,
       `me`.`MT_id`      AS `MT_id`,
       `e`.`event`       AS `event`,
       `me`.`comment`    AS `comment`,
       `me`.`ip`         AS `ip`,
       `me`.`user_agent` AS `user_agent`
    FROM (`Events` `e` JOIN `MT_Events` `me` ON (`e`.`id` = `me`.`event_id`))
    ORDER BY `me`.`time` DESC;

CREATE VIEW IF NOT EXISTS MailVerificationCodes AS
SELECT `fa`.`id`         AS `id`,
       `fa`.`action`     AS `action`,
       `fa`.`code`       AS `code`,
       `fa`.`expires_at` AS `expires_at`,
       `aru`.`uid`       AS `uid`
    FROM ((SELECT `ac`.`id` AS `id`, `ac`.`action` AS `action`, `ac`.`code` AS `code`, `ac`.`expires_at` AS `expires_at`
               FROM `ActionCodes` `ac`
               WHERE `ac`.`action` = (SELECT `a`.`id`
                                          FROM `Actions` `a`
                                          WHERE `a`.`action` = 'verify_email')) `fa` JOIN `ActionReferencesUser` `aru`
          ON (`aru`.`action_id` = `fa`.`id`));

CREATE VIEW IF NOT EXISTS MytokenRecreateCodes AS
SELECT `fa`.`id`           AS `id`,
       `fa`.`action`       AS `action`,
       `fa`.`code`         AS `code`,
       `fa`.`expires_at`   AS `expires_at`,
       `arm`.`MT_id`       AS `MT_id`,
       `mt`.`name`         AS `name`,
       `mt`.`capabilities` AS `capabilities`,
       `mt`.`rotation`     AS `rotation`,
       `mt`.`restrictions` AS `restrictions`,
       `mt`.`created`      AS `token_created`
    FROM (((SELECT `ac`.`id`         AS `id`,
                   `ac`.`action`     AS `action`,
                   `ac`.`code`       AS `code`,
                   `ac`.`expires_at` AS `expires_at`
                FROM `ActionCodes` `ac`
                WHERE `ac`.`action` = (SELECT `a`.`id`
                                           FROM `Actions` `a`
                                           WHERE `a`.`action` = 'recreate_token')) `fa` JOIN `ActionReferencesMytokens` `arm`
           ON (`arm`.`action_id` = `fa`.`id`)) JOIN `MTokens` `mt` ON (`mt`.`id` = `arm`.`MT_id`));

### Procedures

DELIMITER ;;

CREATE OR REPLACE PROCEDURE ActionCodes_AddRecreateToken(IN MTID VARCHAR(128), IN CODE_ VARCHAR(128))
BEGIN
    DECLARE aid BIGINT UNSIGNED;
    DECLARE id BIGINT UNSIGNED;
    SET TIME_ZONE = "+0:00";
    SELECT a.id FROM Actions a WHERE a.`action` = 'recreate_token' INTO aid;
    INSERT INTO ActionCodes (action, code) VALUES (aid, CODE_);
    SELECT LAST_INSERT_ID() INTO id;
    INSERT INTO ActionReferencesMytokens (action_id, MT_id) VALUES (id, MTID);
END;;

CREATE OR REPLACE PROCEDURE ActionCodes_AddRemoveFromCalendar(IN MTID VARCHAR(128), IN CALENDAR VARCHAR(128),
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
                           AND cm.calendar_id = (SELECT c.id
                                                     FROM Calendars c
                                                     WHERE c.name = CALENDAR
                                                       AND c.uid = (SELECT m.user_id
                                                                        FROM MTokens m
                                                                        WHERE m.id = MTID))));
END;;

CREATE OR REPLACE PROCEDURE ActionCodes_AddVerifyMail(IN MTID VARCHAR(128), IN CODE_ VARCHAR(128), IN EXPIRES_IN INT)
BEGIN
    DECLARE aid BIGINT UNSIGNED;
    DECLARE userid BIGINT UNSIGNED;
    DECLARE id BIGINT UNSIGNED;
    SET TIME_ZONE = "+0:00";
    SELECT a.id FROM Actions a WHERE a.`action` = 'verify_email' INTO aid;
    SELECT m.user_id FROM MTokens m WHERE m.id = MTID INTO userid;
    DELETE
        FROM ActionCodes
        WHERE `action` = aid
          AND id IN (SELECT mvc.id FROM MailVerificationCodes mvc WHERE mvc.uid = userid);
    INSERT INTO ActionCodes (action, code, expires_at)
        VALUES (aid, CODE_, (UTC_TIMESTAMP() + INTERVAL EXPIRES_IN SECOND));
    SELECT LAST_INSERT_ID() INTO id;
    INSERT INTO ActionReferencesUser (action_id, uid) VALUES (id, userid);
END;;

CREATE OR REPLACE PROCEDURE ActionCodes_Delete(IN CODE_ VARCHAR(128))
BEGIN
    DELETE FROM ActionCodes WHERE code = CODE_;
END;

CREATE OR REPLACE PROCEDURE ActionCodes_GetRecreateData(IN CODE_ VARCHAR(128))
BEGIN
    SELECT name, capabilities, restrictions, rotation, token_created AS created
        FROM MytokenRecreateCodes
        WHERE code = CODE_;
END;;

CREATE OR REPLACE PROCEDURE ActionCodes_RemoveFromCalendar(IN CODE_ VARCHAR(128))
BEGIN
    SET TIME_ZONE = "+0:00";
    DELETE FROM CalendarMapping WHERE mapping_id = (SELECT mapping_id FROM CalendarRemoveCodes WHERE code = CODE_);
END;;

CREATE OR REPLACE PROCEDURE ActionCodes_UseRecreateToken(IN CODE_ VARCHAR(128))
BEGIN
    CALL ActionCodes_GetRecreateData(CODE_);
    CALL ActionCodes_Delete(CODE_);
END;;

CREATE OR REPLACE PROCEDURE ActionCodes_UseRemoveFromCalendar(IN CODE_ VARCHAR(128))
BEGIN
    CALL ActionCodes_RemoveFromCalendar(CODE_);
    CALL ActionCodes_Delete(CODE_);
END;;

CREATE OR REPLACE PROCEDURE ActionCodes_VerifyMail(IN CODE_ VARCHAR(128))
BEGIN
    SET TIME_ZONE = "+0:00";
    UPDATE Users u
    SET u.email_verified=1
        WHERE u.id =
              (SELECT v.uid FROM MailVerificationCodes v WHERE v.code = CODE_ AND v.expires_at > CURRENT_TIMESTAMP());
END;;

CREATE OR REPLACE PROCEDURE Calendar_AddMytoken(IN MTID VARCHAR(128), IN CALENDARID VARCHAR(128))
BEGIN
    INSERT IGNORE INTO CalendarMapping (calendar_id, MT_id) VALUES (CALENDARID, MTID);
END;;

CREATE OR REPLACE PROCEDURE Calendar_Delete(IN MTID VARCHAR(128), IN NAME_ VARCHAR(128))
BEGIN
    DELETE FROM Calendars WHERE uid = (SELECT m.user_id FROM MTokens m WHERE m.id = MTID) AND name = NAME_;
END;;

CREATE OR REPLACE PROCEDURE Calendar_Get(IN MTID VARCHAR(128), IN NAME_ VARCHAR(128))
BEGIN
    SELECT id, name, ics_path, ics
        FROM Calendars
        WHERE name = NAME_
          AND uid = (SELECT m.user_id FROM MTokens m WHERE m.id = MTID);
END;;

CREATE OR REPLACE PROCEDURE Calendar_GetByID(IN CID VARCHAR(128))
BEGIN
    SELECT id, name, ics_path, ics FROM Calendars WHERE id = CID;
END;;

CREATE OR REPLACE PROCEDURE Calendar_GetMTsInCalendar(IN CALID VARCHAR(128))
BEGIN
    SELECT MT_id FROM CalendarMapping WHERE calendar_id = CALID;
END;;

CREATE OR REPLACE PROCEDURE Calendar_Insert(IN MTID VARCHAR(128), IN CID VARCHAR(128), IN NAME_ VARCHAR(128),
                                            IN PATH_ TEXT,
                                            IN ICS_ LONGTEXT)
BEGIN
    SET TIME_ZONE = "+0:00";
    INSERT INTO Calendars (id, name, uid, ics_path, ics)
        VALUES (CID, NAME_, (SELECT m.user_id FROM MTokens m WHERE m.id = MTID), PATH_, ICS_);
END;;

CREATE OR REPLACE PROCEDURE Calendar_List(IN MTID VARCHAR(128))
BEGIN
    SELECT id, name, ics_path, ics FROM Calendars WHERE uid = (SELECT m.user_id FROM MTokens m WHERE m.id = MTID);
END;;

CREATE OR REPLACE PROCEDURE Calendar_ListForMT(IN MTID VARCHAR(128))
BEGIN
    SELECT id, name, ics_path, ics
        FROM Calendars
        WHERE id IN (SELECT calendar_id FROM CalendarMapping WHERE MT_id = MTID);
END;;

CREATE OR REPLACE PROCEDURE Calendar_Update(IN MTID VARCHAR(128), IN CID VARCHAR(128), IN NAME_ VARCHAR(128),
                                            IN ICS_ LONGTEXT)
BEGIN
    SET TIME_ZONE = "+0:00";
    UPDATE Calendars
    SET name=NAME_, ics=ICS_
        WHERE uid = (SELECT m.user_id FROM MTokens m WHERE m.id = MTID) AND id = CID;
END;;

CREATE OR REPLACE PROCEDURE Calendar_UpdateInternal(IN CID VARCHAR(128), IN NAME_ VARCHAR(128), IN ICS_ LONGTEXT)
BEGIN
    SET TIME_ZONE = "+0:00";
    UPDATE Calendars SET name=NAME_, ics=ICS_ WHERE id = CID;
END;;

CREATE OR REPLACE PROCEDURE Cleanup()
BEGIN
    CALL Cleanup_MTokens();
    CALL Cleanup_AuthInfo();
    CALL Cleanup_ProxyTokens();
    CALL Cleanup_ActionCodes();
END;;

CREATE OR REPLACE PROCEDURE Cleanup_ActionCodes()
BEGIN
    SET TIME_ZONE = "+0:00";
    DELETE FROM ActionCodes WHERE expires_at < CURRENT_TIMESTAMP();
END;;

CREATE OR REPLACE PROCEDURE MTokens_Insert(IN SUB TEXT, IN ISS TEXT, IN MTID VARCHAR(128), IN SEQNO_ BIGINT UNSIGNED,
                                           IN PARENT VARCHAR(128), IN RTID BIGINT UNSIGNED, IN NAME_ TEXT, IN IP TEXT,
                                           IN EXPIRES_AT_ DATETIME, IN CAPABILITIES_ TEXT, IN ROTATION_ TEXT,
                                           IN RESTR TEXT)
BEGIN
    SET TIME_ZONE = "+0:00";
    CALL Users_GetID(SUB, ISS, @UID);
    INSERT INTO MTokens (id, seqno, parent_id, rt_id, name, ip_created, user_id, expires_at, capabilities, rotation,
                         restrictions)
        VALUES (MTID, SEQNO_, PARENT, RTID, NAME_, IP, @UID, EXPIRES_AT_, CAPABILITIES_, ROTATION_, RESTR);
END;;


CREATE OR REPLACE PROCEDURE MTokens_SetMetadata(IN MTID VARCHAR(128), IN CAPABILITIES_ TEXT, IN ROTATION_ TEXT,
                                                IN RESTR TEXT)
BEGIN
    UPDATE MTokens m SET m.capabilities=CAPABILITIES_ WHERE m.id = MTID AND m.capabilities IS NULL;
    UPDATE MTokens m SET m.rotation=ROTATION_ WHERE m.id = MTID AND m.rotation IS NULL;
    UPDATE MTokens m SET m.restrictions=RESTR WHERE m.id = MTID AND m.restrictions IS NULL;
END;;

CREATE OR REPLACE PROCEDURE Mtokens_GetInfo(IN MTID VARCHAR(128))
BEGIN
    SELECT id, parent_id, id AS mom_id, name, created, expires_at, ip_created AS ip
        FROM MTokens
        WHERE id = MTID;
END;;

CREATE OR REPLACE PROCEDURE Notifications_ClearNotificationClasses(IN NID BIGINT UNSIGNED)
BEGIN
    DELETE FROM SubscribedNotificationClasses WHERE notificaton_id = NID;
END;;

CREATE OR REPLACE PROCEDURE Notifications_Create(IN MTID VARCHAR(128), IN TYPE_ VARCHAR(32), IN MNGCODE VARCHAR(128),
                                                 IN WS_ VARCHAR(128), IN USERWIDE TINYINT(1), OUT ID BIGINT UNSIGNED)
BEGIN
    INSERT INTO Notifications (type, management_code, ws, user_wide, uid)
        VALUES (TYPE_, MNGCODE, WS_, USERWIDE, (SELECT m.user_id FROM MTokens m WHERE m.id = MTID));
    SET ID = LAST_INSERT_ID();

END;;

CREATE OR REPLACE PROCEDURE Notifications_CreateForMT(IN MTID VARCHAR(128), IN INCLUDECHILDS TINYINT(1),
                                                      IN TYPE_ VARCHAR(32),
                                                      IN MNGCODE VARCHAR(128), IN WS_ VARCHAR(128))
BEGIN
    CALL Notifications_Create(MTID, TYPE_, MNGCODE, WS_, 0, @ID);
    IF INCLUDECHILDS = 0 THEN
        CALL Notifications_LinkMT(MTID, @ID, INCLUDECHILDS);
    ELSE
        CALL Notifications_LinkMTWithChildren(MTID, @ID);
    END IF;
    SELECT @ID AS notification_id;
END;;

CREATE OR REPLACE PROCEDURE Notifications_CreateUserWide(IN MTID VARCHAR(128), IN TYPE_ VARCHAR(32),
                                                         IN MNGCODE VARCHAR(128),
                                                         IN WS_ VARCHAR(128))
BEGIN
    CALL Notifications_Create(MTID, TYPE_, MNGCODE, WS_, 1, @ID);
    SELECT @ID AS notification_id;
END;;

CREATE OR REPLACE PROCEDURE Notifications_DeleteByManagementCode(IN CODE VARCHAR(128))
BEGIN
    DECLARE nid BIGINT UNSIGNED;
    SELECT id FROM Notifications WHERE management_code = CODE INTO nid;
    DELETE FROM SubscribedNotificationClasses WHERE notificaton_id = nid;
    DELETE FROM MTNotificationsMapping WHERE notification_id = nid;
    DELETE FROM Notifications WHERE id = nid;
END;;

CREATE OR REPLACE PROCEDURE Notifications_ExpandToChildren(IN PARENT VARCHAR(128), IN CHILD VARCHAR(128))
BEGIN
    DECLARE nid BIGINT UNSIGNED;

    SELECT notification_id FROM MTNotificationsMapping WHERE MT_id = PARENT AND include_children = 1 INTO nid;
    IF (nid IS NOT NULL) THEN
        INSERT IGNORE INTO MTNotificationsMapping (MT_id, notification_id, include_children) VALUES (CHILD, nid, 1);
    END IF;
END;;

CREATE OR REPLACE PROCEDURE Notifications_GetForMT(IN MTID VARCHAR(128))
BEGIN
    SELECT n.id, n.type, n.management_code, n.ws, n.user_wide, snc.class
        FROM ((SELECT *
                   FROM Notifications
                   WHERE id IN (
                           (SELECT notification_id FROM MTNotificationsMapping WHERE MT_id = MTID))
                      OR (user_wide = 1 AND uid = (SELECT user_id FROM MTokens m WHERE m.id = MTID))) n JOIN SubscribedNotificationClasses snc
              ON n.id = snc.notificaton_id
                 )
        ORDER BY n.id DESC;
END;;

CREATE OR REPLACE PROCEDURE Notifications_GetForMTAndClass(IN MTID VARCHAR(128), IN _CLASS VARCHAR(128))
BEGIN
    SELECT n.id, n.type, n.management_code, n.ws, n.user_wide
        FROM Notifications n
        WHERE id IN (((SELECT notification_id FROM MTNotificationsMapping WHERE MT_id = MTID)
                      UNION
                      (SELECT id
                           FROM Notifications
                           WHERE user_wide = 1 AND uid = (SELECT user_id FROM MTokens WHERE id = MTID)))
                     INTERSECT
                     (SELECT notificaton_id FROM SubscribedNotificationClasses WHERE class = _CLASS));

END;;

CREATE OR REPLACE PROCEDURE Notifications_GetForManagementCode(IN CODE VARCHAR(128))
BEGIN
    SELECT n.id, n.type, n.management_code, n.ws, n.user_wide, snc.class
        FROM ((SELECT *
                   FROM Notifications
                   WHERE management_code = CODE) n JOIN SubscribedNotificationClasses snc ON n.id = snc.notificaton_id
                 );

END;;

CREATE OR REPLACE PROCEDURE Notifications_GetForUser(IN MTID VARCHAR(128))
BEGIN
    SELECT n.id, n.type, n.management_code, n.ws, n.user_wide, snc.class
        FROM ((SELECT *
                   FROM Notifications
                   WHERE uid = (SELECT user_id FROM MTokens WHERE id = MTID)) n JOIN SubscribedNotificationClasses snc
              ON n.id = snc.notificaton_id
                 )
        ORDER BY n.id DESC;

END;;

CREATE OR REPLACE PROCEDURE Notifications_GetMTsForNotification(IN NID BIGINT UNSIGNED)
BEGIN
    SELECT MT_id FROM MTNotificationsMapping WHERE notification_id = nid;

END;;

CREATE OR REPLACE PROCEDURE Notifications_LinkClass(IN NID BIGINT UNSIGNED, IN NCLASS VARCHAR(128))
BEGIN
    INSERT IGNORE INTO SubscribedNotificationClasses (notificaton_id, class) VALUES (NID, NCLASS);
END;;

CREATE OR REPLACE PROCEDURE Notifications_LinkMT(IN MTID VARCHAR(128), IN NID BIGINT UNSIGNED,
                                                 IN INCLUDECHILDS TINYINT(1))
BEGIN
    INSERT IGNORE INTO MTNotificationsMapping (MT_id, notification_id, include_children)
        VALUES (MTID, NID, INCLUDECHILDS);
END;;

CREATE OR REPLACE PROCEDURE Notifications_LinkMTWithChildren(IN MTID VARCHAR(128), IN NID BIGINT UNSIGNED)
BEGIN
    INSERT IGNORE INTO MTNotificationsMapping(MT_id, notification_id, include_children)
    SELECT c.id, NID, 1
        FROM (WITH RECURSIVE childs AS (SELECT id, parent_id
                                            FROM MTokens
                                            WHERE id = MTID
                                        UNION ALL
                                        SELECT mt.id, mt.parent_id
                                            FROM MTokens mt
                                                     INNER JOIN childs c
                                            WHERE mt.parent_id = c.id)
              SELECT id
                  FROM childs) c;
END;;

CREATE OR REPLACE PROCEDURE Notifications_UnlinkMT(IN MTID VARCHAR(128), IN NID BIGINT UNSIGNED)
BEGIN
    IF (SELECT include_children FROM MTNotificationsMapping WHERE notification_id = NID AND MT_id = MTID) = 0 THEN
        DELETE FROM MTNotificationsMapping WHERE notification_id = NID AND MT_id = MTID;
    ELSE
        DELETE
            FROM MTNotificationsMapping
            WHERE notification_id = NID
              AND MT_id IN (WITH RECURSIVE childs AS (SELECT id, parent_id
                                                          FROM MTokens
                                                          WHERE id = MTID
                                                      UNION ALL
                                                      SELECT mt.id, mt.parent_id
                                                          FROM MTokens mt
                                                                   INNER JOIN childs c
                                                          WHERE mt.parent_id = c.id)
                            SELECT id
                                FROM childs);
    END IF;
END;;

CREATE OR REPLACE PROCEDURE Users_ChangeMail(IN MTID VARCHAR(128), IN MAIL TEXT)
BEGIN
    UPDATE Users SET email=MAIL, email_verified=0 WHERE id = (SELECT m.user_id FROM MTokens m WHERE m.id = MTID);
END;;

CREATE OR REPLACE PROCEDURE Users_ChangePreferredMailType(IN MTID VARCHAR(128), IN PREFER_HTML TINYINT(1))
BEGIN
    UPDATE Users SET prefer_html_mail=PREFER_HTML WHERE id = (SELECT m.user_id FROM MTokens m WHERE m.id = MTID);
END;;

CREATE OR REPLACE PROCEDURE Users_GetMail(IN MTID VARCHAR(128))
BEGIN
    SELECT u.email, u.email_verified, u.prefer_html_mail
        FROM Users u
        WHERE u.id = (SELECT m.user_id FROM MTokens m WHERE m.id = MTID);
END;;

CREATE OR REPLACE PROCEDURE Users_SetMail(IN UID BIGINT UNSIGNED, IN MAIL TEXT, IN VERIFIED BIT)
BEGIN
    UPDATE Users u SET u.email=MAIL, u.email_verified=VERIFIED WHERE u.id = UID;
END;;

CREATE OR REPLACE PROCEDURE Users_SetMailBySub(IN SUB TEXT, IN ISS TEXT, IN MAIL TEXT, IN VERIFIED BIT)
BEGIN
    CALL Users_GetID(SUB, ISS, @UID);
    CALL Users_SetMail(@UID, MAIL, VERIFIED);
END;;

CREATE OR REPLACE PROCEDURE getOIDCIssForManagementCode(IN CODE VARCHAR(128))
BEGIN
    SELECT u.iss FROM Users u WHERE u.id = (SELECT n.uid FROM Notifications n WHERE n.management_code = CODE);
END;;

DELIMITER ;

# Values
INSERT IGNORE INTO Events (event)
    VALUES ('expired');
INSERT IGNORE INTO Events (event)
    VALUES ('revoked');
INSERT IGNORE INTO Events (event)
    VALUES ('notification_subscribed');
INSERT IGNORE INTO Events (event)
    VALUES ('notification_listed');
INSERT IGNORE INTO Events (event)
    VALUES ('notification_unsubscribed');
INSERT IGNORE INTO Events (event)
    VALUES ('notification_subscribed_other');
INSERT IGNORE INTO Events (event)
    VALUES ('notification_unsubscribed_other');
INSERT IGNORE INTO Events (event)
    VALUES ('calendar_created');
INSERT IGNORE INTO Events (event)
    VALUES ('calendar_listed');
INSERT IGNORE INTO Events (event)
    VALUES ('calendar_deleted');
INSERT IGNORE INTO Events (event)
    VALUES ('email_settings_listed');
INSERT IGNORE INTO Events (event)
    VALUES ('email_changed');
INSERT IGNORE INTO Events (event)
    VALUES ('email_mimetype_changed');
INSERT IGNORE INTO Events (event)
    VALUES ('tokeninfo_notifications');
INSERT IGNORE INTO Events (event)
    VALUES ('tokeninfo_notifications_other_token');
INSERT IGNORE INTO Events (event)
    VALUES ('notification_created');
INSERT IGNORE INTO Events (event)
    VALUES ('notification_created_other');

INSERT IGNORE INTO Actions (action)
    VALUES ('verify_email');
INSERT IGNORE INTO Actions (action)
    VALUES ('unsubscribe_notification');
INSERT IGNORE INTO Actions (action)
    VALUES ('recreate_token');
