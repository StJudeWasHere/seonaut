CREATE TABLE IF NOT EXISTS `audios` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `pagereport_id` int unsigned NOT NULL,
  `url` varchar(2048) NOT NULL DEFAULT '',
  `crawl_id` int unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `audios_pagereport` (`pagereport_id`),
  KEY `audios_crawl` (`crawl_id`),
  CONSTRAINT `audios_crawl` FOREIGN KEY (`crawl_id`) REFERENCES `crawls` (`id`) ON DELETE CASCADE,
  CONSTRAINT `audios_pagereport` FOREIGN KEY (`pagereport_id`) REFERENCES `pagereports` (`id`) ON DELETE CASCADE
);