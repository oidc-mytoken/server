/*!40014 SET @OLD_UNIQUE_CHECKS = @@UNIQUE_CHECKS, UNIQUE_CHECKS = 0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS = @@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS = 0 */;

CREATE TABLE IF NOT EXISTS `AT_Attributes`
(
    `AT_id`        bigint(20) unsigned NOT NULL,
    `attribute_id` int(10) unsigned    NOT NULL,
    `attribute`    text                NOT NULL,
    KEY `AT_Attributes_FK_1` (`attribute_id`),
    KEY `AT_Attributes_FK` (`AT_id`),
    CONSTRAINT `AT_Attributes_FK` FOREIGN KEY (`AT_id`) REFERENCES `AccessTokens` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT `AT_Attributes_FK_1` FOREIGN KEY (`attribute_id`) REFERENCES `Attributes` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4;

CREATE TABLE IF NOT EXISTS `AccessTokens`
(
    `id`         bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `token`      text                NOT NULL,
    `created`    datetime            NOT NULL DEFAULT current_timestamp(),
    `ip_created` varchar(32)         NOT NULL,
    `comment`    text                         DEFAULT NULL,
    `MT_id`      varchar(128)        NOT NULL,
    PRIMARY KEY (`id`),
    KEY `AccessTokens_FK` (`MT_id`),
    CONSTRAINT `AccessTokens_FK` FOREIGN KEY (`MT_id`) REFERENCES `MTokens` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4;

CREATE TABLE IF NOT EXISTS `Attributes`
(
    `id`        int(10) unsigned NOT NULL AUTO_INCREMENT,
    `attribute` varchar(100)     NOT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `Attributes_UN` (`attribute`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4;

CREATE TABLE IF NOT EXISTS `AuthInfo`
(
    `state_h`               varchar(128) NOT NULL,
    `iss`                   text         NOT NULL,
    `restrictions`          longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT NULL CHECK (json_valid(`restrictions`)),
    `capabilities`          longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT NULL CHECK (json_valid(`capabilities`)),
    `name`                  varchar(100)                                       DEFAULT NULL,
    `polling_code`          bit(1)       NOT NULL                              DEFAULT b'0',
    `created`               datetime     NOT NULL                              DEFAULT current_timestamp(),
    `expires_in`            int(11)      NOT NULL,
    `expires_at`            datetime     NOT NULL                              DEFAULT (current_timestamp() + interval `expires_in` second),
    `subtoken_capabilities` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT NULL CHECK (json_valid(`subtoken_capabilities`)),
    PRIMARY KEY (`state_h`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4;

CREATE TABLE IF NOT EXISTS `EncryptionKeys`
(
    `id`             bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `encryption_key` text                NOT NULL,
    `created`        datetime            NOT NULL DEFAULT current_timestamp(),
    PRIMARY KEY (`id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4;

CREATE TABLE IF NOT EXISTS `Events`
(
    `id`    int(10) unsigned NOT NULL AUTO_INCREMENT,
    `event` varchar(100)     NOT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `Events_UN` (`event`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4;

CREATE TABLE IF NOT EXISTS `Grants`
(
    `id`         int(10) unsigned NOT NULL AUTO_INCREMENT,
    `grant_type` varchar(100)     NOT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `Grants_UN` (`grant_type`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4;

CREATE TABLE IF NOT EXISTS `MT_Events`
(
    `id`         bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `MT_id`      varchar(128)        NOT NULL,
    `time`       datetime            NOT NULL DEFAULT current_timestamp(),
    `event_id`   int(10) unsigned    NOT NULL DEFAULT 0,
    `comment`    varchar(100)                 DEFAULT NULL,
    `ip`         varchar(32)         NOT NULL,
    `user_agent` text                NOT NULL,
    PRIMARY KEY (`id`),
    KEY `MT_Events_FK_2` (`MT_id`),
    KEY `MT_Events_FK_3` (`event_id`),
    CONSTRAINT `MT_Events_FK_2` FOREIGN KEY (`MT_id`) REFERENCES `MTokens` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT `MT_Events_FK_3` FOREIGN KEY (`event_id`) REFERENCES `Events` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4;

CREATE TABLE IF NOT EXISTS `MTokens`
(
    `id`           varchar(128)        NOT NULL,
    `parent_id`    varchar(128)                 DEFAULT NULL,
    `root_id`      varchar(128)                 DEFAULT NULL,
    `name`         varchar(100)                 DEFAULT NULL,
    `created`      datetime            NOT NULL DEFAULT current_timestamp(),
    `ip_created`   varchar(32)         NOT NULL,
    `user_id`      bigint(20) unsigned NOT NULL,
    `rt_id`        bigint(20) unsigned NOT NULL,
    `seqno`        bigint(20) unsigned NOT NULL,
    `last_rotated` datetime            NOT NULL DEFAULT current_timestamp(),
    PRIMARY KEY (`id`),
    KEY `SessionTokens_parent_id_IDX` (`parent_id`) USING BTREE,
    KEY `SessionTokens_root_id_IDX` (`root_id`) USING BTREE,
    KEY `Mytokens_FK_2` (`user_id`),
    KEY `Mytokens_FK_3` (`rt_id`),
    CONSTRAINT `Mytokens_FK` FOREIGN KEY (`parent_id`) REFERENCES `MTokens` (`id`) ON DELETE SET NULL ON UPDATE CASCADE,
    CONSTRAINT `Mytokens_FK_1` FOREIGN KEY (`root_id`) REFERENCES `MTokens` (`id`) ON DELETE SET NULL ON UPDATE CASCADE,
    CONSTRAINT `Mytokens_FK_2` FOREIGN KEY (`user_id`) REFERENCES `Users` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT `Mytokens_FK_3` FOREIGN KEY (`rt_id`) REFERENCES `RefreshTokens` (`id`) ON UPDATE CASCADE
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4;

CREATE TABLE IF NOT EXISTS `ProxyTokens`
(
    `id`    varchar(128) NOT NULL,
    `jwt`   text         NOT NULL,
    `MT_id` varchar(128) DEFAULT NULL,
    PRIMARY KEY (`id`),
    KEY `ProxyTokens_FK` (`MT_id`),
    CONSTRAINT `ProxyTokens_FK` FOREIGN KEY (`MT_id`) REFERENCES `MTokens` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4;

CREATE TABLE IF NOT EXISTS `RT_EncryptionKeys`
(
    `rt_id`  bigint(20) unsigned NOT NULL,
    `MT_id`  varchar(128)        NOT NULL,
    `key_id` bigint(20) unsigned NOT NULL,
    PRIMARY KEY (`rt_id`, `MT_id`),
    KEY `RT_EncryptionKeys_FK` (`key_id`),
    CONSTRAINT `RT_EncryptionKeys_FK` FOREIGN KEY (`key_id`) REFERENCES `EncryptionKeys` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT `RT_EncryptionKeys_FK_1` FOREIGN KEY (`rt_id`) REFERENCES `RefreshTokens` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT `RT_EncryptionKeys_FK_2` FOREIGN KEY (`MT_id`) REFERENCES `MTokens` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4;

CREATE TABLE IF NOT EXISTS `RefreshTokens`
(
    `id`      bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `rt`      text                NOT NULL,
    `created` datetime            NOT NULL DEFAULT current_timestamp(),
    `updated` datetime            NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
    PRIMARY KEY (`id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4;

CREATE TABLE IF NOT EXISTS `TokenUsages`
(
    `MT_id`            varchar(128)                                       NOT NULL,
    `restriction`      longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL CHECK (json_valid(`restriction`)),
    `usages_AT`        int(10) unsigned                                   NOT NULL DEFAULT 0,
    `usages_other`     int(10) unsigned                                   NOT NULL DEFAULT 0,
    `restriction_hash` char(128)                                          NOT NULL,
    PRIMARY KEY (`MT_id`, `restriction_hash`),
    CONSTRAINT `TokenUsages_FK` FOREIGN KEY (`MT_id`) REFERENCES `MTokens` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4;

CREATE TABLE IF NOT EXISTS `TransferCodesAttributes`
(
    `id`               varchar(128) NOT NULL,
    `created`          datetime     NOT NULL DEFAULT current_timestamp(),
    `expires_in`       int(11)      NOT NULL,
    `expires_at`       datetime     NOT NULL DEFAULT (current_timestamp() + interval `expires_in` second),
    `revoke_MT`        bit(1)       NOT NULL DEFAULT b'0',
    `response_type`    varchar(128) NOT NULL DEFAULT 'token',
    `consent_declined` bit(1)                DEFAULT NULL,
    PRIMARY KEY (`id`),
    CONSTRAINT `TransferCodesAttributes_FK` FOREIGN KEY (`id`) REFERENCES `ProxyTokens` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4;

CREATE TABLE IF NOT EXISTS `UserGrant_Attributes`
(
    `user_id`      bigint(20) unsigned NOT NULL,
    `grant_id`     int(10) unsigned    NOT NULL,
    `attribute_id` int(10) unsigned    NOT NULL,
    `attribute`    text                NOT NULL,
    PRIMARY KEY (`user_id`, `grant_id`),
    KEY `UserGrant_Attributes_FK_3` (`attribute_id`),
    KEY `UserGrant_Attributes_FK_1` (`grant_id`),
    CONSTRAINT `UserGrant_Attributes_FK` FOREIGN KEY (`user_id`) REFERENCES `Users` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT `UserGrant_Attributes_FK_1` FOREIGN KEY (`grant_id`) REFERENCES `Grants` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT `UserGrant_Attributes_FK_3` FOREIGN KEY (`attribute_id`) REFERENCES `Attributes` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4;

CREATE TABLE IF NOT EXISTS `UserGrants`
(
    `user_id`  bigint(20) unsigned NOT NULL,
    `grant_id` int(10) unsigned    NOT NULL,
    `enabled`  bit(1)              NOT NULL,
    PRIMARY KEY (`user_id`, `grant_id`),
    KEY `UserGrants_FK` (`grant_id`),
    CONSTRAINT `UserGrants_FK` FOREIGN KEY (`grant_id`) REFERENCES `Grants` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT `UserGrants_FK_1` FOREIGN KEY (`user_id`) REFERENCES `Users` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4;

CREATE TABLE IF NOT EXISTS `Users`
(
    `id`            bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `sub`           varchar(512)        NOT NULL,
    `iss`           varchar(256)        NOT NULL,
    `token_tracing` tinyint(1)          NOT NULL DEFAULT 1,
    `jwt_pk`        text                         DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `Users_UN` (`sub`, `iss`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4;

CREATE TABLE IF NOT EXISTS `version`
(
    `version` varchar(64) NOT NULL,
    `bef`     datetime DEFAULT NULL,
    `aft`     datetime DEFAULT NULL,
    PRIMARY KEY (`version`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4;

CREATE OR REPLACE ALGORITHM = UNDEFINED VIEW `EventHistory` AS
SELECT `me`.`time`       AS `time`,
       `me`.`MT_id`      AS `MT_id`,
       `e`.`event`       AS `event`,
       `me`.`comment`    AS `comment`,
       `me`.`ip`         AS `ip`,
       `me`.`user_agent` AS `user_agent`
    FROM (`Events` `e`
             JOIN `MT_Events` `me` ON (`e`.`id` = `me`.`event_id`))
    ORDER BY `me`.`time`;

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
       `rts`.`rt`              AS `refresh_token`,
       `rts`.`updated`         AS `rt_updated`,
       `keys`.`encryption_key` AS `encryption_key`
    FROM (((`MTokens` `mt` JOIN `RefreshTokens` `rts` ON (`mt`.`rt_id` = `rts`.`id`)) JOIN `RT_EncryptionKeys` `rkeys` ON (`mt`.`id` = `rkeys`.`MT_id` AND `mt`.`rt_id` = `rkeys`.`rt_id`))
             JOIN `EncryptionKeys` `keys` ON (`rkeys`.`key_id` = `keys`.`id`));

CREATE OR REPLACE ALGORITHM = UNDEFINED VIEW `TransferCodes` AS
SELECT `pt`.`id`                AS `id`,
       `pt`.`jwt`               AS `jwt`,
       `tca`.`created`          AS `created`,
       `tca`.`expires_in`       AS `expires_in`,
       `tca`.`expires_at`       AS `expires_at`,
       `tca`.`revoke_MT`        AS `revoke_MT`,
       `tca`.`response_type`    AS `response_type`,
       `tca`.`consent_declined` AS `consent_declined`
    FROM (`ProxyTokens` `pt`
             JOIN `TransferCodesAttributes` `tca` ON (`pt`.`id` = `tca`.`id`));

# Predefined Values
INSERT IGNORE INTO Attributes(attribute)
    VALUES ('scope');
INSERT IGNORE INTO Attributes(attribute)
    VALUES ('audience');
INSERT IGNORE INTO Attributes(attribute)
    VALUES ('capability');

INSERT IGNORE INTO Events (event)
VALUES ('unknown');
INSERT IGNORE INTO Events (event)
VALUES ('created');
INSERT IGNORE INTO Events (event)
VALUES ('AT_created');
INSERT IGNORE INTO Events (event)
VALUES ('MT_created');
INSERT IGNORE INTO Events (event)
VALUES ('tokeninfo_introspect');
INSERT IGNORE INTO Events (event)
    VALUES ('tokeninfo_history');
INSERT IGNORE INTO Events (event)
    VALUES ('tokeninfo_tree');
INSERT IGNORE INTO Events (event)
    VALUES ('tokeninfo_list_mytokens');
INSERT IGNORE INTO Events (event)
    VALUES ('mng_enabled_AT_grant');
INSERT IGNORE INTO Events (event)
    VALUES ('mng_disabled_AT_grant');
INSERT IGNORE INTO Events (event)
    VALUES ('mng_enabled_JWT_grant');
INSERT IGNORE INTO Events (event)
    VALUES ('mng_disabled_JWT_grant');
INSERT IGNORE INTO Events (event)
    VALUES ('mng_linked_grant');
INSERT IGNORE INTO Events (event)
    VALUES ('mng_unlinked_grant');
INSERT IGNORE INTO Events (event)
    VALUES ('mng_enabled_tracing');
INSERT IGNORE INTO Events (event)
    VALUES ('mng_disabled_tracing');
INSERT IGNORE INTO Events (event)
    VALUES ('inherited_RT');
INSERT IGNORE INTO Events (event)
    VALUES ('transfer_code_created');
INSERT IGNORE INTO Events (event)
    VALUES ('transfer_code_used');

INSERT IGNORE INTO Grants (grant_type)
    VALUES ('mytoken');
INSERT IGNORE INTO Grants (grant_type)
    VALUES ('oidc_flow');
INSERT IGNORE INTO Grants (grant_type)
    VALUES ('polling_code');
INSERT IGNORE INTO Grants (grant_type)
    VALUES ('transfer_code');


DELIMITER ;;
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
