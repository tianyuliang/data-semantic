-- 添加表信息临时表唯一索引
-- 确保 (form_view_id, version) 组合唯一
ALTER TABLE t_form_view_info_temp
    ADD UNIQUE KEY uk_form_view_version (form_view_id, version);
