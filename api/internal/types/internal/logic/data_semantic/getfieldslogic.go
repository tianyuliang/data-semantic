// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package data_semantic

import (
	"context"

	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/types/internal/svc"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/types/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetFieldsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 查询字段语义补全数据
func NewGetFieldsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFieldsLogic {
	return &GetFieldsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetFieldsLogic) GetFields(req *types.GetFieldsReq) (resp *types.GetFieldsResp, err error) {
	// todo: add your logic here and delete this line

	return
}
