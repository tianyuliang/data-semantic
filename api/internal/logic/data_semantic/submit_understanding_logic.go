// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package data_semantic

import (
	"context"

	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/errorx"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/svc"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/types"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/model/data_understanding/business_object"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/model/data_understanding/business_object_attributes"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/model/data_understanding/business_object_attributes_temp"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/model/data_understanding/business_object_temp"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/model/data_understanding/form_view_field_info_temp"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/model/data_understanding/form_view_info_temp"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/model/form_view"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/model/form_view_field"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
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
	logx.Infof("SubmitUnderstanding called with id: %s", req.Id)

	// 1. 状态校验 (仅允许状态 2 提交)
	formViewModel := form_view.NewFormViewModel(l.svcCtx.DB)
	formViewData, err := formViewModel.FindOneById(l.ctx, req.Id)
	if err != nil {
		return nil, errorx.NewQueryFailed("库表视图", err)
	}

	if formViewData.UnderstandStatus != form_view.StatusPendingConfirm {
		return nil, errorx.NewInvalidUnderstandStatus(formViewData.UnderstandStatus)
	}

	// 2. 获取当前版本号
	businessObjectTempModel := business_object_temp.NewBusinessObjectTempModelSqlx(l.svcCtx.DB)
	latestVersion, err := businessObjectTempModel.FindLatestVersionByFormViewId(l.ctx, req.Id)
	if err != nil {
		return nil, errorx.NewQueryFailed("当前版本号", err)
	}
	if latestVersion == 0 {
		return nil, errorx.Newf(errorx.ErrCodeInvalidParam, "没有可提交的数据，版本号为0")
	}

	// 3. 开启事务处理
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 使用事务的 Session 创建 model 实例
		tempModel := business_object_temp.NewBusinessObjectTempModelSession(session)
		tempAttrModel := business_object_attributes_temp.NewBusinessObjectAttributesTempModelSession(session)
		tempFormViewInfoModel := form_view_info_temp.NewFormViewInfoTempModelSession(session)
		tempFormViewFieldInfoModel := form_view_field_info_temp.NewFormViewFieldInfoTempModelSession(session)
		formalModel := business_object.NewBusinessObjectModelSession(session)
		formalAttrModel := business_object_attributes.NewBusinessObjectAttributesModelSession(session)
		formViewModelSession := form_view.NewFormViewModelSession(session)
		formViewFieldModel := form_view_field.NewFormViewFieldModelSession(session)

		// ========== 合并业务对象（基于业务主键：form_view_id + object_name）==========
		objInserted, objUpdated, objDeleted, err := l.mergeBusinessObjects(ctx, req.Id, latestVersion, tempModel, formalModel)
		if err != nil {
			return errorx.NewUpdateFailed("业务对象", err)
		}
		logx.WithContext(ctx).Infof("Merged business objects: inserted=%d, updated=%d, deleted=%d", objInserted, objUpdated, objDeleted)

		// ========== 合并业务对象属性（基于业务对象匹配 + attr_name + form_view_field_id）==========
		attrInserted, attrUpdated, attrDeleted, err := l.mergeBusinessObjectAttributes(ctx, req.Id, latestVersion, tempModel, tempAttrModel, formalModel, formalAttrModel)
		if err != nil {
			return errorx.NewUpdateFailed("属性", err)
		}
		logx.WithContext(ctx).Infof("Merged attributes: inserted=%d, updated=%d, deleted=%d", attrInserted, attrUpdated, attrDeleted)

		// ========== 更新库表业务名称和描述 ==========
		if err := l.updateFormViewInfo(ctx, req.Id, latestVersion, tempFormViewInfoModel, formViewModelSession); err != nil {
			return errorx.NewUpdateFailed("库表信息", err)
		}
		logx.WithContext(ctx).Infof("Updated form view info")

		// ========== 更新字段业务名称、角色和描述 ==========
		fieldUpdated, err := l.updateFormViewFieldInfo(ctx, req.Id, latestVersion, tempFormViewFieldInfoModel, formViewFieldModel)
		if err != nil {
			return errorx.NewUpdateFailed("字段信息", err)
		}
		logx.WithContext(ctx).Infof("Updated form view field info: updated=%d", fieldUpdated)

		// ========== 更新 form_view 状态为 3 (已完成) ==========
		err = formViewModelSession.UpdateUnderstandStatus(ctx, req.Id, form_view.StatusCompleted)
		if err != nil {
			return errorx.NewUpdateFailed("理解状态", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	logx.WithContext(l.ctx).Infof("Submit understanding successful: form_view_id=%s, version=%d", req.Id, latestVersion)

	resp = &types.SubmitUnderstandingResp{
		Success: true,
	}

	return resp, nil
}

// mergeBusinessObjects 合并业务对象（代码层面实现）
// 按 object_name 匹配：存在则跳过（无需更新），不存在则新增，正式表独有的删除
func (l *SubmitUnderstandingLogic) mergeBusinessObjects(
	ctx context.Context,
	formViewId string,
	version int,
	tempModel *business_object_temp.BusinessObjectTempModelSqlx,
	formalModel *business_object.BusinessObjectModelSqlx,
) (inserted, updated, deleted int, err error) {
	// 1. 查询临时表数据
	tempObjs, err := tempModel.FindByFormViewAndVersion(ctx, formViewId, version)
	if err != nil {
		return 0, 0, 0, err
	}

	// 2. 查询正式表数据
	formalObjs, err := formalModel.FindByFormViewId(ctx, formViewId)
	if err != nil {
		return 0, 0, 0, err
	}

	// 3. 构建正式表对象映射（key: object_name）
	formalObjMap := make(map[string]*business_object.BusinessObject)
	for _, obj := range formalObjs {
		formalObjMap[obj.ObjectName] = obj
	}

	// 4. 处理临时表对象
	for _, obj := range tempObjs {
		objFormal, exists := formalObjMap[obj.ObjectName]
		if !exists || objFormal == nil {
			// 新增：使用临时表的 id
			newObj := &business_object.BusinessObject{
				Id:         obj.Id,
				ObjectName: obj.ObjectName,
				FormViewId: obj.FormViewId,
			}
			if _, err := formalModel.Insert(ctx, newObj); err != nil {
				return 0, 0, 0, err
			}
			inserted++
		} else {
			// 已存在，无需更新（通过 object_name 匹配，名字已相同）
			delete(formalObjMap, obj.ObjectName)
		}
	}

	// 5. 删除正式表中独有的对象
	for _, obj := range formalObjMap {
		if err := formalModel.Delete(ctx, obj.Id); err != nil {
			return 0, 0, 0, err
		}
		deleted++
	}

	return inserted, updated, deleted, nil
}

// mergeBusinessObjectAttributes 合并业务对象属性（代码层面实现）
// 按 business_object_id + attr_name + form_view_field_id 匹配：存在则跳过（无需更新），不存在则新增，正式表独有的删除
func (l *SubmitUnderstandingLogic) mergeBusinessObjectAttributes(
	ctx context.Context,
	formViewId string,
	version int,
	tempModel *business_object_temp.BusinessObjectTempModelSqlx,
	tempAttrModel *business_object_attributes_temp.BusinessObjectAttributesTempModelSqlx,
	formalModel *business_object.BusinessObjectModelSqlx,
	formalAttrModel *business_object_attributes.BusinessObjectAttributesModelSqlx,
) (inserted, updated, deleted int, err error) {
	// 1. 查询临时表对象
	tempObjs, err := tempModel.FindByFormViewAndVersion(ctx, formViewId, version)
	if err != nil {
		return 0, 0, 0, err
	}

	// 构建正式表对象 ID 集合（用于判断哪些对象还存在）
	formalObjIds := make(map[string]bool)
	formalObjIdByName := make(map[string]string) // object_name -> obj_id

	// 先查询所有正式表对象
	allFormalObjs, err := formalModel.FindByFormViewId(ctx, formViewId)
	if err != nil {
		return 0, 0, 0, err
	}
	for _, obj := range allFormalObjs {
		formalObjIds[obj.Id] = true
		formalObjIdByName[obj.ObjectName] = obj.Id
	}

	// 2. 构建临时表对象 ID -> 正式表对象 ID 映射
	tempObjIdToFormalObjId := make(map[string]string)
	for _, obj := range tempObjs {
		if formalId, exists := formalObjIdByName[obj.ObjectName]; exists {
			tempObjIdToFormalObjId[obj.Id] = formalId
		}
	}

	// 3. 查询临时表属性
	tempAttrs, err := tempAttrModel.FindByFormViewAndVersion(ctx, formViewId, version)
	if err != nil {
		return 0, 0, 0, err
	}

	// 4. 查询正式表所有属性
	formalAttrs, err := formalAttrModel.FindByFormViewId(ctx, formViewId)
	if err != nil {
		return 0, 0, 0, err
	}

	// 5. 构建正式表属性映射（key: business_object_id + attr_name + form_view_field_id）
	formalAttrMap := make(map[string]*business_object_attributes.BusinessObjectAttributes)
	for _, attr := range formalAttrs {
		// 只保留属于有效业务对象的属性
		if formalObjIds[attr.BusinessObjectId] {
			key := attr.BusinessObjectId + ":" + attr.AttrName + ":" + attr.FormViewFieldId
			formalAttrMap[key] = attr
		} else {
			// 属于已删除业务对象的属性，直接删除
			if err := formalAttrModel.Delete(ctx, attr.Id); err != nil {
				return 0, 0, 0, err
			}
			deleted++
		}
	}

	// 6. 处理临时表属性
	for _, attr := range tempAttrs {
		// 获取正式表的 business_object_id
		formalObjId, ok := tempObjIdToFormalObjId[attr.BusinessObjectId]
		if !ok {
			continue // 跳过无法找到对应正式表对象的属性
		}

		key := formalObjId + ":" + attr.AttrName + ":" + attr.FormViewFieldId
		attrFormal, exists := formalAttrMap[key]
		if !exists || attrFormal == nil {
			// 新增：attr_name 不为空
			if attr.AttrName != "" {
				newAttr := &business_object_attributes.BusinessObjectAttributes{
					Id:               attr.Id,
					FormViewId:       attr.FormViewId,
					BusinessObjectId: formalObjId,
					FormViewFieldId:  attr.FormViewFieldId,
					AttrName:         attr.AttrName,
				}
				if _, err := formalAttrModel.Insert(ctx, newAttr); err != nil {
					return 0, 0, 0, err
				}
				inserted++
			}
		} else {
			// 已存在，无需更新（通过 attr_name + form_view_field_id 匹配，属性名已相同）
			delete(formalAttrMap, key)
		}
	}

	// 7. 删除正式表中独有的属性（属于有效业务对象的）
	for _, attr := range formalAttrMap {
		if err := formalAttrModel.Delete(ctx, attr.Id); err != nil {
			return 0, 0, 0, err
		}
		deleted++
	}

	return inserted, updated, deleted, nil
}

// updateFormViewInfo 更新库表业务名称和描述
func (l *SubmitUnderstandingLogic) updateFormViewInfo(
	ctx context.Context,
	formViewId string,
	version int,
	tempModel *form_view_info_temp.FormViewInfoTempModelSqlx,
	formViewModel *form_view.FormViewModelSqlx,
) error {
	// 查询临时表库表信息
	tempInfo, err := tempModel.FindOneByFormViewAndVersion(ctx, formViewId, version)
	if err != nil {
		// 如果没有找到临时数据，跳过（可能用户没有修改库表信息）
		return nil
	}

	// 更新正式表的库表信息
	return formViewModel.UpdateBusinessInfo(ctx, formViewId, tempInfo.TableBusinessName, tempInfo.TableDescription)
}

// updateFormViewFieldInfo 更新字段业务名称、角色和描述
func (l *SubmitUnderstandingLogic) updateFormViewFieldInfo(
	ctx context.Context,
	formViewId string,
	version int,
	tempModel *form_view_field_info_temp.FormViewFieldInfoTempModelSqlx,
	formViewFieldModel *form_view_field.FormViewFieldModelSqlx,
) (updated int, err error) {
	// 查询临时表字段信息
	tempFields, err := tempModel.FindByFormViewAndVersion(ctx, formViewId, version)
	if err != nil {
		// 如果没有找到临时数据，跳过
		return 0, nil
	}

	// 更新每个字段的信息
	for _, field := range tempFields {
		err := formViewFieldModel.UpdateBusinessInfo(ctx, field.FormViewFieldId, field.FieldBusinessName, field.FieldRole, field.FieldDescription)
		if err != nil {
			return 0, err
		}
		updated++
	}

	return updated, nil
}
