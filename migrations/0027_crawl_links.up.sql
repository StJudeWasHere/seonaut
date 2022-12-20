ALTER TABLE `crawls` ADD COLUMN `links_internal_follow` int NOT NULL DEFAULT '0';
ALTER TABLE `crawls` ADD COLUMN `links_internal_nofollow` int NOT NULL DEFAULT '0';
ALTER TABLE `crawls` ADD COLUMN `links_external_follow` int NOT NULL DEFAULT '0';
ALTER TABLE `crawls` ADD COLUMN `links_external_nofollow` int NOT NULL DEFAULT '0';
ALTER TABLE `crawls` ADD COLUMN `links_sponsored` int NOT NULL DEFAULT '0';
ALTER TABLE `crawls` ADD COLUMN `links_ugc` int NOT NULL DEFAULT '0';

UPDATE crawls SET
	links_internal_nofollow = (select count(id) from links where crawl_id = crawls.id and nofollow = 1),
	links_internal_follow = (select count(id) from links where crawl_id = crawls.id and nofollow = 0),
	links_external_nofollow = (select count(id) from external_links where crawl_id = crawls.id and nofollow = 1),
	links_external_follow = (select count(id) from external_links where crawl_id = crawls.id and nofollow = 0),
	links_sponsored = (select count(id) from external_links where crawl_id = crawls.id and sponsored = 1),
	links_ugc = (select count(id) from external_links where crawl_id = crawls.id and ugc = 1);