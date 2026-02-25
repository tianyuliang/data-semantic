// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package data_semantic

import (
	"context"

	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/types/internal/svc"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/types/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteBusinessObjectsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除识别结果
func NewDeleteBusinessObjectsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteBusinessObjectsLogic {
	return &DeleteBusinessObjectsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteBusinessObjectsLogic) DeleteBusinessObjects(req *types.DeleteBusinessObjectsReq) (resp *types.DeleteBusinessObjectsResp, err error) {
	// todo: add your logic here and delete this line

	return
}
