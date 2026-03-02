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
		return nil, errorx.Detail(errorx.QueryFailed, err, "库表视图")
	}
	understandStatus := formViewData.UnderstandStatus

	// 2. 状态 1 (理解中) - 返回错误，不允许查询
	if understandStatus == form_view.StatusUnderstanding {
		return nil, errorx.Desc(errorx.InvalidUnderstandStatus)
	}

	// 3. 状态 2 (待确认) - 查询临时表最新版本数据
	// 其他状态 (0-未理解, 3-已完成, 4-已发布, 5-理解失败) - 查询正式表，有什么显示什么
	if understandStatus == form_view.StatusPendingConfirm {
		return l.getBusinessObjectsFromTemp(req)
	}
	return l.getBusinessObjectsFromFormal(req)
}

// getBusinessObjectsFromTemp 从临时表查询业务对象（最新版本，融合正式表）
func (l *GetBusinessObjectsLogic) getBusinessObjectsFromTemp(req *types.GetBusinessObjectsReq) (*types.GetBusinessObjectsResp, error) {
	tempModel := business_object_temp.NewBusinessObjectTempModelSqlx(l.svcCtx.DB)
	tempAttrModel := business_object_attributes_temp.NewBusinessObjectAttributesTempModelSqlx(l.svcCtx.DB)
	formalModel := business_object.NewBusinessObjectModelSqlx(l.svcCtx.DB)
	formalAttrModel := business_object_attributes.NewBusinessObjectAttributesModelSqlx(l.svcCtx.DB)

	var objects []types.BusinessObject
	var err error

	if req.ObjectId != nil {
		// 查询单个对象
		objects, err = l.buildObjectFromTemp(*req.ObjectId, tempModel, tempAttrModel, formalModel, formalAttrModel)
	} else {
		// 查询所有对象
		objects, err = l.buildAllObjectsFromTemp(req.Id, tempModel, tempAttrModel, formalModel, formalAttrModel)
		if req.Keyword != nil && *req.Keyword != "" {
			objects = l.filterByKeyword(objects, *req.Keyword)
		}
	}

	if err != nil {
		return nil, err
	}

	// 查询未识别字段
	unidentifiedFields, err := l.getUnidentifiedFields(req.Id, tempAttrModel)
	if err != nil {
		logx.WithContext(l.ctx).Infof("查询未识别字段失败: %v", err)
		unidentifiedFields = []types.UnidentifiedField{} // 查询失败时返回空列表
	}

	return &types.GetBusinessObjectsResp{
		List:               objects,
		UnidentifiedFields: unidentifiedFields,
	}, nil
}

// buildObjectFromTemp 构建单个业务对象（含融合）
func (l *GetBusinessObjectsLogic) buildObjectFromTemp(
	objectId string,
	tempModel *business_object_temp.BusinessObjectTempModelSqlx,
	tempAttrModel *business_object_attributes_temp.BusinessObjectAttributesTempModelSqlx,
	formalModel *business_object.BusinessObjectModelSqlx,
	formalAttrModel *business_object_attributes.BusinessObjectAttributesModelSqlx,
) ([]types.BusinessObject, error) {
	objTemp, err := tempModel.FindOneById(l.ctx, objectId)
	if err != nil {
		return nil, errorx.Detail(errorx.QueryFailed, err, "业务对象")
	}

	attrsTemp, err := tempAttrModel.FindByBusinessObjectIdWithFieldInfo(l.ctx, objectId)
	if err != nil {
		return nil, errorx.Detail(errorx.QueryFailed, err, "业务对象属性")
	}

	// 检查正式表是否存在同名对象
	objFormal, _ := formalModel.FindByFormViewIdAndObjectName(l.ctx, objTemp.FormViewId, objTemp.ObjectName)

	if objFormal == nil {
		// 正式表不存在，返回临时表数据
		return []types.BusinessObject{{
			Id:         objTemp.Id,
			ObjectName: objTemp.ObjectName,
			Attributes: l.toAPIAttrs(attrsTemp),
		}}, nil
	}

	// 正式表存在，融合属性
	attrsFormal, err := formalAttrModel.FindByBusinessObjectIdWithFieldInfo(l.ctx, objFormal.Id)
	if err != nil {
		logx.WithContext(l.ctx).Infof("查询正式表属性失败: %v", err)
		return []types.BusinessObject{{
			Id:         objTemp.Id,
			ObjectName: objTemp.ObjectName,
			Attributes: l.toAPIAttrs(attrsTemp),
		}}, nil
	}

	mergedAttrs := l.mergeAttrs(attrsFormal, attrsTemp)
	return []types.BusinessObject{{
		Id:         objFormal.Id,
		ObjectName: objFormal.ObjectName,
		Attributes: l.toAPIAttrs(mergedAttrs),
	}}, nil
}

// buildAllObjectsFromTemp 构建所有业务对象（含融合）
func (l *GetBusinessObjectsLogic) buildAllObjectsFromTemp(
	formViewId string,
	tempModel *business_object_temp.BusinessObjectTempModelSqlx,
	tempAttrModel *business_object_attributes_temp.BusinessObjectAttributesTempModelSqlx,
	formalModel *business_object.BusinessObjectModelSqlx,
	formalAttrModel *business_object_attributes.BusinessObjectAttributesModelSqlx,
) ([]types.BusinessObject, error) {
	// 查询正式表对象映射（key: object_name）
	formalObjs, _ := formalModel.FindByFormViewId(l.ctx, formViewId)
	formalObjMap := make(map[string]*business_object.BusinessObject)
	for _, obj := range formalObjs {
		formalObjMap[obj.ObjectName] = obj
	}

	// 查询临时表对象和属性
	tempObjs, err := tempModel.FindByFormViewIdLatest(l.ctx, formViewId)
	if err != nil {
		return nil, errorx.Detail(errorx.QueryFailed, err, "业务对象列表")
	}

	allTempAttrs, err := tempAttrModel.FindByFormViewIdLatestWithFieldInfo(l.ctx, formViewId)
	if err != nil {
		return nil, errorx.Detail(errorx.QueryFailed, err, "业务对象属性列表")
	}

	// 按对象ID分组临时表属性
	tempAttrMap := make(map[string][]*business_object_attributes_temp.FieldWithAttrInfoTemp)
	for _, attr := range allTempAttrs {
		tempAttrMap[attr.BusinessObjectId] = append(tempAttrMap[attr.BusinessObjectId], attr)
	}

	objects := make([]types.BusinessObject, 0)

	// 处理临时表对象（与正式表融合）
	for _, obj := range tempObjs {
		attrsTemp := tempAttrMap[obj.Id]
		objFormal, exists := formalObjMap[obj.ObjectName]

		if !exists || objFormal == nil {
			// 新增对象，直接返回临时表数据
			objects = append(objects, types.BusinessObject{
				Id:         obj.Id,
				ObjectName: obj.ObjectName,
				Attributes: l.toAPIAttrs(attrsTemp),
			})
		} else {
			// 已存在对象，融合属性
			attrsFormal, err := formalAttrModel.FindByBusinessObjectIdWithFieldInfo(l.ctx, objFormal.Id)
			if err != nil {
				logx.WithContext(l.ctx).Infof("查询正式表属性失败: %v", err)
				objects = append(objects, types.BusinessObject{
					Id:         obj.Id,
					ObjectName: obj.ObjectName,
					Attributes: l.toAPIAttrs(attrsTemp),
				})
			} else {
				objects = append(objects, types.BusinessObject{
					Id:         objFormal.Id,
					ObjectName: objFormal.ObjectName,
					Attributes: l.toAPIAttrs(l.mergeAttrs(attrsFormal, attrsTemp)),
				})
			}
			delete(formalObjMap, obj.ObjectName)
		}
	}

	// 添加正式表中独有的对象
	for _, obj := range formalObjMap {
		attrsFormal, err := formalAttrModel.FindByBusinessObjectIdWithFieldInfo(l.ctx, obj.Id)
		if err != nil {
			continue
		}
		objects = append(objects, types.BusinessObject{
			Id:         obj.Id,
			ObjectName: obj.ObjectName,
			Attributes: l.toAPIAttrsFormal(attrsFormal),
		})
	}

	return objects, nil
}

// getBusinessObjectsFromFormal 从正式表查询业务对象
func (l *GetBusinessObjectsLogic) getBusinessObjectsFromFormal(req *types.GetBusinessObjectsReq) (*types.GetBusinessObjectsResp, error) {
	model := business_object.NewBusinessObjectModelSqlx(l.svcCtx.DB)
	attrModel := business_object_attributes.NewBusinessObjectAttributesModelSqlx(l.svcCtx.DB)

	var objects []types.BusinessObject

	if req.ObjectId != nil {
		// 查询单个对象
		obj, err := model.FindOneById(l.ctx, *req.ObjectId)
		if err != nil {
			return nil, errorx.Detail(errorx.QueryFailed, err, "业务对象")
		}
		attrs, err := attrModel.FindByBusinessObjectIdWithFieldInfo(l.ctx, *req.ObjectId)
		if err != nil {
			return nil, errorx.Detail(errorx.QueryFailed, err, "业务对象属性")
		}
		objects = []types.BusinessObject{{
			Id:         obj.Id,
			ObjectName: obj.ObjectName,
			Attributes: l.toAPIAttrsFormal(attrs),
		}}
	} else {
		// 查询所有对象
		objs, err := model.FindByFormViewId(l.ctx, req.Id)
		if err != nil {
			return nil, errorx.Detail(errorx.QueryFailed, err, "业务对象列表")
		}
		objects = make([]types.BusinessObject, 0, len(objs))
		for _, obj := range objs {
			attrs, err := attrModel.FindByBusinessObjectIdWithFieldInfo(l.ctx, obj.Id)
			if err != nil {
				continue
			}
			objects = append(objects, types.BusinessObject{
				Id:         obj.Id,
				ObjectName: obj.ObjectName,
				Attributes: l.toAPIAttrsFormal(attrs),
			})
		}
		if req.Keyword != nil && *req.Keyword != "" {
			objects = l.filterByKeyword(objects, *req.Keyword)
		}
	}

	// 查询未识别字段（从正式表查询）
	unidentifiedFields, err := l.getUnidentifiedFieldsFromFormal(req.Id, attrModel)
	if err != nil {
		logx.WithContext(l.ctx).Infof("查询未识别字段失败: %v", err)
		unidentifiedFields = []types.UnidentifiedField{} // 查询失败时返回空列表
	}

	return &types.GetBusinessObjectsResp{
		List:               objects,
		UnidentifiedFields: unidentifiedFields,
	}, nil
}

// toAPIAttrs 转换临时表属性到API格式
func (l *GetBusinessObjectsLogic) toAPIAttrs(attrs []*business_object_attributes_temp.FieldWithAttrInfoTemp) []types.BusinessObjectAttribute {
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
			Description:       attr.Description,
		})
	}
	return result
}

// toAPIAttrsFormal 转换正式表属性到API格式
func (l *GetBusinessObjectsLogic) toAPIAttrsFormal(attrs []*business_object_attributes.FieldWithAttrInfo) []types.BusinessObjectAttribute {
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
			Description:       attr.Description,
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

// mergeAttrs 融合正式表和临时表的属性
// 融合规则：按 attr_name + form_view_field_id 判断是否同一属性
// 1. 临时表新增的属性（正式表中不存在）→ 作为新增（attr_name 不为空）
// 2. 正式表和临时表都存在的属性 → 临时表 attr_name 不为空则使用临时表值，否则使用正式表值
func (l *GetBusinessObjectsLogic) mergeAttrs(attrsFormal []*business_object_attributes.FieldWithAttrInfo, attrsTemp []*business_object_attributes_temp.FieldWithAttrInfoTemp) []*business_object_attributes_temp.FieldWithAttrInfoTemp {
	formalMap := make(map[string]*business_object_attributes.FieldWithAttrInfo, len(attrsFormal))
	for _, attr := range attrsFormal {
		key := attr.FormViewFieldId + ":" + attr.AttrName
		formalMap[key] = attr
	}

	tempMap := make(map[string]*business_object_attributes_temp.FieldWithAttrInfoTemp, len(attrsTemp))
	for _, attr := range attrsTemp {
		key := attr.FormViewFieldId + ":" + attr.AttrName
		tempMap[key] = attr
	}

	result := make([]*business_object_attributes_temp.FieldWithAttrInfoTemp, 0, len(attrsFormal)+len(attrsTemp))

	// 处理正式表属性
	for _, attrFormal := range attrsFormal {
		key := attrFormal.FormViewFieldId + ":" + attrFormal.AttrName
		attrTemp, exists := tempMap[key]
		if !exists || attrTemp.AttrName == "" {
			// 使用正式表数据
			result = append(result, &business_object_attributes_temp.FieldWithAttrInfoTemp{
				Id:                attrFormal.Id,
				BusinessObjectId:  "",
				AttrName:          attrFormal.AttrName,
				FormViewFieldId:   attrFormal.FormViewFieldId,
				FieldTechName:     attrFormal.FieldTechName,
				FieldBusinessName: attrFormal.FieldBusinessName,
				FieldRole:         attrFormal.FieldRole,
				FieldType:         attrFormal.FieldType,
				Description:       attrFormal.Description,
			})
		} else {
			// 使用临时表数据
			result = append(result, attrTemp)
		}
	}

	// 添加临时表新增属性（attr_name 不为空）
	for _, attr := range attrsTemp {
		key := attr.FormViewFieldId + ":" + attr.AttrName
		if _, exists := formalMap[key]; !exists {
			if attr.AttrName != "" {
				result = append(result, attr)
			}
		}
	}

	return result
}

// getUnidentifiedFields 查询未识别字段（attr_name 和 business_object_id 为空的字段）
func (l *GetBusinessObjectsLogic) getUnidentifiedFields(formViewId string, tempAttrModel *business_object_attributes_temp.BusinessObjectAttributesTempModelSqlx) ([]types.UnidentifiedField, error) {
	fields, err := tempAttrModel.FindUnidentifiedFieldsLatest(l.ctx, formViewId)
	if err != nil {
		return nil, err
	}

	result := make([]types.UnidentifiedField, 0, len(fields))
	for _, field := range fields {
		result = append(result, types.UnidentifiedField{
			Id:            field.Id,
			TechnicalName: field.TechnicalName,
			DataType:      field.DataType,
			BusinessName:  field.BusinessName,
			FieldRole:     field.FieldRole,
			Description:   field.Description,
		})
	}
	return result, nil
}

// getUnidentifiedFieldsFromFormal 从正式表查询未识别字段（business_object_id 和 attr_name 都为空的字段）
func (l *GetBusinessObjectsLogic) getUnidentifiedFieldsFromFormal(formViewId string, formalAttrModel *business_object_attributes.BusinessObjectAttributesModelSqlx) ([]types.UnidentifiedField, error) {
	fields, err := formalAttrModel.FindUnrecognizedFields(l.ctx, formViewId)
	if err != nil {
		return nil, err
	}

	result := make([]types.UnidentifiedField, 0, len(fields))
	for _, field := range fields {
		result = append(result, types.UnidentifiedField{
			Id:            field.Id,
			TechnicalName: field.FieldTechName,
			DataType:      field.FieldType,
			BusinessName:  field.FieldBusinessName,
			FieldRole:     field.FieldRole,
			Description:   field.Description,
		})
	}
	return result, nil
}
