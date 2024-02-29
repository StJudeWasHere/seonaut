ALTER TABLE `external_links` ADD COLUMN `status_code` int NOT NULL DEFAULT 0;
ALTER TABLE `projects` ADD COLUMN `check_external_links` tinyint NOT NULL DEFAULT 0;

INSERT INTO issue_types (id, type, priority) VALUES(60, "ERROR_EXTERNAL_LINK_REDIRECT", 3);
INSERT INTO issue_types (id, type, priority) VALUES(61, "ERROR_EXTERNAL_LINK_BROKEN", 3);