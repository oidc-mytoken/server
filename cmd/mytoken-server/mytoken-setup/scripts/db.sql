PREPARE createDB FROM CONCAT('CREATE DATABASE IF NOT EXISTS ', @DB);
EXECUTE createDB;
DEALLOCATE PREPARE createDB;
