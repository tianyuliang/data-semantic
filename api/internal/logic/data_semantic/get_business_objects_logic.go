// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package data_semantic

import (
	"context"
	"fmt"
	"strings"

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
		return nil, fmt.Errorf("查询库表视图失败: %w", err)
	}
	understandStatus := formViewData.UnderstandStatus

	// 2. 状态 1 (理解中) - 返回错误，不允许查询
	if understandStatus == form_view.StatusUnderstanding {
		return nil, fmt.Errorf("当前状态为理解中，请等待处理完成后再查询")
	}

	// 3. 状态 2 (待确认) - 查询临时表最新版本数据
	// 其他状态 (0-未理解, 3-已完成, 4-已发布, 5-理解失败) - 查询正式表，有什么显示什么
	if understandStatus == form_view.StatusPendingConfirm {
		return l.getBusinessObjectsFromTemp(req)
	}
	return l.getBusinessObjectsFromFormal(req)
}

// getBusinessObjectsFromTemp 从临时表查询业务对象（最新版本），并与正式表融合
func (l *GetBusinessObjectsLogic) getBusinessObjectsFromTemp(req *types.GetBusinessObjectsReq) (*types.GetBusinessObjectsResp, error) {
	businessObjectTempModel := business_object_temp.NewBusinessObjectTempModelSqlx(l.svcCtx.DB)
	businessObjectAttrTempModel := business_object_attributes_temp.NewBusinessObjectAttributesTempModelSqlx(l.svcCtx.DB)
	businessObjectModel := business_object.NewBusinessObjectModelSqlx(l.svcCtx.DB)
	businessObjectAttrModel := business_object_attributes.NewBusinessObjectAttributesModelSqlx(l.svcCtx.DB)

	var objects []types.BusinessObject

	// 如果提供了 object_id，查询单个业务对象
	if req.ObjectId != nil {
		// 查询临时表对象
		objDataTemp, err := businessObjectTempModel.FindOneById(l.ctx, *req.ObjectId)
		if err != nil {
			return nil, fmt.Errorf("查询业务对象失败: %w", err)
		}

		// 查询临时表属性
		attrsTemp, err := businessObjectAttrTempModel.FindByBusinessObjectIdWithFieldInfo(l.ctx, *req.ObjectId)
		if err != nil {
			return nil, fmt.Errorf("查询业务对象属性失败: %w", err)
		}

		// 如果有关联的正式表ID，融合正式表数据
		if objDataTemp.FormalId != nil {
			// 查询正式表属性作为基础
			attrsFormal, err := businessObjectAttrModel.FindByBusinessObjectIdWithFieldInfo(l.ctx, *objDataTemp.FormalId)
			if err != nil {
				logx.WithContext(l.ctx).Infof("查询正式表属性失败，仅返回临时表数据: %v", err)
			} else {
				// 融合属性：正式表为基础，临时表为更新
				attrsTemp = l.mergeAttributes(attrsFormal, attrsTemp)
			}
		}

		objects = []types.BusinessObject{{
			Id:         objDataTemp.Id,
			ObjectName: objDataTemp.ObjectName,
			Attributes: l.convertAttrTempToAPI(attrsTemp),
		}}
	} else {
		// 1. 查询正式表所有业务对象（作为基础）
		formalObjList, err := businessObjectModel.FindByFormViewId(l.ctx, req.Id)
		if err != nil && err.Error() != "sql: no rows in result set" {
			logx.WithContext(l.ctx).Infof("查询正式表业务对象失败: %v", err)
		}

		// 构建正式表对象和属性映射
		formalObjMap := make(map[string]*types.BusinessObject) // key: formal_id
		for _, obj := range formalObjList {
			attrs, err := businessObjectAttrModel.FindByBusinessObjectIdWithFieldInfo(l.ctx, obj.Id)
			if err != nil {
				logx.WithContext(l.ctx).Infof("查询正式表业务对象属性失败: %v", err)
				continue
			}
			formalObjMap[obj.Id] = &types.BusinessObject{
				Id:         obj.Id,
				ObjectName: obj.ObjectName,
				Attributes: l.convertAttrFormalToAPI(attrs),
			}
		}

		// 2. 查询临时表最新版本的所有业务对象
		tempObjList, err := businessObjectTempModel.FindByFormViewIdLatest(l.ctx, req.Id)
		if err != nil {
			return nil, fmt.Errorf("查询业务对象列表失败: %w", err)
		}

		// 3. 查询临时表最新版本的所有属性
		allTempAttrs, err := businessObjectAttrTempModel.FindByFormViewIdLatestWithFieldInfo(l.ctx, req.Id)
		if err != nil {
			return nil, fmt.Errorf("查询业务对象属性列表失败: %w", err)
		}

		// 按业务对象ID分组临时表属性
		tempAttrMap := make(map[string][]*business_object_attributes_temp.FieldWithAttrInfoTemp)
		for _, attr := range allTempAttrs {
			tempAttrMap[attr.BusinessObjectId] = append(tempAttrMap[attr.BusinessObjectId], attr)
		}

		// 4. 融合数据
		objects = make([]types.BusinessObject, 0)

		// 先处理与正式表关联的临时表对象
		for _, tempObj := range tempObjList {
			if tempObj.FormalId != nil {
				// 存在正式表关联，进行融合
				if formalObj, exists := formalObjMap[*tempObj.FormalId]; exists {
					// 融合属性：正式表为基础，临时表为更新
					attrsFormal := l.convertAPItoAttrFormal(formalObj.Attributes)
					attrsTemp := tempAttrMap[tempObj.Id]
					mergedAttrs := l.mergeAttributes(attrsFormal, attrsTemp)

					// 使用正式表对象的ID和名称（保持稳定性）
					objects = append(objects, types.BusinessObject{
						Id:         formalObj.Id,
						ObjectName: formalObj.ObjectName,
						Attributes: l.convertAttrTempToAPI(mergedAttrs),
					})
					// 从正式表映射中移除已处理的
					delete(formalObjMap, *tempObj.FormalId)
				}
			}
		}

		// 添加未关联正式表的临时表新对象
		for _, tempObj := range tempObjList {
			if tempObj.FormalId == nil {
				attrsTemp := tempAttrMap[tempObj.Id]
				objects = append(objects, types.BusinessObject{
					Id:         tempObj.Id,
					ObjectName: tempObj.ObjectName,
					Attributes: l.convertAttrTempToAPI(attrsTemp),
				})
			}
		}

		// 添加未被临时表更新的正式表对象
		for _, formalObj := range formalObjMap {
			objects = append(objects, *formalObj)
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
			return nil, fmt.Errorf("查询业务对象失败: %w", err)
		}

		// 查询属性
		attributes, err := businessObjectAttrModel.FindByBusinessObjectIdWithFieldInfo(l.ctx, *req.ObjectId)
		if err != nil {
			return nil, fmt.Errorf("查询业务对象属性失败: %w", err)
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
			return nil, fmt.Errorf("查询业务对象列表失败: %w", err)
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
		result = append(result, l.convertAttrTempItemToAPI(attr))
	}
	return result
}

// convertAttrTempItemToAPI 转换单个临时表属性项
func (l *GetBusinessObjectsLogic) convertAttrTempItemToAPI(attr *business_object_attributes_temp.FieldWithAttrInfoTemp) types.BusinessObjectAttribute {
	return types.BusinessObjectAttribute{
		Id:                attr.Id,
		AttrName:          attr.AttrName,
		FormViewFieldId:   attr.FormViewFieldId,
		FieldTechName:     attr.FieldTechName,
		FieldBusinessName: attr.FieldBusinessName,
		FieldRole:         attr.FieldRole,
		FieldType:         attr.FieldType,
	}
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

// filterByKeyword 按名称过滤业务对象
func (l *GetBusinessObjectsLogic) filterByKeyword(objects []types.BusinessObject, keyword string) []types.BusinessObject {
	result := make([]types.BusinessObject, 0)
	lowerKeyword := strings.ToLower(keyword)
	for _, obj := range objects {
		if strings.Contains(strings.ToLower(obj.ObjectName), lowerKeyword) {
			result = append(result, obj)
		}
	}
	return result
}

// mergeAttributes 融合正式表和临时表的属性
// 规则：正式表为基础，临时表中 form_view_field_id 不存在的属性作为新增/更新
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

// convertAPItoAttrFormal 将API格式转换为正式表属性格式（用于属性融合）
func (l *GetBusinessObjectsLogic) convertAPItoAttrFormal(apiAttrs []types.BusinessObjectAttribute) []*business_object_attributes.FieldWithAttrInfo {
	result := make([]*business_object_attributes.FieldWithAttrInfo, 0, len(apiAttrs))
	for _, attr := range apiAttrs {
		result = append(result, &business_object_attributes.FieldWithAttrInfo{
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
