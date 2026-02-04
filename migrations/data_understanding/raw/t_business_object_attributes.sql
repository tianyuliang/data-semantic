-- 业务对象属性表
-- 用于存储已发布的业务对象的属性
CREATE TABLE IF NOT EXISTS t_business_object_attributes (
    id                   CHAR(36)     NOT NULL                       COMMENT '属性UUID（主键）',
    form_view_id         CHAR(36)     NOT NULL                       COMMENT '关联数据视图UUID',
    business_object_id   CHAR(36)     NOT NULL                       COMMENT '关联业务对象UUID',
    form_view_field_id   CHAR(36)     NOT NULL                       COMMENT '关联字段UUID',
    attr_name            VARCHAR(100) NOT NULL                       COMMENT '属性名称',
    created_at           DATETIME(3)          DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    updated_at           DATETIME(3)          DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    deleted_at           DATETIME(3)          DEFAULT NULL           COMMENT '删除时间(逻辑删除)',
    PRIMARY KEY (`id`),
    KEY idx_form_view_id (form_view_id, deleted_at),
    KEY idx_business_object_id (business_object_id, deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='业务对象属性表';
