ALTER TABLE `crawls` CHANGE `alert_issues` `warning_issues` int NOT NULL DEFAULT '0';
ALTER TABLE `crawls` CHANGE `warning_issues` `notice_issues` int NOT NULL DEFAULT '0';