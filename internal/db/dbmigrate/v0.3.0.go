package dbmigrate

var v0_3_0_Before = []string{
	// Tables
	"ALTER TABLE AuthInfo ADD rotation json NULL",
	"DROP TRIGGER updtrigger",

	// Predefined Values
	"INSERT IGNORE INTO Events (event) VALUES('token_rotated')",
}
