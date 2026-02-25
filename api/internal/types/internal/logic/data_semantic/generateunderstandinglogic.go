// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package data_semantic

import (
	"context"

	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/types/internal/svc"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/types/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GenerateUnderstandingLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 一键生成理解数据
func NewGenerateUnderstandingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GenerateUnderstandingLogic {
	return &GenerateUnderstandingLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GenerateUnderstandingLogic) GenerateUnderstanding(req *types.GenerateUnderstandingReq) (resp *types.GenerateUnderstandingResp, err error) {
	// todo: add your logic here and delete this line

	return
}
