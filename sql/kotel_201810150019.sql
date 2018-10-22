CREATE TABLE `kotel` (
  `DEVICE_ID` varchar(30) NOT NULL,
  `TO` float NOT NULL DEFAULT '0',
  `TP` float DEFAULT '0',
  `KW` int(11) DEFAULT '0',
  `PR` float DEFAULT '0',
  `DEST_TO` float DEFAULT '0',
  `DEST_TP` float DEFAULT '0',
  `DEST_KW` int(11) DEFAULT '0',
  `DEST_PR` float NOT NULL DEFAULT '0',
  `DEST_TC` float DEFAULT '0'
) ENGINE=InnoDB DEFAULT CHARSET=utf8;