# ************************************************************
# Sequel Ace SQL dump
# Version 20025
#
# https://sequel-ace.com/
# https://github.com/Sequel-Ace/Sequel-Ace
#
# Host: 0.0.0.0 (MySQL 5.7.37)
# Database: seo
# Generation Time: 2022-02-10 08:15:59 +0000
# ************************************************************


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
SET NAMES utf8mb4;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE='NO_AUTO_VALUE_ON_ZERO', SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;


# Dump of table crawls
# ------------------------------------------------------------

DROP TABLE IF EXISTS `crawls`;

CREATE TABLE `crawls` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `project_id` int(11) unsigned NOT NULL,
  `start` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `end` timestamp NULL DEFAULT NULL,
  `total_urls` int(11) NOT NULL DEFAULT '0',
  `total_issues` int(11) NOT NULL DEFAULT '0',
  `issues_end` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `crawl_project` (`project_id`),
  CONSTRAINT `crawl_project` FOREIGN KEY (`project_id`) REFERENCES `projects` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;



# Dump of table hreflangs
# ------------------------------------------------------------

DROP TABLE IF EXISTS `hreflangs`;

CREATE TABLE `hreflangs` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `pagereport_id` int(11) unsigned NOT NULL,
  `crawl_id` int(11) unsigned NOT NULL,
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
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;



# Dump of table images
# ------------------------------------------------------------

DROP TABLE IF EXISTS `images`;

CREATE TABLE `images` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `pagereport_id` int(11) unsigned NOT NULL,
  `url` varchar(2048) NOT NULL DEFAULT '',
  `alt` varchar(1024) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `images_pagereport` (`pagereport_id`),
  CONSTRAINT `images_pagereport` FOREIGN KEY (`pagereport_id`) REFERENCES `pagereports` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;



# Dump of table issues
# ------------------------------------------------------------

DROP TABLE IF EXISTS `issues`;

CREATE TABLE `issues` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `pagereport_id` int(11) unsigned NOT NULL,
  `crawl_id` int(11) unsigned NOT NULL,
  `error_type` varchar(50) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  KEY `issue_crawl` (`crawl_id`),
  KEY `issue_pagereport` (`pagereport_id`),
  CONSTRAINT `issue_crawl` FOREIGN KEY (`crawl_id`) REFERENCES `crawls` (`id`) ON DELETE CASCADE,
  CONSTRAINT `issue_pagereport` FOREIGN KEY (`pagereport_id`) REFERENCES `pagereports` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;



# Dump of table links
# ------------------------------------------------------------

DROP TABLE IF EXISTS `links`;

CREATE TABLE `links` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `pagereport_id` int(11) unsigned NOT NULL,
  `crawl_id` int(11) unsigned NOT NULL,
  `url` varchar(2048) NOT NULL DEFAULT '',
  `scheme` varchar(5) NOT NULL,
  `external` tinyint(1) NOT NULL,
  `rel` varchar(100) DEFAULT NULL,
  `text` varchar(1024) DEFAULT NULL,
  `url_hash` varchar(256) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  KEY `links_external` (`pagereport_id`),
  KEY `links_hash` (`url_hash`),
  KEY `links_crawl` (`crawl_id`),
  CONSTRAINT `links_crawl` FOREIGN KEY (`crawl_id`) REFERENCES `crawls` (`id`) ON DELETE CASCADE,
  CONSTRAINT `links_pagereport` FOREIGN KEY (`pagereport_id`) REFERENCES `pagereports` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;



# Dump of table pagereports
# ------------------------------------------------------------

DROP TABLE IF EXISTS `pagereports`;

CREATE TABLE `pagereports` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `crawl_id` int(11) unsigned NOT NULL,
  `url` varchar(2048) NOT NULL DEFAULT '',
  `scheme` varchar(5) DEFAULT NULL,
  `redirect_url` varchar(2048) DEFAULT NULL,
  `refresh` varchar(2048) DEFAULT NULL,
  `status_code` int(11) NOT NULL,
  `content_type` varchar(100) DEFAULT NULL,
  `media_type` varchar(100) DEFAULT NULL,
  `lang` varchar(10) DEFAULT NULL,
  `title` varchar(2048) DEFAULT NULL,
  `description` varchar(2048) DEFAULT NULL,
  `robots` varchar(100) DEFAULT NULL,
  `canonical` varchar(2048) DEFAULT NULL,
  `h1` varchar(1024) DEFAULT NULL,
  `h2` varchar(1024) DEFAULT NULL,
  `words` int(11) DEFAULT NULL,
  `size` int(11) DEFAULT NULL,
  `url_hash` varchar(256) NOT NULL DEFAULT '',
  `redirect_hash` varchar(256) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `pagereport_crawl` (`crawl_id`),
  KEY `pagereport_hash` (`url_hash`),
  KEY `redirect_hash` (`redirect_hash`),
  CONSTRAINT `pagereport_crawl` FOREIGN KEY (`crawl_id`) REFERENCES `crawls` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;



# Dump of table projects
# ------------------------------------------------------------

DROP TABLE IF EXISTS `projects`;

CREATE TABLE `projects` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `user_id` int(11) unsigned DEFAULT NULL,
  `url` varchar(2048) NOT NULL DEFAULT '',
  `created` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `ignore_robotstxt` tinyint(4) NOT NULL DEFAULT '0',
  `use_javascript` tinyint(4) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `user_id` (`user_id`),
  CONSTRAINT `projects_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;



# Dump of table scripts
# ------------------------------------------------------------

DROP TABLE IF EXISTS `scripts`;

CREATE TABLE `scripts` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `pagereport_id` int(11) unsigned NOT NULL,
  `url` varchar(2048) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  KEY `scripts_pagereport` (`pagereport_id`),
  CONSTRAINT `scripts_pagereport` FOREIGN KEY (`pagereport_id`) REFERENCES `pagereports` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;



# Dump of table styles
# ------------------------------------------------------------

DROP TABLE IF EXISTS `styles`;

CREATE TABLE `styles` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `pagereport_id` int(11) unsigned NOT NULL,
  `url` varchar(2048) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  KEY `styles_pagereport` (`pagereport_id`),
  CONSTRAINT `styles_pagereport` FOREIGN KEY (`pagereport_id`) REFERENCES `pagereports` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;



# Dump of table users
# ------------------------------------------------------------

DROP TABLE IF EXISTS `users`;

CREATE TABLE `users` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `email` varchar(256) NOT NULL DEFAULT '',
  `password` varchar(512) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;




/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;
/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
