ALTER TABLE `crawls` ADD COLUMN `robotstxt_exists` tinyint NOT NULL DEFAULT '0';
ALTER TABLE `crawls` ADD COLUMN `sitemap_exists` tinyint NOT NULL DEFAULT '0';