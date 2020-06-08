CREATE DATABASE graph
  DEFAULT CHARACTER SET utf8
  DEFAULT COLLATE utf8_general_ci;
USE graph;
SET NAMES utf8;

-- ----------------------------
-- Table structure for endpoint
-- ----------------------------
DROP TABLE IF EXISTS `endpoint`;
CREATE TABLE `endpoint` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `endpoint` varchar(255) NOT NULL DEFAULT '',
  `ts` int(11) DEFAULT NULL,
  `t_create` datetime NOT NULL COMMENT 'create time',
  `t_modify` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'last modify time',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_endpoint` (`endpoint`)
) ENGINE=InnoDB  DEFAULT CHARSET=utf8;

-- ----------------------------
-- Table structure for endpoint_counter
-- ----------------------------
DROP TABLE IF EXISTS `endpoint_counter`;
CREATE TABLE `endpoint_counter` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `endpoint_id` bigint(20) unsigned NOT NULL,
  `counter` varchar(255) NOT NULL DEFAULT '',
  `step` int(11) NOT NULL DEFAULT '60' COMMENT 'in second',
  `type` varchar(16) NOT NULL COMMENT 'GAUGE|COUNTER|DERIVE',
  `ts` int(11) DEFAULT NULL,
  `t_create` datetime NOT NULL COMMENT 'create time',
  `t_modify` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'last modify time',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_endpoint_id_counter` (`endpoint_id`,`counter`)
) ENGINE=InnoDB  DEFAULT CHARSET=utf8;

-- ----------------------------
-- Table structure for tag_endpoint
-- ----------------------------
DROP TABLE IF EXISTS `tag_endpoint`;
CREATE TABLE `tag_endpoint` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `tag` varchar(255) NOT NULL DEFAULT '' COMMENT 'srv=tv',
  `endpoint_id` bigint(20) unsigned NOT NULL,
  `ts` int(11) DEFAULT NULL,
  `t_create` datetime NOT NULL COMMENT 'create time',
  `t_modify` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'last modify time',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_tag_endpoint_id` (`tag`,`endpoint_id`)
) ENGINE=InnoDB  DEFAULT CHARSET=utf8;
