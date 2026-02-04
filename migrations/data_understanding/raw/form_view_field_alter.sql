-- form_view_field 表扩展
-- 添加字段语义角色和描述字段
ALTER TABLE form_view_field
ADD COLUMN IF NOT EXISTS field_role TINYINT DEFAULT NULL COMMENT '字段角色：1-业务主键, 2-关联标识, 3-业务状态, 4-时间字段, 5-业务指标, 6-业务特征, 7-审计字段, 8-技术字段' AFTER business_name,
ADD COLUMN IF NOT EXISTS field_description VARCHAR(300) DEFAULT NULL COMMENT '字段描述' AFTER comment;
