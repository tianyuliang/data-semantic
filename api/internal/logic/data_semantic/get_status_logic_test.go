// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package data_semantic

import (
	"context"
	"testing"

	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/config"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/svc"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/types"

	"github.com/stretchr/testify/assert"
)

func TestGetStatusLogic_GetStatus(t *testing.T) {
	// 创建测试上下文
	ctx := context.Background()
	svcCtx := &svc.ServiceContext{
		Config: testConfig,
	}
	logic := NewGetStatusLogic(ctx, svcCtx)

	// 测试用例 1: 正常查询
	t.Run("正常查询", func(t *testing.T) {
		req := &types.GetStatusReq{
			Id: "test-form-view-id",
		}

		resp, err := logic.GetStatus(req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		// TODO: 添加具体的响应断言
		// assert.Equal(t, int8(0), resp.UnderstandStatus)
	})

	// 测试用例 2: 空ID
	t.Run("空ID参数验证", func(t *testing.T) {
		req := &types.GetStatusReq{
			Id: "",
		}

		resp, err := logic.GetStatus(req)

		// 参数验证由 Handler 层处理，这里应该返回错误或默认值
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})
}

// 测试配置
var testConfig = config.Config{}
