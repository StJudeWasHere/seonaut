ALTER TABLE `pagereports` DROP COLUMN `noindex`;

DELETE FROM issue_types WHERE id = 26;
DELETE FROM issue_types WHERE id = 27;