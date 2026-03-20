// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package data_semantic

import (
	"context"

	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/svc"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/types"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/internal/pkg/agentretrieval"

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
		Entries: make([]types.MatchResult, 0, len(req.Entries)),
	}

	// 固定查询数量
	limit := 100

	// 分类处理：需要检索的关键词 和 直接返回的 data_source
	var keywords []string
	results := make([]types.MatchResult, len(req.Entries))

	for i, item := range req.Entries {
		results[i] = types.MatchResult{
			Name:       item.Name,
			DataSource: make([]types.ResponseDataSource, 0),
		}

		// 有 data_source 直接返回
		if item.DataSource != nil && item.DataSource.Id != "" {
			results[i].DataSource = append(results[i].DataSource, types.ResponseDataSource{
				Id:         item.DataSource.Id,
				Name:       item.DataSource.Name,
				ObjectName: item.Name,
			})
			continue
		}

		// 收集需要检索的关键词
		if item.Name != "" {
			keywords = append(keywords, item.Name)
		}
	}

	// 一次性查询所有关键词（使用 OR 条件）
	var allResults []agentretrieval.InstanceData
	if len(keywords) > 0 {
		condition := agentretrieval.Condition{
			Operation: "or",
			SubConditions: buildSubConditions(keywords),
		}

		allResults, err = l.svcCtx.AgentRetrieval.QueryObjectInstance(l.ctx, req.KnId, req.OtId, condition, limit*len(keywords))
		if err != nil {
			logx.Errorf("callAgentRetrieval error: %v", err)
		}
	}

	// 匹配结果到每个关键词
	resultMap := buildResultMap(allResults)
	for i, item := range req.Entries {
		if item.DataSource != nil && item.DataSource.Id != "" {
			continue // 已有 data_source，跳过
		}
		if item.Name == "" {
			continue // 空关键词跳过
		}

		// 匹配该关键词的结果
		if matched, ok := resultMap[item.Name]; ok {
			results[i].DataSource = matched
		}
	}

	resp.Entries = results
	return resp, nil
}

// buildSubConditions 为多个关键词构建子条件
func buildSubConditions(keywords []string) []agentretrieval.SubCondition {
	subConditions := make([]agentretrieval.SubCondition, 0, len(keywords))
	for _, keyword := range keywords {
		subConditions = append(subConditions, agentretrieval.SubCondition{
			Field:     "object_name",
			Operation: "like",
			ValueFrom: "const",
			Value:     keyword,
		})
	}
	return subConditions
}

// buildResultMap 将服务返回的结果按关键词分组
func buildResultMap(datas []agentretrieval.InstanceData) map[string][]types.ResponseDataSource {
	resultMap := make(map[string][]types.ResponseDataSource)
	for _, data := range datas {
		// 使用 object_name 作为 key
		resultMap[data.ObjectName] = append(resultMap[data.ObjectName], types.ResponseDataSource{
			Id:         data.MdlId,
			Name:       data.Display,
			ObjectName: data.ObjectName,
		})
	}
	return resultMap
}
