// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package data_semantic

import (
	"context"

	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/svc"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/types"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/model/data_understanding/business_object"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/model/form_view"

	"github.com/zeromicro/go-zero/core/logx"
)

type BatchObjectMatchLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 批量匹配业务对象
func NewBatchObjectMatchLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchObjectMatchLogic {
	return &BatchObjectMatchLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BatchObjectMatchLogic) BatchObjectMatch(req *types.BatchObjectMatchReq) (resp *types.BatchObjectMatchResp, err error) {
	resp = &types.BatchObjectMatchResp{
		Entries:        make([]types.MatchResult, 0, len(req.Entries)),
		NeedUnderstand: make([]string, 0),
	}

	businessObjectModel := business_object.NewBusinessObjectModelSqlx(l.svcCtx.DB)
	formViewModel := form_view.NewFormViewModel(l.svcCtx.DB)

	// 用于去重
	needUnderstandMap := make(map[string]bool)

	for _, item := range req.Entries {
		// 验证 name 不能为空
		if item.Name == "" {
			continue
		}
		result, needUnderstands := l.processObject(item, businessObjectModel, formViewModel)
		resp.Entries = append(resp.Entries, result)

		// 收集需要理解的视图ID（去重）
		for _, viewId := range needUnderstands {
			if !needUnderstandMap[viewId] {
				needUnderstandMap[viewId] = true
				resp.NeedUnderstand = append(resp.NeedUnderstand, viewId)
			}
		}
	}

	return resp, nil
}

// processObject 处理单个业务对象匹配
// 返回: MatchResult 和 需要理解的视图ID列表
func (l *BatchObjectMatchLogic) processObject(
	item types.SourceObject,
	businessObjectModel business_object.BusinessObjectModel,
	formViewModel form_view.FormViewModel,
) (types.MatchResult, []string) {
	result := types.MatchResult{
		Name:       item.Name,
		DataSource: make([]types.ResponseDataSource, 0),
	}
	needUnderstands := make([]string, 0)

	// Step 1: 如果给定了 data_source，直接追加到结果
	if item.DataSource != nil && item.DataSource.Id != "" {
		result.DataSource = append(result.DataSource, types.ResponseDataSource{
			Id:         item.DataSource.Id,
			Name:       item.DataSource.Name,
			ObjectName: item.Name,
		})
		return result, needUnderstands
	}

	// Step 2: 业务对象表模糊匹配
	logx.Infof("Step 2: 业务对象表模糊匹配, name=%s", item.Name)
	objects, err := businessObjectModel.FuzzyMatchByName(l.ctx, item.Name)
	if err != nil {
		logx.Errorf("FuzzyMatchByName error: %v", err)
		return result, needUnderstands
	}
	logx.Infof("Step 2: 找到 %d 条记录", len(objects))

	if len(objects) > 0 {
		// 找到匹配，组装返回（通过 form_view_id 查询视图信息）
		for _, obj := range objects {
			view, err := formViewModel.FindOneById(l.ctx, obj.FormViewId)
			var viewName string
			var mdlId string
			if err == nil && view != nil {
				viewName = view.TechnicalName
				mdlId = view.MdlId
			}
			result.DataSource = append(result.DataSource, types.ResponseDataSource{
				Id:         mdlId,
				Name:       viewName,
				ObjectName: obj.ObjectName,
			})
		}
		return result, needUnderstands
	}

	// Step 3: 视图表模糊匹配
	views, err := formViewModel.FuzzyMatchByName(l.ctx, item.Name)
	if err != nil {
		logx.Errorf("FuzzyMatchByName for form_view error: %v", err)
		return result, needUnderstands
	}

	if len(views) == 0 {
		// 完全无匹配
		return result, needUnderstands
	}

	// Step 4: 检查每个视图的 understand_status
	for _, view := range views {
		// 无论什么状态都追加视图到结果
		objectName := ""
		if view.UnderstandStatus == 3 {
			// 已理解，查询业务对象表获取 object_name
			objs, _ := businessObjectModel.FindByFormViewId(l.ctx, view.Id)
			if len(objs) > 0 {
				objectName = objs[0].ObjectName
			}
		} else {
			// 未理解，记录需要理解的视图ID
			needUnderstands = append(needUnderstands, view.Id)
		}
		result.DataSource = append(result.DataSource, types.ResponseDataSource{
			Id:         view.MdlId,
			Name:       view.TechnicalName,
			ObjectName: objectName,
		})
	}

	return result, needUnderstands
}
