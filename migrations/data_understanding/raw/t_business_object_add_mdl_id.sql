-- 业务对象表添加统一视图ID字段
-- 如果列已存在，先删除再添加
-- ALTER TABLE t_business_object DROP COLUMN IF EXISTS mdl_id;
ALTER TABLE t_business_object
ADD COLUMN mdl_id VARCHAR(36) NOT NULL DEFAULT '' COMMENT '统一视图ID' AFTER form_view_id;
