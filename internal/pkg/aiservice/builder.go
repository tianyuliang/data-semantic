package aiservice

import (
	"fmt"

	"github.com/kweaver-ai/dsg/services/apps/data-semantic/model/form_view"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/model/form_view_field"
)

// BuildFormView builds FormView for AI service from model data
func BuildFormView(formViewID string, formViewData *form_view.FormView, fields []*form_view_field.FormViewField) *FormView {
	aiFields := make([]FormViewField, 0, len(fields))
	for _, f := range fields {
		aiFields = append(aiFields, FormViewFieldFromModel(f))
	}

	businessName := ""
	if formViewData.BusinessName != "" {
		businessName = formViewData.BusinessName
	}

	description := ""
	if formViewData.Description != "" {
		description = formViewData.Description
	}

	return &FormView{
		ID:            formViewID,
		TechnicalName: formViewData.TechnicalName,
		BusinessName:  businessName,
		Description:   description,
		Fields:        aiFields,
	}
}

// FormViewFieldFromModel converts model FormViewField to aiservice FormViewField
func FormViewFieldFromModel(f *form_view_field.FormViewField) FormViewField {
	fieldRole := ""
	if f.FieldRole != nil {
		fieldRole = fmt.Sprintf("%d", *f.FieldRole)
	}

	fieldDesc := ""
	if f.FieldDescription != nil {
		fieldDesc = *f.FieldDescription
	}

	fieldBusinessName := ""
	if f.FieldBusinessName != nil {
		fieldBusinessName = *f.FieldBusinessName
	}

	return FormViewField{
		ID:            f.Id,
		TechnicalName: f.FieldTechName,
		BusinessName:  fieldBusinessName,
		Type:          f.FieldType,
		Role:          fieldRole,
		Description:   fieldDesc,
	}
}
