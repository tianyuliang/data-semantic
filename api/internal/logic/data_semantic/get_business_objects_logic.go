// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package data_semantic

import (
	"context"
	"strings"

	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/errorx"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/svc"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/types"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/model/data_understanding/business_object"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/model/data_understanding/business_object_attributes"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/model/data_understanding/business_object_attributes_temp"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/model/data_understanding/business_object_temp"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/model/form_view"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetBusinessObjectsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 查询业务对象识别结果
func NewGetBusinessObjectsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBusinessObjectsLogic {
	return &GetBusinessObjectsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetBusinessObjectsLogic) GetBusinessObjects(req *types.GetBusinessObjectsReq) (resp *types.GetBusinessObjectsResp, err error) {
	logx.Infof("GetBusinessObjects called with id: %s, objectId: %v, keyword: %v",
		req.Id, req.ObjectId, req.Keyword)

	// 1. 查询 form_view 的 understand_status
	formViewModel := form_view.NewFormViewModel(l.svcCtx.DB)
	formViewData, err := formViewModel.FindOneById(l.ctx, req.Id)
	if err != nil {
		return nil, errorx.NewQueryFailed("库表视图", err)
	}
	understandStatus := formViewData.UnderstandStatus

	// 2. 状态 1 (理解中) - 返回错误，不允许查询
	if understandStatus == form_view.StatusUnderstanding {
		return nil, errorx.Newf(errorx.ErrCodeInvalidUnderstandStatus,
			"当前状态为理解中，请等待处理完成后再查询")
	}

	// 3. 状态 2 (待确认) - 查询临时表最新版本数据
	// 其他状态 (0-未理解, 3-已完成, 4-已发布, 5-理解失败) - 查询正式表，有什么显示什么
	if understandStatus == form_view.StatusPendingConfirm {
		return l.getBusinessObjectsFromTemp(req)
	}
	return l.getBusinessObjectsFromFormal(req)
}

// getBusinessObjectsFromTemp 从临时表查询业务对象（最新版本）
func (l *GetBusinessObjectsLogic) getBusinessObjectsFromTemp(req *types.GetBusinessObjectsReq) (*types.GetBusinessObjectsResp, error) {
	businessObjectTempModel := business_object_temp.NewBusinessObjectTempModelSqlx(l.svcCtx.DB)
	businessObjectAttrTempModel := business_object_attributes_temp.NewBusinessObjectAttributesTempModelSqlx(l.svcCtx.DB)

	var objects []types.BusinessObject

	// 如果提供了 object_id，查询单个业务对象
	if req.ObjectId != nil {
		// 查询临时表对象
		objDataTemp, err := businessObjectTempModel.FindOneById(l.ctx, *req.ObjectId)
		if err != nil {
			return nil, errorx.NewQueryFailed("业务对象", err)
		}

		// 查询临时表属性
		attrsTemp, err := businessObjectAttrTempModel.FindByBusinessObjectIdWithFieldInfo(l.ctx, *req.ObjectId)
		if err != nil {
			return nil, errorx.NewQueryFailed("业务对象属性", err)
		}

		// 检查正式表是否存在该业务对象（通过 object_name 匹配）
		businessObjectModel := business_object.NewBusinessObjectModelSqlx(l.svcCtx.DB)
		businessObjectAttrModel := business_object_attributes.NewBusinessObjectAttributesModelSqlx(l.svcCtx.DB)

		objFormal, err := businessObjectModel.FindByFormViewIdAndObjectName(l.ctx, objDataTemp.FormViewId, objDataTemp.ObjectName)
		if err != nil || objFormal == nil {
			// 正式表不存在该对象，直接返回临时表数据（不涉及融合）
			objects = []types.BusinessObject{{
				Id:         objDataTemp.Id,
				ObjectName: objDataTemp.ObjectName,
				Attributes: l.convertAttrTempToAPI(attrsTemp),
			}}
		} else {
			// 正式表存在该对象，查询正式表属性并进行融合
			attrsFormal, err := businessObjectAttrModel.FindByBusinessObjectIdWithFieldInfo(l.ctx, objFormal.Id)
			if err != nil {
				logx.WithContext(l.ctx).Infof("查询正式表属性失败，仅返回临时表数据: %v", err)
				objects = []types.BusinessObject{{
					Id:         objDataTemp.Id,
					ObjectName: objDataTemp.ObjectName,
					Attributes: l.convertAttrTempToAPI(attrsTemp),
				}}
			} else {
				// 融合属性：正式表为基础，临时表中 form_view_field_id 不存在的属性作为新增
				attrsTemp = l.mergeAttributes(attrsFormal, attrsTemp)
				// 使用正式表对象的ID（保持稳定性）
				objects = []types.BusinessObject{{
					Id:         objFormal.Id,
					ObjectName: objFormal.ObjectName,
					Attributes: l.convertAttrTempToAPI(attrsTemp),
				}}
			}
		}
	} else {
		// 查询临时表最新版本的所有业务对象
		tempObjList, err := businessObjectTempModel.FindByFormViewIdLatest(l.ctx, req.Id)
		if err != nil {
			return nil, errorx.NewQueryFailed("业务对象列表", err)
		}

		// 查询临时表最新版本的所有属性
		allTempAttrs, err := businessObjectAttrTempModel.FindByFormViewIdLatestWithFieldInfo(l.ctx, req.Id)
		if err != nil {
			return nil, errorx.NewQueryFailed("业务对象属性列表", err)
		}

		// 按业务对象ID分组临时表属性
		tempAttrMap := make(map[string][]*business_object_attributes_temp.FieldWithAttrInfoTemp)
		for _, attr := range allTempAttrs {
			tempAttrMap[attr.BusinessObjectId] = append(tempAttrMap[attr.BusinessObjectId], attr)
		}

		// 构建响应
		objects = make([]types.BusinessObject, 0, len(tempObjList))
		for _, obj := range tempObjList {
			attrsTemp, ok := tempAttrMap[obj.Id]
			if !ok {
				attrsTemp = []*business_object_attributes_temp.FieldWithAttrInfoTemp{}
			}
			objects = append(objects, types.BusinessObject{
				Id:         obj.Id,
				ObjectName: obj.ObjectName,
				Attributes: l.convertAttrTempToAPI(attrsTemp),
			})
		}

		// 如果提供了 keyword，按名称过滤
		if req.Keyword != nil && *req.Keyword != "" {
			objects = l.filterByKeyword(objects, *req.Keyword)
		}
	}

	return &types.GetBusinessObjectsResp{
		List: objects,
	}, nil
}

// getBusinessObjectsFromFormal 从正式表查询业务对象
func (l *GetBusinessObjectsLogic) getBusinessObjectsFromFormal(req *types.GetBusinessObjectsReq) (*types.GetBusinessObjectsResp, error) {
	businessObjectModel := business_object.NewBusinessObjectModelSqlx(l.svcCtx.DB)
	businessObjectAttrModel := business_object_attributes.NewBusinessObjectAttributesModelSqlx(l.svcCtx.DB)

	var objects []types.BusinessObject

	// 如果提供了 object_id，查询单个业务对象
	if req.ObjectId != nil {
		objData, err := businessObjectModel.FindOneById(l.ctx, *req.ObjectId)
		if err != nil {
			return nil, errorx.NewQueryFailed("业务对象", err)
		}

		// 查询属性
		attributes, err := businessObjectAttrModel.FindByBusinessObjectIdWithFieldInfo(l.ctx, *req.ObjectId)
		if err != nil {
			return nil, errorx.NewQueryFailed("业务对象属性", err)
		}

		objects = []types.BusinessObject{{
			Id:         objData.Id,
			ObjectName: objData.ObjectName,
			Attributes: l.convertAttrFormalToAPI(attributes),
		}}
	} else {
		// 查询所有业务对象
		objList, err := businessObjectModel.FindByFormViewId(l.ctx, req.Id)
		if err != nil {
			return nil, errorx.NewQueryFailed("业务对象列表", err)
		}

		// 构建响应
		objects = make([]types.BusinessObject, 0, len(objList))
		for _, obj := range objList {
			// 查询每个业务对象的属性
			attributes, err := businessObjectAttrModel.FindByBusinessObjectIdWithFieldInfo(l.ctx, obj.Id)
			if err != nil {
				logx.WithContext(l.ctx).Errorf("查询业务对象属性失败: %v", err)
				continue
			}

			objects = append(objects, types.BusinessObject{
				Id:         obj.Id,
				ObjectName: obj.ObjectName,
				Attributes: l.convertAttrFormalToAPI(attributes),
			})
		}

		// 如果提供了 keyword，按名称过滤
		if req.Keyword != nil && *req.Keyword != "" {
			objects = l.filterByKeyword(objects, *req.Keyword)
		}
	}

	return &types.GetBusinessObjectsResp{
		List: objects,
	}, nil
}

// convertAttrTempToAPI 转换临时表属性到API格式
func (l *GetBusinessObjectsLogic) convertAttrTempToAPI(attrs []*business_object_attributes_temp.FieldWithAttrInfoTemp) []types.BusinessObjectAttribute {
	result := make([]types.BusinessObjectAttribute, 0, len(attrs))
	for _, attr := range attrs {
		result = append(result, types.BusinessObjectAttribute{
			Id:                attr.Id,
			AttrName:          attr.AttrName,
			FormViewFieldId:   attr.FormViewFieldId,
			FieldTechName:     attr.FieldTechName,
			FieldBusinessName: attr.FieldBusinessName,
			FieldRole:         attr.FieldRole,
			FieldType:         attr.FieldType,
		})
	}
	return result
}

// convertAttrFormalToAPI 转换正式表属性到API格式
func (l *GetBusinessObjectsLogic) convertAttrFormalToAPI(attrs []*business_object_attributes.FieldWithAttrInfo) []types.BusinessObjectAttribute {
	result := make([]types.BusinessObjectAttribute, 0, len(attrs))
	for _, attr := range attrs {
		result = append(result, types.BusinessObjectAttribute{
			Id:                attr.Id,
			AttrName:          attr.AttrName,
			FormViewFieldId:   attr.FormViewFieldId,
			FieldTechName:     attr.FieldTechName,
			FieldBusinessName: attr.FieldBusinessName,
			FieldRole:         attr.FieldRole,
			FieldType:         attr.FieldType,
		})
	}
	return result
}

// filterByKeyword 按属性名称/字段业务名称过滤业务对象
func (l *GetBusinessObjectsLogic) filterByKeyword(objects []types.BusinessObject, keyword string) []types.BusinessObject {
	result := make([]types.BusinessObject, 0)
	lowerKeyword := strings.ToLower(keyword)

	for _, obj := range objects {
		// 检查对象名称、属性名称或字段业务名称是否匹配关键词
		matched := strings.Contains(strings.ToLower(obj.ObjectName), lowerKeyword)
		if !matched {
			// 检查属性中是否匹配
			for _, attr := range obj.Attributes {
				if strings.Contains(strings.ToLower(attr.AttrName), lowerKeyword) ||
					(attr.FieldBusinessName != nil && strings.Contains(strings.ToLower(*attr.FieldBusinessName), lowerKeyword)) {
					matched = true
					break
				}
			}
		}
		if matched {
			result = append(result, obj)
		}
	}
	return result
}

// mergeAttributes 融合正式表和临时表的属性
// 规则：正式表为基础，临时表中 form_view_field_id 不存在的属性作为新增
func (l *GetBusinessObjectsLogic) mergeAttributes(attrsFormal []*business_object_attributes.FieldWithAttrInfo, attrsTemp []*business_object_attributes_temp.FieldWithAttrInfoTemp) []*business_object_attributes_temp.FieldWithAttrInfoTemp {
	// 构建正式表属性的 form_view_field_id 集合（用于快速查找）
	formalFieldIds := make(map[string]bool)
	for _, attr := range attrsFormal {
		formalFieldIds[attr.FormViewFieldId] = true
	}

	// 融合结果：正式表属性 + 临时表中的新增/更新属性
	result := make([]*business_object_attributes_temp.FieldWithAttrInfoTemp, 0, len(attrsFormal)+len(attrsTemp))

	// 1. 先将正式表属性转换为临时表格式
	for _, attr := range attrsFormal {
		result = append(result, &business_object_attributes_temp.FieldWithAttrInfoTemp{
			Id:                attr.Id,
			BusinessObjectId:  "", // 临时表中不需要
			AttrName:          attr.AttrName,
			FormViewFieldId:   attr.FormViewFieldId,
			FieldTechName:     attr.FieldTechName,
			FieldBusinessName: attr.FieldBusinessName,
			FieldRole:         attr.FieldRole,
			FieldType:         attr.FieldType,
		})
	}

	// 2. 添加临时表中的新增/更新属性（form_view_field_id 在正式表中不存在的）
	for _, attr := range attrsTemp {
		if !formalFieldIds[attr.FormViewFieldId] {
			result = append(result, attr)
		}
	}

	return result
}
