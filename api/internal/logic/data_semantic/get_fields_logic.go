// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package data_semantic

import (
	"context"
	"fmt"
	"strings"

	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/svc"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/types"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/model/data_understanding/form_view_field_info_temp"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/model/data_understanding/form_view_info_temp"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/model/form_view"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/model/form_view_field"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetFieldsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 查询字段语义补全数据
func NewGetFieldsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFieldsLogic {
	return &GetFieldsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetFieldsLogic) GetFields(req *types.GetFieldsReq) (resp *types.GetFieldsResp, err error) {
	logx.Infof("GetFields called with id: %s, keyword: %v, only_incomplete: %v",
		req.Id, req.Keyword, req.OnlyIncomplete)

	// 1. 查询 form_view 表获取 understand_status 和 table_tech_name
	formViewModel := form_view.NewFormViewModel(l.svcCtx.DB)
	tableInfo, err := formViewModel.GetTableInfo(l.ctx, req.Id)
	if err != nil {
		return nil, fmt.Errorf("查询库表视图信息失败: %w", err)
	}

	// 2. 根据状态返回不同数据源
	understandStatus := tableInfo.UnderstandStatus
	tableTechName := tableInfo.TechnicalName

	// 状态 1 (理解中) - 返回错误，不允许查询
	if understandStatus == form_view.StatusUnderstanding {
		return nil, fmt.Errorf("当前状态为理解中，请等待处理完成后再查询")
	}

	// 状态 2 (待确认) - 查询临时表
	if understandStatus == form_view.StatusPendingConfirm {
		return l.getFieldsFromTemp(req, tableTechName)
	}

	// 其他状态 (0-未理解, 3-已完成, 4-已发布, 5-理解失败) - 查询正式表
	return l.getFieldsFromFormal(req, tableTechName)
}

// getFieldsFromTemp 从临时表查询数据，并与正式表融合
func (l *GetFieldsLogic) getFieldsFromTemp(req *types.GetFieldsReq, tableTechName string) (*types.GetFieldsResp, error) {
	// 查询临时表信息
	formViewInfoTempModel := form_view_info_temp.NewFormViewInfoTempModelSqlx(l.svcCtx.DB)
	tableInfoTemp, err := formViewInfoTempModel.FindLatestByFormViewId(l.ctx, req.Id)
	if err != nil {
		return nil, fmt.Errorf("查询库表信息临时表失败: %w", err)
	}

	// 查询字段信息临时表
	formViewFieldInfoTempModel := form_view_field_info_temp.NewFormViewFieldInfoTempModelSqlx(l.svcCtx.DB)
	fieldsTemp, err := formViewFieldInfoTempModel.FindLatestByFormViewId(l.ctx, req.Id)
	if err != nil {
		return nil, fmt.Errorf("查询字段信息临时表失败: %w", err)
	}

	// 查询正式表的完整字段信息（包含语义信息）作为基础
	formViewFieldModel := form_view_field.NewFormViewFieldModel(l.svcCtx.DB)
	formalFields, err := formViewFieldModel.FindFullByFormViewId(l.ctx, req.Id)
	if err != nil {
		// 正式表无数据时继续执行，仅返回临时表数据
		logx.WithContext(l.ctx).Infof("查询正式表完整字段信息失败: %v", err)
		formalFields = []*form_view_field.FormViewField{}
	}

	// 构建字段信息：正式表为基础，临时表为更新
	fields := l.mergeFieldInfo(formalFields, fieldsTemp)

	// 应用过滤条件
	fields = l.applyFilters(fields, req.Keyword, req.OnlyIncomplete)

	return &types.GetFieldsResp{
		CurrentVersion:    tableInfoTemp.Version,
		TableBusinessName: tableInfoTemp.TableBusinessName,
		TableTechName:     tableTechName,
		TableDescription:  tableInfoTemp.TableDescription,
		Fields:            fields,
	}, nil
}

// getFieldsFromFormal 从正式表查询数据
func (l *GetFieldsLogic) getFieldsFromFormal(req *types.GetFieldsReq, tableTechName string) (*types.GetFieldsResp, error) {
	// 查询正式表的字段完整信息 (从 form_view_field 获取包含语义信息的完整数据)
	formViewFieldModel := form_view_field.NewFormViewFieldModel(l.svcCtx.DB)
	fullFields, err := formViewFieldModel.FindFullByFormViewId(l.ctx, req.Id)
	if err != nil {
		return nil, fmt.Errorf("查询字段完整信息失败: %w", err)
	}

	// 构建字段信息
	fields := make([]types.FieldSemanticInfo, 0, len(fullFields))
	for _, f := range fullFields {
		fields = append(fields, types.FieldSemanticInfo{
			FormViewFieldId:   f.Id,
			FieldBusinessName: f.FieldBusinessName,
			FieldTechName:     f.FieldTechName,
			FieldType:         f.FieldType,
			FieldRole:         f.FieldRole,
			FieldDescription:  f.FieldDescription,
		})
	}

	// 应用过滤条件
	fields = l.applyFilters(fields, req.Keyword, req.OnlyIncomplete)

	return &types.GetFieldsResp{
		CurrentVersion:    0, // 正式表无版本号概念
		TableBusinessName: nil,
		TableTechName:     tableTechName,
		TableDescription:  nil,
		Fields:            fields,
	}, nil
}

// buildFieldInfo 构建字段信息
func (l *GetFieldsLogic) buildFieldInfo(fieldsTemp []*form_view_field_info_temp.FormViewFieldInfoTemp, baseFieldMap map[string]*form_view_field.FormViewFieldBase) []types.FieldSemanticInfo {
	fields := make([]types.FieldSemanticInfo, 0, len(fieldsTemp))
	for _, ft := range fieldsTemp {
		baseInfo, exists := baseFieldMap[ft.FormViewFieldId]
		if !exists {
			logx.WithContext(l.ctx).Infof("字段 %s 在基础表 中不存在", ft.FormViewFieldId)
			continue
		}
		fields = append(fields, types.FieldSemanticInfo{
			FormViewFieldId:   ft.FormViewFieldId,
			FieldBusinessName: ft.FieldBusinessName,
			FieldTechName:     baseInfo.FieldTechName,
			FieldType:         baseInfo.FieldType,
			FieldRole:         ft.FieldRole,
			FieldDescription:  ft.FieldDescription,
		})
	}
	return fields
}

// applyFilters 应用过滤条件
func (l *GetFieldsLogic) applyFilters(fields []types.FieldSemanticInfo, keyword *string, onlyIncomplete *bool) []types.FieldSemanticInfo {
	result := fields

	// 关键字过滤
	if keyword != nil && *keyword != "" {
		filtered := make([]types.FieldSemanticInfo, 0)
		for _, f := range result {
			if strings.Contains(strings.ToLower(f.FieldTechName), strings.ToLower(*keyword)) ||
				(f.FieldBusinessName != nil && strings.Contains(strings.ToLower(*f.FieldBusinessName), strings.ToLower(*keyword))) {
				filtered = append(filtered, f)
			}
		}
		result = filtered
	}

	// only_incomplete 过滤 (只返回未补全的字段)
	if onlyIncomplete != nil && *onlyIncomplete {
		filtered := make([]types.FieldSemanticInfo, 0)
		for _, f := range result {
			// 判断是否未补全：field_business_name 为空 或 field_role 为空
			if f.FieldBusinessName == nil || f.FieldRole == nil {
				filtered = append(filtered, f)
			}
		}
		result = filtered
	}

	return result
}

// mergeFieldInfo 融合正式表和临时表的字段信息
// 规则：正式表为基础，临时表中存在的字段用临时表数据覆盖（作为更新）
func (l *GetFieldsLogic) mergeFieldInfo(formalFields []*form_view_field.FormViewField, fieldsTemp []*form_view_field_info_temp.FormViewFieldInfoTemp) []types.FieldSemanticInfo {
	// 构建临时表字段的映射 (key: form_view_field_id)
	tempFieldMap := make(map[string]*form_view_field_info_temp.FormViewFieldInfoTemp)
	for _, ft := range fieldsTemp {
		tempFieldMap[ft.FormViewFieldId] = ft
	}

	// 融合结果
	result := make([]types.FieldSemanticInfo, 0, len(formalFields))

	// 遍历正式表字段，如果临时表存在则用临时表数据，否则用正式表数据
	for _, ff := range formalFields {
		if tempData, exists := tempFieldMap[ff.Id]; exists {
			// 使用临时表数据（更新）
			result = append(result, types.FieldSemanticInfo{
				FormViewFieldId:   ff.Id,
				FieldBusinessName: tempData.FieldBusinessName,
				FieldTechName:     ff.FieldTechName,
				FieldType:         ff.FieldType,
				FieldRole:         tempData.FieldRole,
				FieldDescription:  tempData.FieldDescription,
			})
		} else {
			// 使用正式表数据
			result = append(result, types.FieldSemanticInfo{
				FormViewFieldId:   ff.Id,
				FieldBusinessName: ff.FieldBusinessName,
				FieldTechName:     ff.FieldTechName,
				FieldType:         ff.FieldType,
				FieldRole:         ff.FieldRole,
				FieldDescription:  ff.FieldDescription,
			})
		}
	}

	return result
}
