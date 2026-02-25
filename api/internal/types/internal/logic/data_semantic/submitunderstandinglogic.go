// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package data_semantic

import (
	"context"

	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/types/internal/svc"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/types/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SubmitUnderstandingLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 提交确认理解数据
func NewSubmitUnderstandingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubmitUnderstandingLogic {
	return &SubmitUnderstandingLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SubmitUnderstandingLogic) SubmitUnderstanding(req *types.SubmitUnderstandingReq) (resp *types.SubmitUnderstandingResp, err error) {
	// todo: add your logic here and delete this line

	return
}
