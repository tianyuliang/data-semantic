// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package data_semantic

import (
	"context"

	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/types/internal/svc"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/types/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegenerateBusinessObjectsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 重新识别业务对象
func NewRegenerateBusinessObjectsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegenerateBusinessObjectsLogic {
	return &RegenerateBusinessObjectsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegenerateBusinessObjectsLogic) RegenerateBusinessObjects(req *types.RegenerateBusinessObjectsReq) (resp *types.RegenerateBusinessObjectsResp, err error) {
	// todo: add your logic here and delete this line

	return
}
