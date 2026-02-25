// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package data_semantic

import (
	"context"

	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/types/internal/svc"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/types/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MoveAttributeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 调整属性归属业务对象
func NewMoveAttributeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MoveAttributeLogic {
	return &MoveAttributeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MoveAttributeLogic) MoveAttribute(req *types.MoveAttributeReq) (resp *types.MoveAttributeResp, err error) {
	// todo: add your logic here and delete this line

	return
}
