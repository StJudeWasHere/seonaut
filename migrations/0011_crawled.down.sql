ALTER TABLE `pagereports` DROP COLUMN `crawled`;

ALTER TABLE `crawls` DROP COLUMN `blocked_by_robotstxt`;
ALTER TABLE `crawls` DROP COLUMN `noindex`;