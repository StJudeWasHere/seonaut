ALTER TABLE `pagereports` ADD COLUMN `crawled` tinyint NOT NULL DEFAULT '1';

ALTER TABLE `crawls` ADD COLUMN `blocked_by_robotstxt` int NOT NULL DEFAULT 0;
ALTER TABLE `crawls` ADD COLUMN `noindex` int NOT NULL DEFAULT 0;