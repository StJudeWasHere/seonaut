ALTER TABLE `pagereports` ADD COLUMN `valid_lang` tinyint NOT NULL DEFAULT '1';
INSERT INTO issue_types (id, type, priority) VALUES(35, "INVALID_LANG", 2);