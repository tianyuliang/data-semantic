-- 库表字段信息临时表
-- 用于版本控制和编辑中的字段语义信息
CREATE TABLE IF NOT EXISTS t_form_view_field_info_temp (
    id                   CHAR(36)     NOT NULL                       COMMENT '记录UUID（主键）',
    form_view_id         CHAR(36)     NOT NULL                       COMMENT '关联数据视图UUID',
    form_view_field_id   CHAR(36)     NOT NULL                       COMMENT '关联字段UUID',
    user_id              CHAR(36)                                         COMMENT '为空代表模型操作，不为空代表某用户操作',
    version              INT          NOT NULL DEFAULT 10            COMMENT '版本号（存储格式：10=1.0，11=1.1，每次递增1表示0.1版本）',
    field_business_name  VARCHAR(255)        DEFAULT NULL            COMMENT '字段业务名称',
    field_role           TINYINT             DEFAULT NULL            COMMENT '字段角色：1-业务主键, 2-关联标识, 3-业务状态, 4-时间字段, 5-业务指标, 6-业务特征, 7-审计字段, 8-技术字段',
    field_description    VARCHAR(300)        DEFAULT NULL            COMMENT '字段描述',
    created_at           DATETIME(3)          DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    updated_at           DATETIME(3)          DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    deleted_at           DATETIME(3)          DEFAULT NULL           COMMENT '删除时间(逻辑删除)',
    PRIMARY KEY (id),
    KEY idx_form_view_version (form_view_id, version, deleted_at),
    KEY idx_form_view_field (form_view_field_id, deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='库表字段信息临时表';
