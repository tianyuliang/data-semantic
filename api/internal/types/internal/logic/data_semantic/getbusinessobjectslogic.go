// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package data_semantic

import (
	"context"

	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/types/internal/svc"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/types/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetBusinessObjectsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 查询业务对象识别结果
func NewGetBusinessObjectsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBusinessObjectsLogic {
	return &GetBusinessObjectsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetBusinessObjectsLogic) GetBusinessObjects(req *types.GetBusinessObjectsReq) (resp *types.GetBusinessObjectsResp, err error) {
	// todo: add your logic here and delete this line

	return
}
