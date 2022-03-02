CREATE TABLE IF NOT EXISTS `users` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `email` varchar(256) NOT NULL DEFAULT '',
  `password` varchar(512) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`)
);

CREATE TABLE IF NOT EXISTS `issue_types` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `type` varchar(256) DEFAULT NULL,
  `priority` int DEFAULT NULL,
  PRIMARY KEY (`id`)
);

INSERT INTO `issue_types` (`id`, `type`, `priority`)
VALUES
	(1,'ERROR_30x',1),
	(2,'ERROR_40x',1),
	(3,'ERROR_50x',1),
	(4,'ERROR_DUPLICATED_TITLE',2),
	(5,'ERROR_DUPLICATED_DESCRIPTION',2),
	(6,'ERROR_EMPTY_TITLE',2),
	(7,'ERROR_SHORT_TITLE',2),
	(8,'ERROR_LONG_TITLE',2),
	(9,'ERROR_EMPTY_DESCRIPTION',2),
	(10,'ERROR_SHORT_DESCRIPTION',2),
	(11,'ERROR_LONG_DESCRIPTION',2),
	(12,'ERROR_LITTLE_CONTENT',3),
	(13,'ERROR_IMAGES_NO_ALT',2),
	(14,'ERROR_REDIRECT_CHAIN',1),
	(15,'ERROR_NO_H1',2),
	(16,'ERROR_NO_LANG',3),
	(17,'ERROR_HTTP_LINKS',2),
	(18,'ERROR_HREFLANG_RETURN',2),
	(19,'ERROR_TOO_MANY_LINKS',3),
	(20,'ERROR_INTERNAL_NOFOLLOW',3),
	(21,'ERROR_EXTERNAL_WITHOUT_NOFOLLOW',3),
	(22,'ERROR_CANONICALIZED_NON_CANONICAL',2),
	(23,'ERROR_REDIRECT_LOOP',1),
	(24,'ERROR_NOT_VALID_HEADINGS',2);

CREATE TABLE IF NOT EXISTS `projects` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `user_id` int unsigned DEFAULT NULL,
  `url` varchar(2048) NOT NULL DEFAULT '',
  `created` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `ignore_robotstxt` tinyint NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `user_id` (`user_id`),
  CONSTRAINT `projects_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS `crawls` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `project_id` int unsigned NOT NULL,
  `start` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `end` timestamp NULL DEFAULT NULL,
  `total_urls` int NOT NULL DEFAULT '0',
  `total_issues` int NOT NULL DEFAULT '0',
  `issues_end` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `crawl_project` (`project_id`),
  CONSTRAINT `crawl_project` FOREIGN KEY (`project_id`) REFERENCES `projects` (`id`) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS `pagereports` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `crawl_id` int unsigned NOT NULL,
  `url` varchar(2048) NOT NULL DEFAULT '',
  `scheme` varchar(5) DEFAULT NULL,
  `redirect_url` varchar(2048) DEFAULT NULL,
  `refresh` varchar(2048) DEFAULT NULL,
  `status_code` int NOT NULL,
  `content_type` varchar(100) DEFAULT NULL,
  `media_type` varchar(100) DEFAULT NULL,
  `lang` varchar(10) DEFAULT NULL,
  `title` varchar(2048) DEFAULT NULL,
  `description` varchar(2048) DEFAULT NULL,
  `robots` varchar(100) DEFAULT NULL,
  `canonical` varchar(2048) DEFAULT NULL,
  `h1` varchar(1024) DEFAULT NULL,
  `h2` varchar(1024) DEFAULT NULL,
  `words` int DEFAULT NULL,
  `size` int DEFAULT NULL,
  `url_hash` varchar(256) NOT NULL DEFAULT '',
  `redirect_hash` varchar(256) DEFAULT NULL,
  `valid_headings` tinyint NOT NULL DEFAULT '1',
  PRIMARY KEY (`id`),
  KEY `pagereport_crawl` (`crawl_id`),
  KEY `pagereport_hash` (`url_hash`),
  KEY `redirect_hash` (`redirect_hash`),
  CONSTRAINT `pagereport_crawl` FOREIGN KEY (`crawl_id`) REFERENCES `crawls` (`id`) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS `external_links` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `pagereport_id` int unsigned NOT NULL,
  `crawl_id` int unsigned NOT NULL,
  `url` varchar(2048) NOT NULL DEFAULT '',
  `rel` varchar(100) DEFAULT NULL,
  `text` varchar(1024) DEFAULT NULL,
  `nofollow` tinyint NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `external_links_pagereport` (`pagereport_id`),
  KEY `external_links_crawl` (`crawl_id`),
  CONSTRAINT `external_links_crawl` FOREIGN KEY (`crawl_id`) REFERENCES `crawls` (`id`) ON DELETE CASCADE,
  CONSTRAINT `external_links_pagereport` FOREIGN KEY (`pagereport_id`) REFERENCES `pagereports` (`id`) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS `hreflangs` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `pagereport_id` int unsigned NOT NULL,
  `crawl_id` int unsigned NOT NULL,
  `from_lang` varchar(10) DEFAULT NULL,
  `to_url` varchar(2048) NOT NULL DEFAULT '',
  `to_lang` varchar(10) DEFAULT NULL,
  `from_hash` varchar(256) NOT NULL DEFAULT '',
  `to_hash` varchar(256) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  KEY `hreflangs_from_hash` (`from_hash`),
  KEY `hreflangs_to_hash` (`to_hash`),
  KEY `hreflangs_pagereport` (`pagereport_id`),
  KEY `hreflangs_crawl` (`crawl_id`),
  KEY `hreflangs_crawl_from_to` (`crawl_id`,`from_hash`,`to_hash`),
  CONSTRAINT `hreflangs_crawl` FOREIGN KEY (`crawl_id`) REFERENCES `crawls` (`id`) ON DELETE CASCADE,
  CONSTRAINT `hreflangs_pagereport` FOREIGN KEY (`pagereport_id`) REFERENCES `pagereports` (`id`) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS `images` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `pagereport_id` int unsigned NOT NULL,
  `url` varchar(2048) NOT NULL DEFAULT '',
  `alt` varchar(1024) DEFAULT NULL,
  `crawl_id` int unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `images_pagereport` (`pagereport_id`),
  KEY `images_crawl` (`crawl_id`),
  CONSTRAINT `images_crawl` FOREIGN KEY (`crawl_id`) REFERENCES `crawls` (`id`) ON DELETE CASCADE,
  CONSTRAINT `images_pagereport` FOREIGN KEY (`pagereport_id`) REFERENCES `pagereports` (`id`) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS `issues` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `pagereport_id` int unsigned NOT NULL,
  `crawl_id` int unsigned NOT NULL,
  `issue_type_id` int unsigned NOT NULL,
  PRIMARY KEY (`id`),
  KEY `issue_crawl` (`crawl_id`),
  KEY `issue_pagereport` (`pagereport_id`),
  KEY `issues_issue_type` (`issue_type_id`),
  CONSTRAINT `issue_crawl` FOREIGN KEY (`crawl_id`) REFERENCES `crawls` (`id`) ON DELETE CASCADE,
  CONSTRAINT `issue_pagereport` FOREIGN KEY (`pagereport_id`) REFERENCES `pagereports` (`id`) ON DELETE CASCADE,
  CONSTRAINT `issues_issue_type` FOREIGN KEY (`issue_type_id`) REFERENCES `issue_types` (`id`) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS `links` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `pagereport_id` int unsigned NOT NULL,
  `crawl_id` int unsigned NOT NULL,
  `url` varchar(2048) NOT NULL DEFAULT '',
  `scheme` varchar(5) NOT NULL,
  `rel` varchar(100) DEFAULT NULL,
  `text` varchar(1024) DEFAULT NULL,
  `url_hash` varchar(256) NOT NULL DEFAULT '',
  `nofollow` tinyint NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `links_external` (`pagereport_id`),
  KEY `links_hash` (`url_hash`),
  KEY `links_crawl` (`crawl_id`),
  CONSTRAINT `links_crawl` FOREIGN KEY (`crawl_id`) REFERENCES `crawls` (`id`) ON DELETE CASCADE,
  CONSTRAINT `links_pagereport` FOREIGN KEY (`pagereport_id`) REFERENCES `pagereports` (`id`) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS `scripts` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `pagereport_id` int unsigned NOT NULL,
  `url` varchar(2048) NOT NULL DEFAULT '',
  `crawl_id` int unsigned NOT NULL,
  PRIMARY KEY (`id`),
  KEY `scripts_pagereport` (`pagereport_id`),
  KEY `scripts_crawl` (`crawl_id`),
  CONSTRAINT `scripts_crawl` FOREIGN KEY (`crawl_id`) REFERENCES `crawls` (`id`) ON DELETE CASCADE,
  CONSTRAINT `scripts_pagereport` FOREIGN KEY (`pagereport_id`) REFERENCES `pagereports` (`id`) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS `styles` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `pagereport_id` int unsigned NOT NULL,
  `url` varchar(2048) NOT NULL DEFAULT '',
  `crawl_id` int unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `styles_pagereport` (`pagereport_id`),
  KEY `styles_crawl` (`crawl_id`),
  CONSTRAINT `styles_crawl` FOREIGN KEY (`crawl_id`) REFERENCES `crawls` (`id`) ON DELETE CASCADE,
  CONSTRAINT `styles_pagereport` FOREIGN KEY (`pagereport_id`) REFERENCES `pagereports` (`id`) ON DELETE CASCADE
);
