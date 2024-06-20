ALTER TABLE `pagereports` ADD COLUMN `ttfb` int NOT NULL DEFAULT '0';
INSERT INTO issue_types (id, type, priority) VALUES(64, "ERROR_SLOW_TTFB", 2);