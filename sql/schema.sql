# ************************************************************
# Sequel Pro SQL dump
# Version 4541
#
# http://www.sequelpro.com/
# https://github.com/sequelpro/sequelpro
#
# Host: 0.0.0.0 (MySQL 5.7.36)
# Database: seo
# Generation Time: 2022-01-07 11:51:07 +0000
# ************************************************************


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;


# Dump of table hreflangs
# ------------------------------------------------------------

DROP TABLE IF EXISTS `hreflangs`;

CREATE TABLE `hreflangs` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `pagereport_id` int(11) NOT NULL,
  `url` varchar(2000) NOT NULL DEFAULT '',
  `lang` varchar(10) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;



# Dump of table images
# ------------------------------------------------------------

DROP TABLE IF EXISTS `images`;

CREATE TABLE `images` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `pagereport_id` int(11) NOT NULL,
  `url` varchar(2000) NOT NULL DEFAULT '',
  `alt` varchar(1000) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;



# Dump of table links
# ------------------------------------------------------------

DROP TABLE IF EXISTS `links`;

CREATE TABLE `links` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `pagereport_id` int(11) NOT NULL,
  `url` varchar(2000) NOT NULL DEFAULT '',
  `rel` varchar(100) DEFAULT NULL,
  `text` varchar(1000) DEFAULT NULL,
  `external` tinyint(1) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;



# Dump of table pagereports
# ------------------------------------------------------------

DROP TABLE IF EXISTS `pagereports`;

CREATE TABLE `pagereports` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `url` varchar(2000) NOT NULL DEFAULT '',
  `redirect_url` varchar(2000) DEFAULT NULL,
  `refresh` varchar(2000) DEFAULT NULL,
  `status_code` int(11) NOT NULL,
  `content_type` varchar(100) DEFAULT NULL,
  `lang` varchar(10) DEFAULT NULL,
  `title` varchar(2000) DEFAULT NULL,
  `description` varchar(2000) DEFAULT NULL,
  `robots` varchar(100) DEFAULT NULL,
  `canonical` varchar(2000) DEFAULT NULL,
  `h1` varchar(1000) DEFAULT NULL,
  `h2` varchar(1000) DEFAULT NULL,
  `words` int(11) DEFAULT NULL,
  `size` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;



# Dump of table scripts
# ------------------------------------------------------------

DROP TABLE IF EXISTS `scripts`;

CREATE TABLE `scripts` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `pagereport_id` int(11) NOT NULL,
  `url` varchar(2000) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;



# Dump of table styles
# ------------------------------------------------------------

DROP TABLE IF EXISTS `styles`;

CREATE TABLE `styles` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `pagereport_id` int(11) NOT NULL,
  `url` varchar(2000) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;




/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;
/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
