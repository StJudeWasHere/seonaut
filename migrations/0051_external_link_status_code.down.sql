ALTER TABLE `external_links` DROM COLUMN `status_code`;
ALTER TABLE `projects` DROM COLUMN `check_external_links`;

DELETE FROM issue_types WHERE id = 60;
DELETE FROM issue_types WHERE id = 61;