ALTER TABLE `pagereports` DROP COLUMN `valid_lang`;
DELETE FROM issue_types WHERE id = 35;