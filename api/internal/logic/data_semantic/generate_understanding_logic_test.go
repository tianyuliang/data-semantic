// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package data_semantic

import (
	"context"
	"testing"

	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/svc"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/types"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/model/form_view"

	"github.com/stretchr/testify/assert"
)

func TestGenerateUnderstandingLogic_GenerateUnderstanding(t *testing.T) {
	// 创建测试上下文
	ctx := context.Background()
	svcCtx := &svc.ServiceContext{
		Config: testConfig,
	}
	logic := NewGenerateUnderstandingLogic(ctx, svcCtx)

	// 测试用例 1: 状态 0 (未理解) - 允许生成
	t.Run("状态0-未理解允许生成", func(t *testing.T) {
		t.Skip("需要数据库连接")

		req := &types.GenerateUnderstandingReq{
			Id: "test-form-view-id-status0",
		}

		resp, err := logic.GenerateUnderstanding(req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, form_view.StatusUnderstanding, resp.UnderstandStatus)
	})

	// 测试用例 2: 状态 3 (已完成) - 允许重新生成
	t.Run("状态3-已完成允许重新生成", func(t *testing.T) {
		t.Skip("需要数据库连接")

		req := &types.GenerateUnderstandingReq{
			Id: "test-form-view-id-status3",
		}

		resp, err := logic.GenerateUnderstanding(req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, form_view.StatusUnderstanding, resp.UnderstandStatus)
	})

	// 测试用例 3: 状态 1 (理解中) - 不允许生成
	t.Run("状态1-理解中不允许生成", func(t *testing.T) {
		t.Skip("需要数据库连接")

		req := &types.GenerateUnderstandingReq{
			Id: "test-form-view-id-status1",
		}

		resp, err := logic.GenerateUnderstanding(req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "当前状态不允许生成")
	})

	// 测试用例 4: 状态 2 (待确认) - 不允许生成
	t.Run("状态2-待确认不允许生成", func(t *testing.T) {
		t.Skip("需要数据库连接")

		req := &types.GenerateUnderstandingReq{
			Id: "test-form-view-id-status2",
		}

		resp, err := logic.GenerateUnderstanding(req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "当前状态不允许生成")
	})

	// 测试用例 5: 状态 4 (已发布) - 不允许生成
	t.Run("状态4-已发布不允许生成", func(t *testing.T) {
		t.Skip("需要数据库连接")

		req := &types.GenerateUnderstandingReq{
			Id: "test-form-view-id-status4",
		}

		resp, err := logic.GenerateUnderstanding(req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "当前状态不允许生成")
	})

	// 测试用例 6: 空ID参数验证
	t.Run("空ID参数验证", func(t *testing.T) {
		t.Skip("需要数据库连接")

		req := &types.GenerateUnderstandingReq{
			Id: "",
		}

		resp, err := logic.GenerateUnderstanding(req)

		// 数据库查询会返回错误，这是预期的行为
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}
