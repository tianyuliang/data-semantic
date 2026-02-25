// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package data_semantic

import (
	"context"

	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/types/internal/svc"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/types/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SaveBusinessObjectsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 保存业务对象及属性
func NewSaveBusinessObjectsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveBusinessObjectsLogic {
	return &SaveBusinessObjectsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SaveBusinessObjectsLogic) SaveBusinessObjects(req *types.SaveBusinessObjectsReq) (resp *types.SaveBusinessObjectsResp, err error) {
	// todo: add your logic here and delete this line

	return
}
