// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package data_semantic

import (
	"context"

	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/types/internal/svc"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/types/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SaveSemanticInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 保存库表信息补全数据
func NewSaveSemanticInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveSemanticInfoLogic {
	return &SaveSemanticInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SaveSemanticInfoLogic) SaveSemanticInfo(req *types.SaveSemanticInfoReq) (resp *types.SaveSemanticInfoResp, err error) {
	// todo: add your logic here and delete this line

	return
}
