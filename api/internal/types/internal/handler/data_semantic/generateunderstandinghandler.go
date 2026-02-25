// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package data_semantic

import (
	"net/http"

	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/types/internal/logic/data_semantic"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/types/internal/svc"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/types/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 一键生成理解数据
func GenerateUnderstandingHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GenerateUnderstandingReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := data_semantic.NewGenerateUnderstandingLogic(r.Context(), svcCtx)
		resp, err := l.GenerateUnderstanding(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
