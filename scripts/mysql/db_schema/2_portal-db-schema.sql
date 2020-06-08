CREATE DATABASE falcon_portal_hl
  DEFAULT CHARACTER SET utf8
  DEFAULT COLLATE utf8_general_ci;
USE falcon_portal_hl;
SET NAMES utf8;

-- ----------------------------
-- Table structure for action
-- ----------------------------
DROP TABLE IF EXISTS `action`;
CREATE TABLE `action` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `uic` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `url` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `callback` tinyint(4) NOT NULL DEFAULT '0',
  `before_callback_sms` tinyint(4) NOT NULL DEFAULT '0',
  `before_callback_mail` tinyint(4) NOT NULL DEFAULT '0',
  `after_callback_sms` tinyint(4) NOT NULL DEFAULT '0',
  `after_callback_mail` tinyint(4) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB  DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

-- ----------------------------
-- Table structure for alert_link
-- ----------------------------
DROP TABLE IF EXISTS `alert_link`;
CREATE TABLE `alert_link` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `path` varchar(16) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `content` text COLLATE utf8_unicode_ci NOT NULL,
  `create_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `alert_path` (`path`)
) ENGINE=InnoDB  DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

-- ----------------------------
-- Table structure for cluster
-- ----------------------------
DROP TABLE IF EXISTS `cluster`;
CREATE TABLE `cluster` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `grp_id` int(11) NOT NULL,
  `numerator` varchar(10240) COLLATE utf8_unicode_ci NOT NULL,
  `denominator` varchar(10240) COLLATE utf8_unicode_ci NOT NULL,
  `endpoint` varchar(255) COLLATE utf8_unicode_ci NOT NULL,
  `metric` varchar(255) COLLATE utf8_unicode_ci NOT NULL,
  `tags` varchar(255) COLLATE utf8_unicode_ci NOT NULL,
  `ds_type` varchar(255) COLLATE utf8_unicode_ci NOT NULL,
  `step` int(11) NOT NULL,
  `last_update` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `creator` varchar(255) COLLATE utf8_unicode_ci NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB  DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

-- ----------------------------
-- Table structure for expression
-- ----------------------------
DROP TABLE IF EXISTS `expression`;
CREATE TABLE `expression` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `expression` varchar(1024) COLLATE utf8_unicode_ci NOT NULL,
  `func` varchar(16) COLLATE utf8_unicode_ci NOT NULL DEFAULT 'all(#1)',
  `op` varchar(8) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `right_value` varchar(16) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `max_step` int(11) NOT NULL DEFAULT '1',
  `priority` tinyint(4) NOT NULL DEFAULT '0',
  `note` varchar(1024) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `action_id` int(10) unsigned NOT NULL DEFAULT '0',
  `create_user` varchar(64) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `pause` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB  DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

-- ----------------------------
-- Table structure for grp
-- ----------------------------
DROP TABLE IF EXISTS `grp`;
CREATE TABLE `grp` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `grp_name` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `create_user` varchar(64) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `create_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `come_from` tinyint(4) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_host_grp_grp_name` (`grp_name`)
) ENGINE=InnoDB  DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

-- ----------------------------
-- Table structure for grp_host
-- ----------------------------
DROP TABLE IF EXISTS `grp_host`;
CREATE TABLE `grp_host` (
  `grp_id` int(10) unsigned NOT NULL,
  `host_id` bigint(20) unsigned NOT NULL,
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  PRIMARY KEY (`id`),
  KEY `idx_grp_host_grp_id` (`grp_id`),
  KEY `idx_grp_host_host_id` (`host_id`)
) ENGINE=InnoDB  DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

-- ----------------------------
-- Table structure for grp_tpl
-- ----------------------------
DROP TABLE IF EXISTS `grp_tpl`;
CREATE TABLE `grp_tpl` (
  `grp_id` int(10) unsigned NOT NULL,
  `tpl_id` int(10) unsigned NOT NULL,
  `bind_user` varchar(64) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  PRIMARY KEY (`id`),
  KEY `idx_grp_tpl_grp_id` (`grp_id`),
  KEY `idx_grp_tpl_tpl_id` (`tpl_id`)
) ENGINE=InnoDB  DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

-- ----------------------------
-- Table structure for host
-- ----------------------------
DROP TABLE IF EXISTS `host`;
CREATE TABLE `host` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `hostname` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `ip` varchar(16) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `agent_version` varchar(16) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `plugin_version` varchar(128) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `maintain_begin` int(10) unsigned NOT NULL DEFAULT '0',
  `maintain_end` int(10) unsigned NOT NULL DEFAULT '0',
  `update_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_host_hostname` (`hostname`)
) ENGINE=InnoDB  DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

-- ----------------------------
-- Table structure for mockcfg
-- ----------------------------
DROP TABLE IF EXISTS `mockcfg`;
CREATE TABLE `mockcfg` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '' COMMENT 'name of mockcfg, used for uuid',
  `obj` varchar(10240) COLLATE utf8_unicode_ci NOT NULL DEFAULT '' COMMENT 'desc of object',
  `obj_type` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '' COMMENT 'type of object, host or group or other',
  `metric` varchar(128) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `tags` varchar(1024) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `dstype` varchar(32) COLLATE utf8_unicode_ci NOT NULL DEFAULT 'GAUGE',
  `step` int(11) unsigned NOT NULL DEFAULT '60',
  `mock` double NOT NULL DEFAULT '0' COMMENT 'mocked value when nodata occurs',
  `creator` varchar(64) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `t_create` datetime NOT NULL COMMENT 'create time',
  `t_modify` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'last modify time',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_name` (`name`)
) ENGINE=InnoDB  DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

-- ----------------------------
-- Table structure for plugin_dir
-- ----------------------------
DROP TABLE IF EXISTS `plugin_dir`;
CREATE TABLE `plugin_dir` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `grp_id` int(10) unsigned NOT NULL,
  `dir` varchar(255) COLLATE utf8_unicode_ci NOT NULL,
  `create_user` varchar(64) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `create_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_plugin_dir_grp_id` (`grp_id`)
) ENGINE=InnoDB  DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

-- ----------------------------
-- Table structure for strategy
-- ----------------------------
DROP TABLE IF EXISTS `strategy`;
CREATE TABLE `strategy` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `metric` varchar(128) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `tags` varchar(256) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `max_step` int(11) NOT NULL DEFAULT '1',
  `priority` tinyint(4) NOT NULL DEFAULT '0',
  `func` varchar(16) COLLATE utf8_unicode_ci NOT NULL DEFAULT 'all(#1)',
  `op` varchar(8) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `right_value` varchar(64) COLLATE utf8_unicode_ci NOT NULL,
  `note` varchar(128) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `run_begin` varchar(16) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `run_end` varchar(16) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `tpl_id` int(10) unsigned NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_strategy_tpl_id` (`tpl_id`)
) ENGINE=InnoDB  DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

-- ----------------------------
-- Table structure for tpl
-- ----------------------------
DROP TABLE IF EXISTS `tpl`;
CREATE TABLE `tpl` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `tpl_name` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `parent_id` int(10) unsigned NOT NULL DEFAULT '0',
  `action_id` int(10) unsigned NOT NULL DEFAULT '0',
  `create_user` varchar(64) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `create_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_tpl_name` (`tpl_name`),
  KEY `idx_tpl_create_user` (`create_user`)
) ENGINE=InnoDB  DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;
