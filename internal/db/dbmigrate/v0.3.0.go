package dbmigrate

var v0_3_0_Before = []string{
	// Tables
	"ALTER TABLE AuthInfo ADD rotation json NULL",
	"DROP TRIGGER updtrigger",
	"ALTER TABLE RT_EncryptionKeys ADD CONSTRAINT RT_EncryptionKeys_FK_2 FOREIGN KEY (MT_id) REFERENCES MTokens(id) ON DELETE CASCADE ON UPDATE CASCADE",

	// Predefined Values
	"INSERT IGNORE INTO Events (event) VALUES('token_rotated')",
}
