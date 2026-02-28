package middleware

import (
	"net/http"

	"github.com/kweaver-ai/idrm-go-frame/core/errorx/agcodes"
	"github.com/kweaver-ai/idrm-go-frame/core/errorx/agerrors"
	"github.com/zeromicro/go-zero/rest/httpx"
)

type HttpError struct {
	Code        string      `json:"code"`
	Description string      `json:"description"`
	Solution    string      `json:"solution,omitempty"`
	Cause       string      `json:"cause,omitempty"`
	Detail      interface{} `json:"detail,omitempty"`
	Data        interface{} `json:"data,omitempty"`
}

// ResOKJson success Json Response
func ResOKJson(w http.ResponseWriter, data interface{}) {
	if data == nil {
		data = map[string]interface{}{}
	}
	httpx.WriteJson(w, http.StatusOK, data)
}

// ResList list Response
func ResList(w http.ResponseWriter, list interface{}, totalCount int) {
	httpx.WriteJson(w, http.StatusOK, map[string]interface{}{
		"entries":     list,
		"total_count": totalCount,
	})
}

// ResBadRequestJson bad request Json Response
func ResBadRequestJson(w http.ResponseWriter, err error) {
	ResErrJsonWithCode(w, http.StatusBadRequest, err)
}

// ResErrJsonWithCode failed Json Response with custom status code
func ResErrJsonWithCode(w http.ResponseWriter, code int, err error) {
	w.WriteHeader(code)
	ResErrJson(w, err)
}

// ResErrJson failed Json Response
func ResErrJson(w http.ResponseWriter, err error) {
	var code agcodes.Coder
	if err == nil {
		code = agcodes.CodeOK
	} else {
		code = agerrors.Code(err)
	}

	httpx.WriteJson(w, http.StatusOK, HttpError{
		Code:        code.GetErrorCode(),
		Description: code.GetDescription(),
		Solution:    code.GetSolution(),
		Cause:       code.GetCause(),
		Detail:      code.GetErrorDetails(),
	})
}

// AbortResponseWithCode abort request with custom status code and error
func AbortResponseWithCode(w http.ResponseWriter, code int, err error) {
	w.WriteHeader(code)
	AbortResponse(w, err)
}

// AbortResponse abort request with error
func AbortResponse(w http.ResponseWriter, err error) {
	var code = agerrors.Code(err)
	if err == nil {
		code = agcodes.CodeNotAuthorized
	}
	httpx.WriteJson(w, http.StatusOK, HttpError{
		Code:        code.GetErrorCode(),
		Description: code.GetDescription(),
		Solution:    code.GetSolution(),
		Cause:       code.GetCause(),
		Detail:      code.GetErrorDetails(),
	})
}
