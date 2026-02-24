-- 添加字段信息临时表唯一索引
-- 确保 (form_view_field_id, version) 组合唯一
ALTER TABLE t_form_view_field_info_temp
    ADD UNIQUE KEY uk_form_view_field_version (form_view_field_id, version);
