ALTER TABLE `crawls` ADD COLUMN `critical_issues` int NOT NULL DEFAULT '0';
ALTER TABLE `crawls` ADD COLUMN `warning_issues` int NOT NULL DEFAULT '0';
ALTER TABLE `crawls` ADD COLUMN `notice_issues` int NOT NULL DEFAULT '0';