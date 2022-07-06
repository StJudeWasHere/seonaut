CREATE TABLE IF NOT EXISTS `videos` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `pagereport_id` int unsigned NOT NULL,
  `url` varchar(2048) NOT NULL DEFAULT '',
  `crawl_id` int unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `videos_pagereport` (`pagereport_id`),
  KEY `videos_crawl` (`crawl_id`),
  CONSTRAINT `videos_crawl` FOREIGN KEY (`crawl_id`) REFERENCES `crawls` (`id`) ON DELETE CASCADE,
  CONSTRAINT `videos_pagereport` FOREIGN KEY (`pagereport_id`) REFERENCES `pagereports` (`id`) ON DELETE CASCADE
);