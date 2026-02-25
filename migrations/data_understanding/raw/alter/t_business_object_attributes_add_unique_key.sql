-- 添加唯一索引，支持基于业务主键的增量更新
-- 用于 INSERT ... ON DUPLICATE KEY UPDATE 语法
ALTER TABLE t_business_object_attributes
ADD UNIQUE KEY uk_business_object_field (business_object_id, attr_name, form_view_field_id, deleted_at);
