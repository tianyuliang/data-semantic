-- Kafka 消息处理记录表
-- 用于记录 AI 服务响应的处理状态，防止重复消费
CREATE TABLE IF NOT EXISTS t_kafka_message_log (
    id CHAR(36) NOT NULL COMMENT '主键UUID',
    message_id CHAR(36) NOT NULL COMMENT 'Kafka消息ID',
    form_view_id CHAR(36) NOT NULL COMMENT '关联数据视图UUID',
    processed_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) COMMENT '处理时间',
    status TINYINT DEFAULT 1 COMMENT '状态：1-处理成功，2-处理失败',
    error_msg TEXT COMMENT '错误信息',
    PRIMARY KEY (id),
    UNIQUE KEY uk_message_id (message_id),
    KEY idx_form_view_id (form_view_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Kafka消息处理记录表';
