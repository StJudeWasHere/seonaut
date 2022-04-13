ALTER TABLE links ADD INDEX `external_crawl_nofollow` (`crawl_id`,`nofollow`);
ALTER TABLE external_links ADD INDEX `external_crawl_nofollow` (`crawl_id`,`nofollow`);
