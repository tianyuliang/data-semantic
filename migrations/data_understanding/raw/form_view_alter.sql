-- form_view 表扩展
-- 添加库表理解状态字段
ALTER TABLE form_view
ADD COLUMN IF NOT EXISTS understand_status TINYINT NOT NULL DEFAULT 0 COMMENT '理解状态：0-未理解,1-理解中,2-待确认,3-已完成,4-已发布';
