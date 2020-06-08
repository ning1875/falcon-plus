-- MySQL dump 10.13  Distrib 5.5.31, for Linux (x86_64)
--
-- Host: 127.0.0.1    Database: dashboard
-- ------------------------------------------------------
-- Server version	5.5.31-log

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `dashboard_graph`
--

CREATE DATABASE dashboard
  DEFAULT CHARACTER SET utf8
  DEFAULT COLLATE utf8_general_ci;
USE dashboard;
SET NAMES utf8;

-- ----------------------------
-- Table structure for dashboard_graph
-- ----------------------------
DROP TABLE IF EXISTS `dashboard_graph`;
CREATE TABLE `dashboard_graph` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `title` char(128) NOT NULL,
  `hosts` varchar(10240) NOT NULL DEFAULT '',
  `counters` varchar(1024) NOT NULL DEFAULT '',
  `screen_id` int(11) unsigned NOT NULL,
  `timespan` int(11) unsigned NOT NULL DEFAULT '3600',
  `graph_type` char(2) NOT NULL DEFAULT 'h',
  `method` char(8) DEFAULT '',
  `position` int(11) unsigned NOT NULL DEFAULT '0',
  `falcon_tags` varchar(512) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  KEY `idx_sid` (`screen_id`)
) ENGINE=InnoDB  DEFAULT CHARSET=utf8;

-- ----------------------------
-- Table structure for dashboard_screen
-- ----------------------------
DROP TABLE IF EXISTS `dashboard_screen`;
CREATE TABLE `dashboard_screen` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `pid` int(11) unsigned NOT NULL DEFAULT '0',
  `name` char(128) NOT NULL,
  `time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_pid` (`pid`)
) ENGINE=InnoDB  DEFAULT CHARSET=utf8;

-- ----------------------------
-- Table structure for tmp_graph
-- ----------------------------
DROP TABLE IF EXISTS `tmp_graph`;
CREATE TABLE `tmp_graph` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `endpoints` varchar(10240) NOT NULL DEFAULT '',
  `counters` varchar(10240) NOT NULL DEFAULT '',
  `ck` varchar(32) NOT NULL,
  `time_` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_ck` (`ck`)
) ENGINE=InnoDB  DEFAULT CHARSET=utf8;
