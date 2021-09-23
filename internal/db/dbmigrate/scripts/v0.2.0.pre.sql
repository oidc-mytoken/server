/*!40101 SET @OLD_CHARACTER_SET_CLIENT = @@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS = @@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION = @@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE = @@TIME_ZONE */;
/*!40103 SET TIME_ZONE = '+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS = @@UNIQUE_CHECKS, UNIQUE_CHECKS = 0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS = @@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS = 0 */;
/*!40101 SET @OLD_SQL_MODE = @@SQL_MODE, SQL_MODE = 'NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES = @@SQL_NOTES, SQL_NOTES = 0 */;

-- Table structure for table `AT_Attributes`;
/*!40101 SET @saved_cs_client = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
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
/*!40101 SET character_set_client = @saved_cs_client */;

-- Table structure for table `AccessTokens`;
/*!40101 SET @saved_cs_client = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
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
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4;
/*!40101 SET character_set_client = @saved_cs_client */;

-- Table structure for table `Attributes`;
/*!40101 SET @saved_cs_client = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE IF NOT EXISTS `Attributes`
(
    `id`        int(10) unsigned NOT NULL AUTO_INCREMENT,
    `attribute` varchar(100)     NOT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `Attributes_UN` (`attribute`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4;
/*!40101 SET character_set_client = @saved_cs_client */;

-- Table structure for table `AuthInfo`;
/*!40101 SET @saved_cs_client = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
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
/*!40101 SET character_set_client = @saved_cs_client */;

-- Table structure for table `EncryptionKeys`;
/*!40101 SET @saved_cs_client = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE IF NOT EXISTS `EncryptionKeys`
(
    `id`             bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `encryption_key` text                NOT NULL,
    `created`        datetime            NOT NULL DEFAULT current_timestamp(),
    PRIMARY KEY (`id`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4;
/*!40101 SET character_set_client = @saved_cs_client */;

-- Temporary table structure for view `EventHistory`;
SET @saved_cs_client = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE TABLE IF NOT EXISTS `EventHistory`
(
    `time`       tinyint NOT NULL,
    `MT_id`      tinyint NOT NULL,
    `event`      tinyint NOT NULL,
    `comment`    tinyint NOT NULL,
    `ip`         tinyint NOT NULL,
    `user_agent` tinyint NOT NULL
) ENGINE = MyISAM */;
SET character_set_client = @saved_cs_client;

-- Table structure for table `Events`;
/*!40101 SET @saved_cs_client = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE IF NOT EXISTS `Events`
(
    `id`    int(10) unsigned NOT NULL AUTO_INCREMENT,
    `event` varchar(100)     NOT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `Events_UN` (`event`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4;
/*!40101 SET character_set_client = @saved_cs_client */;

-- Table structure for table `Grants`;
/*!40101 SET @saved_cs_client = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE IF NOT EXISTS `Grants`
(
    `id`         int(10) unsigned NOT NULL AUTO_INCREMENT,
    `grant_type` varchar(100)     NOT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `Grants_UN` (`grant_type`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4;
/*!40101 SET character_set_client = @saved_cs_client */;

-- Table structure for table `MT_Events`;
/*!40101 SET @saved_cs_client = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
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
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4;
/*!40101 SET character_set_client = @saved_cs_client */;

-- Table structure for table `MTokens`;
/*!40101 SET @saved_cs_client = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
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
/*!40101 SET character_set_client = @saved_cs_client */;
/*!50003 SET @saved_cs_client = @@character_set_client */;
/*!50003 SET @saved_cs_results = @@character_set_results */;
/*!50003 SET @saved_col_connection = @@collation_connection */;
/*!50003 SET character_set_client = utf8mb4 */;
/*!50003 SET character_set_results = utf8mb4 */;
/*!50003 SET collation_connection = utf8mb4_general_ci */;
/*!50003 SET @saved_sql_mode = @@sql_mode */;
/*!50003 SET sql_mode =
        'IGNORE_SPACE,STRICT_TRANS_TABLES,ERROR_FOR_DIVISION_BY_ZERO,NO_AUTO_CREATE_USER,NO_ENGINE_SUBSTITUTION' */;
/*!50003 SET sql_mode = @saved_sql_mode */;
/*!50003 SET character_set_client = @saved_cs_client */;
/*!50003 SET character_set_results = @saved_cs_results */;
/*!50003 SET collation_connection = @saved_col_connection */;

-- Temporary table structure for view `MyTokens`;
SET @saved_cs_client = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE TABLE IF NOT EXISTS `MyTokens`
(
    `id`             tinyint NOT NULL,
    `seqno`          tinyint NOT NULL,
    `parent_id`      tinyint NOT NULL,
    `root_id`        tinyint NOT NULL,
    `name`           tinyint NOT NULL,
    `created`        tinyint NOT NULL,
    `ip_created`     tinyint NOT NULL,
    `user_id`        tinyint NOT NULL,
    `rt_id`          tinyint NOT NULL,
    `refresh_token`  tinyint NOT NULL,
    `rt_updated`     tinyint NOT NULL,
    `encryption_key` tinyint NOT NULL
) ENGINE = MyISAM */;
SET character_set_client = @saved_cs_client;

-- Table structure for table `ProxyTokens`;
/*!40101 SET @saved_cs_client = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
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
/*!40101 SET character_set_client = @saved_cs_client */;

-- Table structure for table `RT_EncryptionKeys`;
/*!40101 SET @saved_cs_client = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
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
/*!40101 SET character_set_client = @saved_cs_client */;

-- Table structure for table `RefreshTokens`;
/*!40101 SET @saved_cs_client = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE IF NOT EXISTS `RefreshTokens`
(
    `id`      bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `rt`      text                NOT NULL,
    `created` datetime            NOT NULL DEFAULT current_timestamp(),
    `updated` datetime            NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
    PRIMARY KEY (`id`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4;
/*!40101 SET character_set_client = @saved_cs_client */;

-- Table structure for table `TokenUsages`;
/*!40101 SET @saved_cs_client = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
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
/*!40101 SET character_set_client = @saved_cs_client */;

-- Temporary table structure for view `TransferCodes`;
SET @saved_cs_client = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE TABLE IF NOT EXISTS `TransferCodes`
(
    `id`               tinyint NOT NULL,
    `jwt`              tinyint NOT NULL,
    `created`          tinyint NOT NULL,
    `expires_in`       tinyint NOT NULL,
    `expires_at`       tinyint NOT NULL,
    `revoke_MT`        tinyint NOT NULL,
    `response_type`    tinyint NOT NULL,
    `consent_declined` tinyint NOT NULL
) ENGINE = MyISAM */;
SET character_set_client = @saved_cs_client;

-- Table structure for table `TransferCodesAttributes`;
/*!40101 SET @saved_cs_client = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
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
/*!40101 SET character_set_client = @saved_cs_client */;

-- Table structure for table `UserGrant_Attributes`;
/*!40101 SET @saved_cs_client = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
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
/*!40101 SET character_set_client = @saved_cs_client */;

-- Table structure for table `UserGrants`;
/*!40101 SET @saved_cs_client = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
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
/*!40101 SET character_set_client = @saved_cs_client */;

-- Table structure for table `Users`;
/*!40101 SET @saved_cs_client = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
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
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4;
/*!40101 SET character_set_client = @saved_cs_client */;

-- Table structure for table `version`;
CREATE TABLE IF NOT EXISTS `version`
(
    `version` varchar(64) NOT NULL,
    `bef`     datetime DEFAULT NULL,
    `aft`     datetime DEFAULT NULL,
    PRIMARY KEY (`version`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4;

-- Final view structure for view `EventHistory`;
/*!50001 DROP TABLE IF EXISTS `EventHistory`*/;;
/*!50001 DROP VIEW IF EXISTS `EventHistory`*/;;
/*!50001 SET @saved_cs_client = @@character_set_client */;;
/*!50001 SET @saved_cs_results = @@character_set_results */;;
/*!50001 SET @saved_col_connection = @@collation_connection */;;
/*!50001 SET character_set_client = utf8mb4 */;;
/*!50001 SET character_set_results = utf8mb4 */;;
/*!50001 SET collation_connection = utf8mb4_general_ci */;;
/*!50001 CREATE ALGORITHM = UNDEFINED */ /*!50001 VIEW `EventHistory` AS
select `me`.`time`       AS `time`,
       `me`.`MT_id`      AS `MT_id`,
       `e`.`event`       AS `event`,
       `me`.`comment`    AS `comment`,
       `me`.`ip`         AS `ip`,
       `me`.`user_agent` AS `user_agent`
from (`Events` `e`
         join `MT_Events` `me` on (`e`.`id` = `me`.`event_id`))
order by `me`.`time`
*/;;
/*!50001 SET character_set_client = @saved_cs_client */;;
/*!50001 SET character_set_results = @saved_cs_results */;;
/*!50001 SET collation_connection = @saved_col_connection */;

-- Final view structure for view `MyTokens`;
/*!50001 DROP TABLE IF EXISTS `MyTokens`*/;;
/*!50001 DROP VIEW IF EXISTS `MyTokens`*/;;
/*!50001 SET @saved_cs_client = @@character_set_client */;;
/*!50001 SET @saved_cs_results = @@character_set_results */;;
/*!50001 SET @saved_col_connection = @@collation_connection */;;
/*!50001 SET character_set_client = utf8mb4 */;;
/*!50001 SET character_set_results = utf8mb4 */;;
/*!50001 SET collation_connection = utf8mb4_general_ci */;;
/*!50001 CREATE ALGORITHM = UNDEFINED */ /*!50001 VIEW `MyTokens` AS
select `mt`.`id`               AS `id`,
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
from (((`MTokens` `mt` join `RefreshTokens` `rts` on (`mt`.`rt_id` = `rts`.`id`)) join `RT_EncryptionKeys` `rkeys` on (`mt`.`id` = `rkeys`.`MT_id` and `mt`.`rt_id` = `rkeys`.`rt_id`))
         join `EncryptionKeys` `keys` on (`rkeys`.`key_id` = `keys`.`id`))
*/;;
/*!50001 SET character_set_client = @saved_cs_client */;;
/*!50001 SET character_set_results = @saved_cs_results */;;
/*!50001 SET collation_connection = @saved_col_connection */;

-- Final view structure for view `TransferCodes`;
/*!50001 DROP TABLE IF EXISTS `TransferCodes`*/;;
/*!50001 DROP VIEW IF EXISTS `TransferCodes`*/;;
/*!50001 SET @saved_cs_client = @@character_set_client */;;
/*!50001 SET @saved_cs_results = @@character_set_results */;;
/*!50001 SET @saved_col_connection = @@collation_connection */;;
/*!50001 SET character_set_client = utf8mb4 */;;
/*!50001 SET character_set_results = utf8mb4 */;;
/*!50001 SET collation_connection = utf8mb4_general_ci */;;
/*!50001 CREATE ALGORITHM = UNDEFINED */ /*!50001 VIEW `TransferCodes` AS
select `pt`.`id`                AS `id`,
       `pt`.`jwt`               AS `jwt`,
       `tca`.`created`          AS `created`,
       `tca`.`expires_in`       AS `expires_in`,
       `tca`.`expires_at`       AS `expires_at`,
       `tca`.`revoke_MT`        AS `revoke_MT`,
       `tca`.`response_type`    AS `response_type`,
       `tca`.`consent_declined` AS `consent_declined`
from (`ProxyTokens` `pt`
         join `TransferCodesAttributes` `tca` on (`pt`.`id` = `tca`.`id`))
*/;;
/*!50001 SET character_set_client = @saved_cs_client */;;
/*!50001 SET character_set_results = @saved_cs_results */;;
/*!50001 SET collation_connection = @saved_col_connection */;;
/*!40103 SET TIME_ZONE = @OLD_TIME_ZONE */;;
/*!40101 SET SQL_MODE = @OLD_SQL_MODE */;;
/*!40014 SET FOREIGN_KEY_CHECKS = @OLD_FOREIGN_KEY_CHECKS */;;
/*!40014 SET UNIQUE_CHECKS = @OLD_UNIQUE_CHECKS */;;
/*!40101 SET CHARACTER_SET_CLIENT = @OLD_CHARACTER_SET_CLIENT */;;
/*!40101 SET CHARACTER_SET_RESULTS = @OLD_CHARACTER_SET_RESULTS */;;
/*!40101 SET COLLATION_CONNECTION = @OLD_COLLATION_CONNECTION */;;
/*!40111 SET SQL_NOTES = @OLD_SQL_NOTES */;

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
