ALTER TABLE `pagereports` ADD COLUMN `noindex` tinyint NOT NULL DEFAULT '0';

INSERT INTO issue_types (id, type, priority) VALUES(26, "ERROR_NOFOLLOW_INDEXABLE", 2);
INSERT INTO issue_types (id, type, priority) VALUES(27, "ERROR_NO_INDEXABLE", 3);