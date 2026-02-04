-- 业务对象临时表
-- 用于版本控制和编辑中的业务对象
CREATE TABLE IF NOT EXISTS t_business_object_temp (
    id             CHAR(36)     NOT NULL                       COMMENT '业务对象UUID（主键）',
    form_view_id   CHAR(36)     NOT NULL                       COMMENT '关联数据视图UUID',
    user_id        CHAR(36)                                         COMMENT '为空代表模型操作，不为空代表某用户操作',
    version        INT          NOT NULL DEFAULT 10            COMMENT '版本号（存储格式：10=1.0，11=1.1，每次递增1表示0.1版本）',
    object_name    VARCHAR(100) NOT NULL                       COMMENT '业务对象名称',
    created_at     DATETIME(3)          DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    updated_at     DATETIME(3)          DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    deleted_at     DATETIME(3)          DEFAULT NULL           COMMENT '删除时间(逻辑删除)',
    PRIMARY KEY (id),
    KEY idx_form_view_version (form_view_id, version, deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='业务对象临时表';
