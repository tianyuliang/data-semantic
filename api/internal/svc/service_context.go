// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"sync"
	"time"

	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/config"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/internal/pkg/aiservice"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"golang.org/x/time/rate"
)

type ServiceContext struct {
	Config       config.Config
	DB           sqlx.SqlConn              // 数据库连接
	Redis        *redis.Redis              // Redis 客户端
	AIClient     aiservice.ClientInterface // AI 服务客户端
	rateLimiters sync.Map                  // formViewId -> *rate.Limiter (限流器缓存)
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 初始化数据库连接
	db := sqlx.NewMysql(c.DB.Default.DataSource())

	// 初始化 Redis
	redisClient := initRedis(c)

	// 初始化 AI 服务客户端（URL 为空时使用 Mock 客户端用于测试）
	timeout := time.Duration(c.AIService.TimeoutSeconds) * time.Second
	var aiClient aiservice.ClientInterface
	if c.AIService.URL == "" {
		aiClient = aiservice.NewMockClient()
	} else {
		aiClient = aiservice.NewClient(c.AIService.URL, timeout)
	}

	return &ServiceContext{
		Config:   c,
		DB:       db,
		Redis:    redisClient,
		AIClient: aiClient,
	}
}

// initRedis 初始化 Redis 客户端
func initRedis(c config.Config) *redis.Redis {
	return redis.MustNewRedis(redis.RedisConf{
		Host: c.Redis.Addr(),
		Pass: c.Redis.Password,
		Type: redis.NodeType,
	})
}

// GetRateLimiter 获取或创建指定 formViewId 的限流器
// 使用 1 秒窗口，允许 1 次请求
func (s *ServiceContext) GetRateLimiter(formViewId string) *rate.Limiter {
	// 尝试从缓存中获取
	if limiter, ok := s.rateLimiters.Load(formViewId); ok {
		return limiter.(*rate.Limiter)
	}

	// 创建新的限流器：1 秒内最多 1 次请求
	limiter := rate.NewLimiter(rate.Every(time.Second), 1)

	// 存入缓存（如果已存在则使用已存在的）
	actual, _ := s.rateLimiters.LoadOrStore(formViewId, limiter)
	return actual.(*rate.Limiter)
}
