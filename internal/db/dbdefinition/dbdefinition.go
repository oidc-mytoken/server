package dbdefinition

// DDL holds all commands to create the necessary database tables
var DDL = []string {
  ""+
  "/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;",
  "/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;",
  "/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;",
  "/*!40101 SET NAMES utf8mb4 */;",
  "/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;",
  "/*!40103 SET TIME_ZONE='+00:00' */;",
  "/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;",
  "/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;",
  "/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;",
  "/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;",
  ""+
  "--",
  "-- Table structure for table `AT_Attributes`",
  "--",
  ""+
  "DROP TABLE IF EXISTS `AT_Attributes`;",
  "/*!40101 SET @saved_cs_client     = @@character_set_client */;",
  "/*!40101 SET character_set_client = utf8 */;",
  "CREATE TABLE `AT_Attributes` ("+
  "  `AT_id` bigint(20) unsigned NOT NULL,"+
  "  `attribute_id` int(10) unsigned NOT NULL,"+
  "  `attribute` text NOT NULL,"+
  "  KEY `AT_Attributes_FK_1` (`attribute_id`),"+
  "  KEY `AT_Attributes_FK` (`AT_id`),"+
  "  CONSTRAINT `AT_Attributes_FK` FOREIGN KEY (`AT_id`) REFERENCES `AccessTokens` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,"+
  "  CONSTRAINT `AT_Attributes_FK_1` FOREIGN KEY (`attribute_id`) REFERENCES `Attributes` (`id`) ON DELETE CASCADE ON UPDATE CASCADE"+
  ") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;",
  "/*!40101 SET character_set_client = @saved_cs_client */;",
  ""+
  "--",
  "-- Table structure for table `AccessTokens`",
  "--",
  ""+
  "DROP TABLE IF EXISTS `AccessTokens`;",
  "/*!40101 SET @saved_cs_client     = @@character_set_client */;",
  "/*!40101 SET character_set_client = utf8 */;",
  "CREATE TABLE `AccessTokens` ("+
  "  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,"+
  "  `token` text NOT NULL,"+
  "  `created` datetime NOT NULL DEFAULT current_timestamp(),"+
  "  `ip_created` varchar(32) NOT NULL,"+
  "  `comment` text DEFAULT NULL,"+
  "  `ST_id` varchar(64) NOT NULL,"+
  "  PRIMARY KEY (`id`),"+
  "  KEY `AccessTokens_FK` (`ST_id`),"+
  "  CONSTRAINT `AccessTokens_FK` FOREIGN KEY (`ST_id`) REFERENCES `SuperTokens` (`id`) ON DELETE CASCADE ON UPDATE CASCADE"+
  ") ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;",
  "/*!40101 SET character_set_client = @saved_cs_client */;",
  ""+
  "--",
  "-- Table structure for table `Attributes`",
  "--",
  ""+
  "DROP TABLE IF EXISTS `Attributes`;",
  "/*!40101 SET @saved_cs_client     = @@character_set_client */;",
  "/*!40101 SET character_set_client = utf8 */;",
  "CREATE TABLE `Attributes` ("+
  "  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,"+
  "  `attribute` varchar(100) NOT NULL,"+
  "  PRIMARY KEY (`id`),"+
  "  UNIQUE KEY `Attributes_UN` (`attribute`)"+
  ") ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;",
  "/*!40101 SET character_set_client = @saved_cs_client */;",
  ""+
  "--",
  "-- Table structure for table `AuthInfo`",
  "--",
  ""+
  "DROP TABLE IF EXISTS `AuthInfo`;",
  "/*!40101 SET @saved_cs_client     = @@character_set_client */;",
  "/*!40101 SET character_set_client = utf8 */;",
  "CREATE TABLE `AuthInfo` ("+
  "  `state_h` varchar(128) NOT NULL,"+
  "  `iss` text NOT NULL,"+
  "  `restrictions` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT NULL CHECK (json_valid(`restrictions`)),"+
  "  `capabilities` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT NULL CHECK (json_valid(`capabilities`)),"+
  "  `name` varchar(100) DEFAULT NULL,"+
  "  `polling_code` bit(1) NOT NULL DEFAULT b'0',"+
  "  `created` datetime NOT NULL DEFAULT current_timestamp(),"+
  "  `expires_in` int(11) NOT NULL,"+
  "  `expires_at` datetime NOT NULL DEFAULT (current_timestamp() + interval `expires_in` second),"+
  "  `subtoken_capabilities` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT NULL CHECK (json_valid(`subtoken_capabilities`)),"+
  "  PRIMARY KEY (`state_h`)"+
  ") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;",
  "/*!40101 SET character_set_client = @saved_cs_client */;",
  ""+
  "--",
  "-- Table structure for table `Events`",
  "--",
  ""+
  "DROP TABLE IF EXISTS `Events`;",
  "/*!40101 SET @saved_cs_client     = @@character_set_client */;",
  "/*!40101 SET character_set_client = utf8 */;",
  "CREATE TABLE `Events` ("+
  "  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,"+
  "  `event` varchar(100) NOT NULL,"+
  "  PRIMARY KEY (`id`),"+
  "  UNIQUE KEY `Events_UN` (`event`)"+
  ") ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;",
  "/*!40101 SET character_set_client = @saved_cs_client */;",
  ""+
  "--",
  "-- Table structure for table `Grants`",
  "--",
  ""+
  "DROP TABLE IF EXISTS `Grants`;",
  "/*!40101 SET @saved_cs_client     = @@character_set_client */;",
  "/*!40101 SET character_set_client = utf8 */;",
  "CREATE TABLE `Grants` ("+
  "  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,"+
  "  `grant_type` varchar(100) NOT NULL,"+
  "  PRIMARY KEY (`id`),"+
  "  UNIQUE KEY `Grants_UN` (`grant_type`)"+
  ") ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;",
  "/*!40101 SET character_set_client = @saved_cs_client */;",
  ""+
  "--",
  "-- Table structure for table `ProxyTokens`",
  "--",
  ""+
  "DROP TABLE IF EXISTS `ProxyTokens`;",
  "/*!40101 SET @saved_cs_client     = @@character_set_client */;",
  "/*!40101 SET character_set_client = utf8 */;",
  "CREATE TABLE `ProxyTokens` ("+
  "  `id` varchar(128) NOT NULL,"+
  "  `jwt` text NOT NULL,"+
  "  PRIMARY KEY (`id`)"+
  ") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;",
  "/*!40101 SET character_set_client = @saved_cs_client */;",
  ""+
  "--",
  "-- Table structure for table `ST_Events`",
  "--",
  ""+
  "DROP TABLE IF EXISTS `ST_Events`;",
  "/*!40101 SET @saved_cs_client     = @@character_set_client */;",
  "/*!40101 SET character_set_client = utf8 */;",
  "CREATE TABLE `ST_Events` ("+
  "  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,"+
  "  `ST_id` varchar(64) NOT NULL,"+
  "  `time` datetime NOT NULL DEFAULT current_timestamp(),"+
  "  `event_id` int(10) unsigned NOT NULL DEFAULT 0,"+
  "  `comment` varchar(100) DEFAULT NULL,"+
  "  `ip` varchar(32) NOT NULL,"+
  "  `user_agent` text NOT NULL,"+
  "  PRIMARY KEY (`id`),"+
  "  KEY `ST_Events_FK_2` (`ST_id`),"+
  "  KEY `ST_Events_FK_3` (`event_id`),"+
  "  CONSTRAINT `ST_Events_FK_2` FOREIGN KEY (`ST_id`) REFERENCES `SuperTokens` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,"+
  "  CONSTRAINT `ST_Events_FK_3` FOREIGN KEY (`event_id`) REFERENCES `Events` (`id`) ON DELETE CASCADE ON UPDATE CASCADE"+
  ") ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;",
  "/*!40101 SET character_set_client = @saved_cs_client */;",
  ""+
  "--",
  "-- Table structure for table `SuperTokens`",
  "--",
  ""+
  "DROP TABLE IF EXISTS `SuperTokens`;",
  "/*!40101 SET @saved_cs_client     = @@character_set_client */;",
  "/*!40101 SET character_set_client = utf8 */;",
  "CREATE TABLE `SuperTokens` ("+
  "  `id` varchar(64) NOT NULL,"+
  "  `parent_id` varchar(64) DEFAULT NULL,"+
  "  `root_id` varchar(64) DEFAULT NULL,"+
  "  `refresh_token` text NOT NULL,"+
  "  `name` varchar(100) DEFAULT NULL,"+
  "  `created` datetime NOT NULL DEFAULT current_timestamp(),"+
  "  `ip_created` varchar(32) NOT NULL,"+
  "  `user_id` bigint(20) unsigned NOT NULL,"+
  "  `rt_hash` char(128) NOT NULL,"+
  "  PRIMARY KEY (`id`),"+
  "  KEY `SessionTokens_parent_id_IDX` (`parent_id`) USING BTREE,"+
  "  KEY `SessionTokens_root_id_IDX` (`root_id`) USING BTREE,"+
  "  KEY `SuperTokens_FK_2` (`user_id`),"+
  "  CONSTRAINT `SuperTokens_FK` FOREIGN KEY (`parent_id`) REFERENCES `SuperTokens` (`id`) ON DELETE SET NULL ON UPDATE CASCADE,"+
  "  CONSTRAINT `SuperTokens_FK_1` FOREIGN KEY (`root_id`) REFERENCES `SuperTokens` (`id`) ON DELETE SET NULL ON UPDATE CASCADE,"+
  "  CONSTRAINT `SuperTokens_FK_2` FOREIGN KEY (`user_id`) REFERENCES `Users` (`id`) ON DELETE CASCADE ON UPDATE CASCADE"+
  ") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;",
  "/*!40101 SET character_set_client = @saved_cs_client */;",
  ""+
  "--",
  "-- Table structure for table `TokenUsages`",
  "--",
  ""+
  "DROP TABLE IF EXISTS `TokenUsages`;",
  "/*!40101 SET @saved_cs_client     = @@character_set_client */;",
  "/*!40101 SET character_set_client = utf8 */;",
  "CREATE TABLE `TokenUsages` ("+
  "  `ST_id` varchar(64) NOT NULL,"+
  "  `restriction` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL CHECK (json_valid(`restriction`)),"+
  "  `usages_AT` int(10) unsigned NOT NULL DEFAULT 0,"+
  "  `usages_other` int(10) unsigned NOT NULL DEFAULT 0,"+
  "  `restriction_hash` char(128) NOT NULL,"+
  "  PRIMARY KEY (`ST_id`,`restriction_hash`),"+
  "  CONSTRAINT `TokenUsages_FK` FOREIGN KEY (`ST_id`) REFERENCES `SuperTokens` (`id`) ON DELETE CASCADE ON UPDATE CASCADE"+
  ") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;",
  "/*!40101 SET character_set_client = @saved_cs_client */;",
  ""+
  "--",
  "-- Temporary table structure for view `TransferCodes`",
  "--",
  ""+
  "DROP TABLE IF EXISTS `TransferCodes`;",
  "/*!50001 DROP VIEW IF EXISTS `TransferCodes`*/;",
  "SET @saved_cs_client     = @@character_set_client;",
  "SET character_set_client = utf8;",
  "/*!50001 CREATE TABLE `TransferCodes` ("+
  "  `id` tinyint NOT NULL,"+
  "  `jwt` tinyint NOT NULL,"+
  "  `created` tinyint NOT NULL,"+
  "  `expires_in` tinyint NOT NULL,"+
  "  `expires_at` tinyint NOT NULL,"+
  "  `revoke_ST` tinyint NOT NULL,"+
  "  `response_type` tinyint NOT NULL,"+
  "  `redirect` tinyint NOT NULL"+
  ") ENGINE=MyISAM */;",
  "SET character_set_client = @saved_cs_client;",
  ""+
  "--",
  "-- Table structure for table `TransferCodesAttributes`",
  "--",
  ""+
  "DROP TABLE IF EXISTS `TransferCodesAttributes`;",
  "/*!40101 SET @saved_cs_client     = @@character_set_client */;",
  "/*!40101 SET character_set_client = utf8 */;",
  "CREATE TABLE `TransferCodesAttributes` ("+
  "  `id` varchar(128) NOT NULL,"+
  "  `created` datetime NOT NULL DEFAULT current_timestamp(),"+
  "  `expires_in` int(11) NOT NULL,"+
  "  `expires_at` datetime NOT NULL DEFAULT (current_timestamp() + interval `expires_in` second),"+
  "  `revoke_ST` bit(1) NOT NULL DEFAULT b'0',"+
  "  `response_type` varchar(128) NOT NULL DEFAULT 'token',"+
  "  `redirect` text DEFAULT NULL,"+
  "  PRIMARY KEY (`id`),"+
  "  CONSTRAINT `TransferCodesAttributes_FK` FOREIGN KEY (`id`) REFERENCES `ProxyTokens` (`id`) ON DELETE CASCADE ON UPDATE CASCADE"+
  ") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;",
  "/*!40101 SET character_set_client = @saved_cs_client */;",
  ""+
  "--",
  "-- Table structure for table `UserGrant_Attributes`",
  "--",
  ""+
  "DROP TABLE IF EXISTS `UserGrant_Attributes`;",
  "/*!40101 SET @saved_cs_client     = @@character_set_client */;",
  "/*!40101 SET character_set_client = utf8 */;",
  "CREATE TABLE `UserGrant_Attributes` ("+
  "  `user_id` bigint(20) unsigned NOT NULL,"+
  "  `grant_id` int(10) unsigned NOT NULL,"+
  "  `attribute_id` int(10) unsigned NOT NULL,"+
  "  `attribute` text NOT NULL,"+
  "  PRIMARY KEY (`user_id`,`grant_id`),"+
  "  KEY `UserGrant_Attributes_FK_3` (`attribute_id`),"+
  "  KEY `UserGrant_Attributes_FK_1` (`grant_id`),"+
  "  CONSTRAINT `UserGrant_Attributes_FK` FOREIGN KEY (`user_id`) REFERENCES `Users` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,"+
  "  CONSTRAINT `UserGrant_Attributes_FK_1` FOREIGN KEY (`grant_id`) REFERENCES `Grants` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,"+
  "  CONSTRAINT `UserGrant_Attributes_FK_3` FOREIGN KEY (`attribute_id`) REFERENCES `Attributes` (`id`) ON DELETE CASCADE ON UPDATE CASCADE"+
  ") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;",
  "/*!40101 SET character_set_client = @saved_cs_client */;",
  ""+
  "--",
  "-- Table structure for table `UserGrants`",
  "--",
  ""+
  "DROP TABLE IF EXISTS `UserGrants`;",
  "/*!40101 SET @saved_cs_client     = @@character_set_client */;",
  "/*!40101 SET character_set_client = utf8 */;",
  "CREATE TABLE `UserGrants` ("+
  "  `user_id` bigint(20) unsigned NOT NULL,"+
  "  `grant_id` int(10) unsigned NOT NULL,"+
  "  `enabled` tinyint(1) NOT NULL,"+
  "  PRIMARY KEY (`user_id`,`grant_id`),"+
  "  KEY `UserGrants_FK` (`grant_id`),"+
  "  CONSTRAINT `UserGrants_FK` FOREIGN KEY (`grant_id`) REFERENCES `Grants` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,"+
  "  CONSTRAINT `UserGrants_FK_1` FOREIGN KEY (`user_id`) REFERENCES `Users` (`id`) ON DELETE CASCADE ON UPDATE CASCADE"+
  ") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;",
  "/*!40101 SET character_set_client = @saved_cs_client */;",
  ""+
  "--",
  "-- Table structure for table `Users`",
  "--",
  ""+
  "DROP TABLE IF EXISTS `Users`;",
  "/*!40101 SET @saved_cs_client     = @@character_set_client */;",
  "/*!40101 SET character_set_client = utf8 */;",
  "CREATE TABLE `Users` ("+
  "  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,"+
  "  `sub` varchar(512) NOT NULL,"+
  "  `iss` varchar(256) NOT NULL,"+
  "  `token_tracing` tinyint(1) NOT NULL DEFAULT 1,"+
  "  `jwt_pk` text DEFAULT NULL,"+
  "  PRIMARY KEY (`id`),"+
  "  UNIQUE KEY `Users_UN` (`sub`,`iss`)"+
  ") ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;",
  "/*!40101 SET character_set_client = @saved_cs_client */;",
  ""+
  "--",
  "-- Final view structure for view `TransferCodes`",
  "--",
  ""+
  "/*!50001 DROP TABLE IF EXISTS `TransferCodes`*/;",
  "/*!50001 DROP VIEW IF EXISTS `TransferCodes`*/;",
  "/*!50001 SET @saved_cs_client          = @@character_set_client */;",
  "/*!50001 SET @saved_cs_results         = @@character_set_results */;",
  "/*!50001 SET @saved_col_connection     = @@collation_connection */;",
  "/*!50001 SET character_set_client      = utf8mb4 */;",
  "/*!50001 SET character_set_results     = utf8mb4 */;",
  "/*!50001 SET collation_connection      = utf8mb4_general_ci */;",
  "/*!50001 CREATE ALGORITHM=UNDEFINED */"+
  "/*!50001 VIEW `TransferCodes` AS select `pt`.`id` AS `id`,`pt`.`jwt` AS `jwt`,`tca`.`created` AS `created`,`tca`.`expires_in` AS `expires_in`,`tca`.`expires_at` AS `expires_at`,`tca`.`revoke_ST` AS `revoke_ST`,`tca`.`response_type` AS `response_type`,`tca`.`redirect` AS `redirect` from (`ProxyTokens` `pt` join `TransferCodesAttributes` `tca` on(`pt`.`id` = `tca`.`id`)) */;",
  "/*!50001 SET character_set_client      = @saved_cs_client */;",
  "/*!50001 SET character_set_results     = @saved_cs_results */;",
  "/*!50001 SET collation_connection      = @saved_col_connection */;",
  "/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;",
  ""+
  "/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;",
  "/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;",
  "/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;",
  "/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;",
  "/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;",
  "/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;",
  "/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;",
}
