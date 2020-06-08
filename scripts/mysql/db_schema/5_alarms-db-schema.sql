CREATE DATABASE alarms
  DEFAULT CHARACTER SET utf8
  DEFAULT COLLATE utf8_general_ci;
USE alarms;
SET NAMES utf8;

/*
* 建立告警归档资料表, 主要存储各个告警的最后触发状况
*/
DROP TABLE IF EXISTS `event_cases`;
CREATE TABLE `event_cases` (
  `id` varchar(50) NOT NULL DEFAULT '',
  `endpoint` varchar(100) NOT NULL,
  `metric` varchar(200) NOT NULL,
  `func` varchar(50) DEFAULT NULL,
  `cond` varchar(200) NOT NULL,
  `note` varchar(500) DEFAULT NULL,
  `max_step` int(10) unsigned DEFAULT NULL,
  `current_step` int(10) unsigned DEFAULT NULL,
  `priority` int(6) NOT NULL,
  `status` varchar(20) NOT NULL,
  `timestamp` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `update_at` timestamp NULL DEFAULT NULL,
  `closed_at` timestamp NULL DEFAULT NULL,
  `closed_note` varchar(250) DEFAULT NULL,
  `user_modified` int(10) unsigned DEFAULT NULL,
  `tpl_creator` varchar(64) DEFAULT NULL,
  `expression_id` int(10) unsigned DEFAULT NULL,
  `strategy_id` int(10) unsigned DEFAULT NULL,
  `template_id` int(10) unsigned DEFAULT NULL,
  `process_note` mediumint(9) DEFAULT NULL,
  `process_status` varchar(20) DEFAULT 'unresolved',
  PRIMARY KEY (`id`),
  KEY `endpoint` (`endpoint`,`strategy_id`,`template_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


/*
* 建立告警归档资料表, 存储各个告警触发状况的历史状态
*/
DROP TABLE IF EXISTS `events`;
CREATE TABLE `events` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `event_caseId` varchar(50) DEFAULT NULL,
  `step` int(10) unsigned DEFAULT NULL,
  `cond` varchar(200) NOT NULL,
  `status` int(3) unsigned DEFAULT '0',
  `timestamp` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `event_caseId` (`event_caseId`),
  CONSTRAINT `events_ibfk_1` FOREIGN KEY (`event_caseId`) REFERENCES `event_cases` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB  DEFAULT CHARSET=utf8;


/*
* 告警留言表
*/
DROP TABLE IF EXISTS `event_note`;
CREATE TABLE `event_note` (
  `id` mediumint(9) NOT NULL AUTO_INCREMENT,
  `event_caseId` varchar(50) DEFAULT NULL,
  `note` varchar(300) DEFAULT NULL,
  `case_id` varchar(20) DEFAULT NULL,
  `status` varchar(15) DEFAULT NULL,
  `timestamp` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `user_id` int(10) unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `event_caseId` (`event_caseId`),
  KEY `user_id` (`user_id`),
  CONSTRAINT `event_note_ibfk_1` FOREIGN KEY (`event_caseId`) REFERENCES `event_cases` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `event_note_ibfk_2` FOREIGN KEY (`user_id`) REFERENCES `uic`.`user` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;