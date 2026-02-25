-- 添加唯一索引，支持基于业务主键的增量更新
-- 用于 INSERT ... ON DUPLICATE KEY UPDATE 语法
ALTER TABLE t_business_object
ADD UNIQUE KEY uk_form_view_object_name (form_view_id, object_name, deleted_at);
