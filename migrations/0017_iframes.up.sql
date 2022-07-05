CREATE TABLE IF NOT EXISTS `iframes` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `pagereport_id` int unsigned NOT NULL,
  `url` varchar(2048) NOT NULL DEFAULT '',
  `crawl_id` int unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `iframes_pagereport` (`pagereport_id`),
  KEY `iframes_crawl` (`crawl_id`),
  CONSTRAINT `iframes_crawl` FOREIGN KEY (`crawl_id`) REFERENCES `crawls` (`id`) ON DELETE CASCADE,
  CONSTRAINT `iframes_pagereport` FOREIGN KEY (`pagereport_id`) REFERENCES `pagereports` (`id`) ON DELETE CASCADE
);