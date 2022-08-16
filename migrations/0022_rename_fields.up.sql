ALTER TABLE `crawls` CHANGE `warning_issues` `alert_issues` int NOT NULL DEFAULT '0';
ALTER TABLE `crawls` CHANGE `notice_issues` `warning_issues` int NOT NULL DEFAULT '0';