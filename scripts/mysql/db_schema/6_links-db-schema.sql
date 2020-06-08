CREATE DATABASE falcon_links
  DEFAULT CHARACTER SET utf8
  DEFAULT COLLATE utf8_general_ci;
USE falcon_links;
SET NAMES utf8;

-- ----------------------------
-- Table structure for alert
-- ----------------------------
DROP TABLE IF EXISTS `alert`;
CREATE TABLE `alert` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `path` varchar(16) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `content` text COLLATE utf8_unicode_ci NOT NULL,
  `create_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `alert_path` (`path`)
) ENGINE=InnoDB  DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;
