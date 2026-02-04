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

func TestSaveSemanticInfoLogic_SaveSemanticInfo(t *testing.T) {
	// 创建测试上下文
	ctx := context.Background()
	svcCtx := &svc.ServiceContext{
		Config: testConfig,
	}
	logic := NewSaveSemanticInfoLogic(ctx, svcCtx)

	// 测试用例 1: 保存库表信息
	t.Run("保存库表信息", func(t *testing.T) {
		tableId := "test-table-info-id"
		tableBusinessName := "测试业务表"
		tableDescription := "测试表描述"

		req := &types.SaveSemanticInfoReq{
			Id: "test-form-view-id",
			TableData: &types.SaveSemanticInfoTableData{
				Id:                &tableId,
				TableBusinessName: &tableBusinessName,
				TableDescription:  &tableDescription,
			},
			FieldData: nil,
		}

		resp, err := logic.SaveSemanticInfo(req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, int32(0), resp.Code)
	})

	// 测试用例 2: 保存字段信息
	t.Run("保存字段信息", func(t *testing.T) {
		fieldId := "test-field-info-id"
		fieldBusinessName := "测试字段"
		fieldRole := int8(1) // 业务主键
		fieldDescription := "测试字段描述"

		req := &types.SaveSemanticInfoReq{
			Id:        "test-form-view-id",
			TableData: nil,
			FieldData: &types.SaveSemanticInfoFieldData{
				Id:                &fieldId,
				FieldBusinessName: &fieldBusinessName,
				FieldRole:         &fieldRole,
				FieldDescription:  &fieldDescription,
			},
		}

		resp, err := logic.SaveSemanticInfo(req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, int32(0), resp.Code)
	})

	// 测试用例 3: 同时保存库表和字段信息
	t.Run("同时保存库表和字段信息", func(t *testing.T) {
		tableId := "test-table-info-id-2"
		tableBusinessName := "测试业务表2"
		fieldId := "test-field-info-id-2"
		fieldBusinessName := "测试字段2"
		fieldRole := int8(2) // 关联标识

		req := &types.SaveSemanticInfoReq{
			Id: "test-form-view-id",
			TableData: &types.SaveSemanticInfoTableData{
				Id:                &tableId,
				TableBusinessName: &tableBusinessName,
				TableDescription:  nil,
			},
			FieldData: &types.SaveSemanticInfoFieldData{
				Id:                &fieldId,
				FieldBusinessName: &fieldBusinessName,
				FieldRole:         &fieldRole,
				FieldDescription:  nil,
			},
		}

		resp, err := logic.SaveSemanticInfo(req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, int32(0), resp.Code)
	})

	// 测试用例 4: 空数据
	t.Run("空数据", func(t *testing.T) {
		req := &types.SaveSemanticInfoReq{
			Id:        "test-form-view-id",
			TableData: nil,
			FieldData: nil,
		}

		resp, err := logic.SaveSemanticInfo(req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, int32(0), resp.Code)
	})
}
