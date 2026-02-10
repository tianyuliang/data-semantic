// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package data_semantic

import (
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/config"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// 测试配置
var testConfig = config.Config{
	DB: config.DBConfig{
		Default: config.DatabaseConfig{
			Host:     "localhost",
			Port:     3306,
			Database: "data-semantic",
			Username: "root",
			Password: "root123456",
			Charset:  "utf8mb4",
		},
	},
}

// 测试数据库连接
var testDB sqlx.SqlConn

func init() {
	dataSource := testConfig.DB.Default.DataSource()
	testDB = sqlx.NewMysql(dataSource)
}
