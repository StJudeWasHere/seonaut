ALTER TABLE `projects` ADD COLUMN `schedule_interval` varchar(16) NOT NULL DEFAULT '';
ALTER TABLE `projects` ADD COLUMN `next_run` timestamp NULL DEFAULT NULL;
