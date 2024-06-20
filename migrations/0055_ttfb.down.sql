ALTER TABLE `pagereports` DROP COLUMN `ttfb`;
DELETE FROM issue_types WHERE id = 64;