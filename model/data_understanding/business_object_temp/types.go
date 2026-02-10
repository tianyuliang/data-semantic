// Package business_object_temp 业务对象临时表Model
package business_object_temp

import "time"

// BusinessObjectTemp 业务对象临时表结构
type BusinessObjectTemp struct {
	Id         string     `db:"id"`
	FormViewId string     `db:"form_view_id"`
	UserId     *string    `db:"user_id"`
	Version    int        `db:"version"`
	ObjectName string     `db:"object_name"`
	FormalId   *string    `db:"formal_id"` // 关联正式表ID（用于增量更新匹配）
	CreatedAt  time.Time  `db:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at"`
	DeletedAt  *time.Time `db:"deleted_at"`
}

// TableName 表名
func (BusinessObjectTemp) TableName() string {
	return "t_business_object_temp"
}
