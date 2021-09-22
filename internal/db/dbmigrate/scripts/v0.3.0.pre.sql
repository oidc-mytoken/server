#  Tables
TRUNCATE TABLE AuthInfo;
ALTER TABLE AuthInfo
    ADD COLUMN IF NOT EXISTS rotation json NULL;
ALTER TABLE AuthInfo
    ADD COLUMN IF NOT EXISTS response_type varchar(128) NOT NULL;
ALTER TABLE AuthInfo
    ADD COLUMN IF NOT EXISTS max_token_len INT DEFAULT NULL NULL;
ALTER TABLE AuthInfo
    ADD COLUMN IF NOT EXISTS code_verifier varchar(128) NULL;
ALTER TABLE TransferCodesAttributes
    ADD COLUMN IF NOT EXISTS max_token_len INT NULL;

CREATE OR REPLACE ALGORITHM = UNDEFINED VIEW `TransferCodes`
AS
SELECT `pt`.`id`                AS `id`,
       `pt`.`jwt`               AS `jwt`,
       `tca`.`created`          AS `created`,
       `tca`.`expires_in`       AS `expires_in`,
       `tca`.`expires_at`       AS `expires_at`,
       `tca`.`revoke_MT`        AS `revoke_MT`,
       `tca`.`response_type`    AS `response_type`,
       `tca`.`max_token_len`    AS `max_token_len`,
       `tca`.`consent_declined` AS `consent_declined`
FROM (`ProxyTokens` `pt`
         JOIN `TransferCodesAttributes` `tca` ON (`pt`.`id` = `tca`.`id`));

# Predefined Values
INSERT IGNORE INTO Events (event)
VALUES ('token_rotated');