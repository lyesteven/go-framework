
CREATE TABLE `trans_msg` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键，唯一',
  `trans_id` varchar(64) COLLATE utf8mb4_bin NOT NULL COMMENT '本地事务ID，唯一',
  `topic` varchar(64) COLLATE utf8mb4_bin NOT NULL DEFAULT 'TransTopic' COMMENT '主题',
  `tag` varchar(64) COLLATE utf8mb4_bin NOT NULL DEFAULT '*' COMMENT '子类型',
  `data` mediumblob NOT NULL COMMENT '消息体，最大64KB',
  `trans_state` tinyint(4) NOT NULL DEFAULT '0' COMMENT '事务执行结果[unknown (0)|commit(1)|rollback(2)]',
  `send_state` tinyint(4) NOT NULL DEFAULT '0' COMMENT '消息发送结果[unsent (0)|sent(1)|fail(2)]',
  `retry_times` tinyint(4) NOT NULL DEFAULT '0' COMMENT '重试次数[0 … 16]',
  `del_flag` tinyint(4) NOT NULL DEFAULT '0' COMMENT '删除标记：0-正常；1-已删除',
  `add_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '入库时间',
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `NewIndex1` (`trans_state`,`send_state`,`retry_times`)
) ENGINE=InnoDB AUTO_INCREMENT=56 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;


