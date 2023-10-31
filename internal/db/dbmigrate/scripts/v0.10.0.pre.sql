# Tables
ALTER TABLE Users
    ADD email TEXT NULL;
ALTER TABLE Users
    ADD email_verified BOOL DEFAULT 0 NOT NULL;

ALTER TABLE MTokens
    ADD capabilities JSON NULL;
ALTER TABLE MTokens
    ADD rotation JSON NULL;
ALTER TABLE MTokens
    ADD restrictions JSON NULL;

CREATE OR REPLACE TABLE Calendars
(
    id                    VARCHAR(128)    NOT NULL
        PRIMARY KEY,
    name                  VARCHAR(128)    NULL,
    uid                   BIGINT UNSIGNED NOT NULL,
    ics_path              VARCHAR(128)    NOT NULL, ics LONGTEXT NOT NULL,
    CONSTRAINT Calendars_UN
        UNIQUE (ics_path),
    CONSTRAINT Calendars_UN_1
        UNIQUE (name, uid),
    CONSTRAINT Calendars_FK
        FOREIGN KEY (uid) REFERENCES Users (id)
            ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE OR REPLACE TABLE NotificationSchedulerSync
(
    sync_i BIGINT NOT NULL
        PRIMARY KEY
);

CREATE OR REPLACE TABLE Notifications
(
    id                    VARCHAR(128) NOT NULL
        PRIMARY KEY, type VARCHAR(32)  NOT NULL, created DATETIME DEFAULT CURRENT_TIMESTAMP() NOT NULL
);

CREATE OR REPLACE TABLE CalendarMapping
(
    calendar_id VARCHAR(128) NOT NULL, MT_id VARCHAR(128) NOT NULL, CONSTRAINT CalendarMapping_FK
    FOREIGN KEY (MT_id) REFERENCES MTokens (id)
        ON UPDATE CASCADE ON DELETE CASCADE, CONSTRAINT CalendarMapping_FK_1
        FOREIGN KEY (calendar_id) REFERENCES Calendars (id)
            ON UPDATE CASCADE ON DELETE CASCADE
);


CREATE OR REPLACE TABLE NotificationEventSubscriptions
(
    notification_id VARCHAR(128) NOT NULL, event INT UNSIGNED NOT NULL, CONSTRAINT NotificationEventSubscriptions_FK
    FOREIGN KEY (event) REFERENCES Events (id)
        ON UPDATE CASCADE ON DELETE CASCADE, CONSTRAINT NotificationEventSubscriptions_FK_1
        FOREIGN KEY (notification_id) REFERENCES Notifications (id)
            ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE OR REPLACE TABLE NotificationSubscriptions
(
    notification_id VARCHAR(128)                         NOT NULL,
    MT_id           VARCHAR(128)                         NOT NULL,
    added           DATETIME DEFAULT CURRENT_TIMESTAMP() NOT NULL,
    subscription_id VARCHAR(128)                         NOT NULL
        PRIMARY KEY,
    CONSTRAINT NotificationSubscriptions_UN
        UNIQUE (notification_id, MT_id),
    CONSTRAINT NotificationSubscriptions_FK
        FOREIGN KEY (notification_id) REFERENCES Notifications (id)
            ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT NotificationSubscriptions_FK_1
        FOREIGN KEY (MT_id) REFERENCES MTokens (id)
            ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE OR REPLACE TABLE NotificationSchedule
(
    due_time        DATETIME     NOT NULL,
    subscription_id VARCHAR(128) NOT NULL,
    type            VARCHAR(128) NOT NULL,
    CONSTRAINT NotificationSchedule_FK
        FOREIGN KEY (subscription_id) REFERENCES NotificationSubscriptions (subscription_id)
            ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE INDEX Notifications_FK ON Notifications (type);

CREATE OR REPLACE TABLE Actions
(
    id                      INT UNSIGNED AUTO_INCREMENT
        PRIMARY KEY, action VARCHAR(128) NOT NULL, CONSTRAINT Actions_UN
        UNIQUE (action)
);

CREATE OR REPLACE TABLE ActionCodes
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

CREATE OR REPLACE TABLE ActionReferencesMytokens
(
    action_id BIGINT UNSIGNED NOT NULL, MT_id VARCHAR(128) NOT NULL, CONSTRAINT ActionReferencesMytokens_FK
    FOREIGN KEY (action_id) REFERENCES ActionCodes (id)
        ON UPDATE CASCADE ON DELETE CASCADE, CONSTRAINT ActionReferencesMytokens_FK_1
        FOREIGN KEY (MT_id) REFERENCES MTokens (id)
            ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE OR REPLACE TABLE ActionReferencesNotification
(
    action_id       BIGINT UNSIGNED NOT NULL,
    notification_id VARCHAR(128)    NOT NULL,
    CONSTRAINT ActionReferencesNotification_FK
        FOREIGN KEY (action_id) REFERENCES ActionCodes (id)
            ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT ActionReferencesNotification_FK_1
        FOREIGN KEY (notification_id) REFERENCES Notifications (id)
            ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE OR REPLACE TABLE ActionReferencesUser
(
    action_id BIGINT UNSIGNED NOT NULL, uid BIGINT UNSIGNED NOT NULL, CONSTRAINT ActionReferencesUser_FK
    FOREIGN KEY (uid) REFERENCES Users (id)
        ON UPDATE CASCADE ON DELETE CASCADE, CONSTRAINT ActionReferencesUser_FK_1
        FOREIGN KEY (action_id) REFERENCES ActionCodes (id)
            ON UPDATE CASCADE ON DELETE CASCADE
);

# Views

CREATE OR REPLACE VIEW MailVerificationCodes AS
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

CREATE OR REPLACE VIEW MytokenRecreateAndUnsubscribeCodes AS
SELECT `fa`.`id`               AS `id`,
       `fa`.`action`           AS `action`,
       `fa`.`code`             AS `code`,
       `fa`.`expires_at`       AS `expires_at`,
       `arn`.`notification_id` AS `notification_id`,
       `arm`.`MT_id`           AS `MT_id`
    FROM (((SELECT `ac`.`id`         AS `id`,
                   `ac`.`action`     AS `action`,
                   `ac`.`code`       AS `code`,
                   `ac`.`expires_at` AS `expires_at`
                FROM `ActionCodes` `ac`
                WHERE `ac`.`action` = (SELECT `a`.`id`
                                           FROM `Actions` `a`
                                           WHERE `a`.`action` = 'token_recreate_unsubscribe')) `fa` JOIN `ActionReferencesNotification` `arn`
           ON (`arn`.`action_id` = `fa`.`id`)) JOIN `ActionReferencesMytokens` `arm`
          ON (`arm`.`action_id` = `fa`.`id`));

CREATE OR REPLACE VIEW NotificationUnsubscribeCodes AS
SELECT `fa`.`id`               AS `id`,
       `fa`.`action`           AS `action`,
       `fa`.`code`             AS `code`,
       `fa`.`expires_at`       AS `expires_at`,
       `arn`.`notification_id` AS `notification_id`
    FROM ((SELECT `ac`.`id` AS `id`, `ac`.`action` AS `action`, `ac`.`code` AS `code`, `ac`.`expires_at` AS `expires_at`
               FROM `ActionCodes` `ac`
               WHERE `ac`.`action` = (SELECT `a`.`id`
                                          FROM `Actions` `a`
                                          WHERE `a`.`action` = 'unsubscribe_notification')) `fa` JOIN `ActionReferencesNotification` `arn`
          ON (`arn`.`action_id` = `fa`.`id`));


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

INSERT IGNORE INTO Actions (action)
    VALUES ('verify_email');
INSERT IGNORE INTO Actions (action)
    VALUES ('unsubscribe_notification');
INSERT IGNORE INTO Actions (action)
    VALUES ('recreate_token');


# Procedures
DELIMITER ;;

CREATE OR REPLACE PROCEDURE MTokens_Insert(IN SUB TEXT, IN ISS TEXT, IN MTID VARCHAR(128),
                                           IN SEQNO_ BIGINT UNSIGNED, IN PARENT VARCHAR(128),
                                           IN RTID BIGINT UNSIGNED, IN NAME_ TEXT, IN IP TEXT,
                                           IN EXPIRES_AT_ DATETIME, IN CAPABILITIES_ TEXT,
                                           IN ROTATION_ TEXT, IN RESTR TEXT)
BEGIN
    SET TIME_ZONE = "+0:00";
    CALL Users_GetID(SUB, ISS, @UID);
    INSERT INTO MTokens (id, seqno, parent_id, rt_id, name, ip_created, user_id, expires_at, capabilities, rotation,
                         restrictions)
        VALUES (MTID, SEQNO_, PARENT, RTID, NAME_, IP, @UID, EXPIRES_AT_, CAPABILITIES_, ROTATION_, RESTR);
END;;

CREATE OR REPLACE PROCEDURE MTokens_SetMetadata(IN MTID VARCHAR(128), IN CAPABILITIES_ TEXT,
                                                IN ROTATION_ TEXT, IN RESTR TEXT)
BEGIN
    UPDATE MTokens m SET m.capabilities=CAPABILITIES_ WHERE m.id = MTID AND m.capabilities IS NULL;
    UPDATE MTokens m SET m.rotation=ROTATION_ WHERE m.id = MTID AND m.rotation IS NULL;
    UPDATE MTokens m SET m.restrictions=RESTR WHERE m.id = MTID AND m.restrictions IS NULL;
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

CREATE OR REPLACE PROCEDURE Cleanup_ActionCodes()
BEGIN
    SET TIME_ZONE = "+0:00";
    DELETE FROM ActionCodes WHERE expires_at < CURRENT_TIMESTAMP();
END;;

CREATE OR REPLACE PROCEDURE Cleanup()
BEGIN
    CALL Cleanup_MTokens();
    CALL Cleanup_AuthInfo();
    CALL Cleanup_ProxyTokens();
    CALL Cleanup_ActionCodes();
END;;

CREATE OR REPLACE PROCEDURE Mtokens_GetInfo(IN MTID VARCHAR(128))
BEGIN
    SELECT id, parent_id, id AS mom_id, name, created, expires_at, ip_created AS ip
        FROM MTokens
        WHERE id = MTID;
END;


DELIMITER ;
