package dbmigrate

var v0_3_0_Before = []string{
	// Tables
	"DROP TRIGGER updtrigger",
	"ALTER TABLE RT_EncryptionKeys ADD CONSTRAINT RT_EncryptionKeys_FK_2 FOREIGN KEY (MT_id) REFERENCES MTokens(id) ON DELETE CASCADE ON UPDATE CASCADE",
	"TRUNCATE TABLE AuthInfo",
	"ALTER TABLE AuthInfo ADD rotation json NULL",
	"ALTER TABLE AuthInfo ADD response_type varchar(128) NOT NULL",
	"ALTER TABLE AuthInfo ADD max_token_len INT DEFAULT NULL NULL",
	"ALTER TABLE AuthInfo ADD code_verifier varchar(128) NULL",
	"ALTER TABLE TransferCodesAttributes ADD max_token_len INT NULL",
	"CREATE OR REPLACE" +
		"ALGORITHM = UNDEFINED VIEW `TransferCodes` AS" +
		"select" +
		"`pt`.`id` AS `id`," +
		"`pt`.`jwt` AS `jwt`," +
		"`tca`.`created` AS `created`," +
		"`tca`.`expires_in` AS `expires_in`," +
		"`tca`.`expires_at` AS `expires_at`," +
		"`tca`.`revoke_MT` AS `revoke_MT`," +
		"`tca`.`response_type` AS `response_type`," +
		"`tca`.`max_token_len` AS `max_token_len`," +
		"`tca`.`consent_declined` AS `consent_declined`" +
		"from" +
		"(`ProxyTokens` `pt`" +
		"join `TransferCodesAttributes` `tca` on" +
		"(`pt`.`id` = `tca`.`id`))",

	// Predefined Values
	"INSERT IGNORE INTO Events (event) VALUES('token_rotated')",
}
