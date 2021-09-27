SET @DB_TABLES = CONCAT(@DB, '.*');
PREPARE createUser FROM CONCAT('CREATE OR REPLACE USER ', @USER, ' IDENTIFIED BY "', @PASSWORD, '"');
PREPARE grantRights FROM CONCAT('GRANT Execute, Select, Show view, Insert, Update, Delete ON ', @DB_TABLES, ' TO ',
                                @USER);
EXECUTE createUser;
EXECUTE grantRights;
FLUSH PRIVILEGES;
DEALLOCATE PREPARE createUser;
DEALLOCATE PREPARE grantRights;

