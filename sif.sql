--
-- Table structure for table `sessions`
--

DROP TABLE IF EXISTS `sessions`;
CREATE TABLE `sessions` (
  `hash` varchar(128) NOT NULL,
  `steamid` varchar(63) NOT NULL,
  `creation_timestamp` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `ip` varchar(63) DEFAULT NULL,
  `useragent` varchar(255) DEFAULT NULL,
  `expired` int(10) DEFAULT NULL,
  PRIMARY KEY (`hash`),
  KEY `steamid` (`steamid`),
  CONSTRAINT `sessions_ibfk_1` FOREIGN KEY (`steamid`) REFERENCES `users` (`steamid`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

--
-- Table structure for table `users`
--

DROP TABLE IF EXISTS `users`;
CREATE TABLE `users` (
  `steamid` varchar(63) NOT NULL,
  `creation_timestamp` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `status` tinyint(1) NOT NULL DEFAULT '1',
  `last_activity` int(10) DEFAULT NULL,
  PRIMARY KEY (`steamid`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;
