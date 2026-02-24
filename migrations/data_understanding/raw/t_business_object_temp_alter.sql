-- 添加业务对象临时表唯一索引
-- 确保 (form_view_id, object_name, version) 组合唯一
ALTER TABLE t_business_object_temp
    ADD UNIQUE KEY uk_form_view_object_version (form_view_id, object_name, version);
