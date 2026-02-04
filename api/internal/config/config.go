// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/jinguoxing/idrm-go-base/telemetry"
)

type Config struct {
	rest.RestConf

	// Telemetry 配置
	Telemetry telemetry.Config

	// JWT 认证配置
	Auth AuthConfig

	// Swagger 配置
	Swagger SwaggerConfig
}

// AuthConfig JWT 认证配置
type AuthConfig struct {
	AccessSecret string `json:",optional"` // JWT 签名密钥
	AccessExpire int64  `json:",default=7200"` // Token 过期时间(秒)
}

// SwaggerConfig Swagger 文档配置
type SwaggerConfig struct {
	Enabled bool   `json:",default=true"`
	Path    string `json:",default=api/doc/swagger"`
}
