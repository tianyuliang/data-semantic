// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package data_semantic

import (
	"context"
	"testing"

	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/svc"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/types"

	"github.com/stretchr/testify/assert"
)

func TestGenerateUnderstandingLogic_GenerateUnderstanding(t *testing.T) {
	// 创建测试上下文
	ctx := context.Background()
	svcCtx := &svc.ServiceContext{
		Config: testConfig,
	}
	logic := NewGenerateUnderstandingLogic(ctx, svcCtx)

	// 测试用例 1: 正常生成
	t.Run("正常生成", func(t *testing.T) {
		req := &types.GenerateUnderstandingReq{
			Id: "test-form-view-id",
		}

		resp, err := logic.GenerateUnderstanding(req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, int8(1), resp.UnderstandStatus)
	})

	// 测试用例 2: 空ID
	t.Run("空ID参数验证", func(t *testing.T) {
		req := &types.GenerateUnderstandingReq{
			Id: "",
		}

		resp, err := logic.GenerateUnderstanding(req)

		// 参数验证由 Handler 层处理
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})
}
