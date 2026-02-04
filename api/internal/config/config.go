// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package config

import (
	"fmt"

	"github.com/jinguoxing/idrm-go-base/telemetry"
	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf

	// Telemetry 配置
	Telemetry telemetry.Config

	// JWT 认证配置
	Auth AuthConfig

	// Swagger 配置
	Swagger SwaggerConfig

	// 数据库配置
	DB DBConfig
}

// DBConfig 数据库配置
type DBConfig struct {
	Default DatabaseConfig `json:",default"`
}

// DatabaseConfig 数据库连接配置
type DatabaseConfig struct {
	Host            string `json:",default=localhost"`
	Port            int    `json:",default=3306"`
	Database        string `json:",optional"`
	Username        string `json:",optional"`
	Password        string `json:",optional"`
	Charset         string `json:",default=utf8mb4"`
	MaxIdleConns    int    `json:",default=10"`
	MaxOpenConns    int    `json:",default=100"`
	ConnMaxLifetime int    `json:",default=3600"`
	ConnMaxIdleTime int    `json:",default=600"`
}

// DataSource 生成数据源连接字符串
func (d DatabaseConfig) DataSource() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=true&loc=Local",
		d.Username,
		d.Password,
		d.Host,
		d.Port,
		d.Database,
		d.Charset,
	)
}

// AuthConfig JWT 认证配置
type AuthConfig struct {
	AccessSecret string `json:",optional"`     // JWT 签名密钥
	AccessExpire int64  `json:",default=7200"` // Token 过期时间(秒)
}

// SwaggerConfig Swagger 文档配置
type SwaggerConfig struct {
	Enabled bool   `json:",default=true"`
	Path    string `json:",default=api/doc/swagger"`
}
