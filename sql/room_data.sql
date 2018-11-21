CREATE TABLE `room_data` (
  `device_id` varchar(20) NOT NULL,
  `sensor_type` varchar(30) NOT NULL,
  `t` float DEFAULT NULL,
  `h` float DEFAULT NULL,
  `p` float DEFAULT NULL,
  `date` datetime NOT NULL,
  KEY `room_data_device_id_IDX` (`device_id`,`date`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
