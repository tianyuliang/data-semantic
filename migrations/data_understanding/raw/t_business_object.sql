-- 业务对象表
-- 用于存储已发布的业务对象
CREATE TABLE IF NOT EXISTS t_business_object (
    id             CHAR(36)     NOT NULL                       COMMENT '业务对象UUID（主键）',
    object_name    VARCHAR(100) NOT NULL                       COMMENT '业务对象名称',
    object_type    TINYINT      NOT NULL DEFAULT 0             COMMENT '对象类型：0-候选业务对象,1-已发布业务对象',
    form_view_id   CHAR(36)     NOT NULL                       COMMENT '关联数据视图UUID',
    status         TINYINT      NOT NULL DEFAULT 1             COMMENT '状态：0-禁用,1-启用',
    created_at     DATETIME(3)          DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    updated_at     DATETIME(3)          DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    deleted_at     DATETIME(3)          DEFAULT NULL           COMMENT '删除时间(逻辑删除)',
    PRIMARY KEY (`id`),
    KEY idx_form_view_id (form_view_id, deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='业务对象表';
