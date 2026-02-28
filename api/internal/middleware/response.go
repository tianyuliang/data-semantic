package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/kweaver-ai/idrm-go-frame/core/errorx/agcodes"
	"github.com/kweaver-ai/idrm-go-frame/core/errorx/agerrors"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// HttpError 错误响应
type HttpError struct {
	Code        string `json:"code"`
	Description string `json:"description"`
	Solution    string `json:"solution,omitempty"`
	Cause       string `json:"cause,omitempty"`
	Detail      any    `json:"detail,omitempty"`
}

type Response struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

type responseWrapper struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func (w *responseWrapper) WriteHeader(code int) {
	w.statusCode = code
}

func (w *responseWrapper) Write(data []byte) (int, error) {
	return w.body.Write(data)
}

// isHttpError 检测是否为 HttpError 结构
func isHttpError(data map[string]any) bool {
	_, hasCode := data["code"]
	_, hasDesc := data["description"]
	return hasCode && hasDesc
}

// respondRaw 直接写入原始响应
func respondRaw(w http.ResponseWriter, data []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

// ResponseWrapper 统一响应格式中间件
func ResponseWrapper() func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			wrapper := &responseWrapper{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
				body:           bytes.NewBuffer(nil),
			}

			next(wrapper, r)

			body := wrapper.body.Bytes()
			if len(body) == 0 {
				return
			}

			var data map[string]any
			if json.Unmarshal(body, &data) != nil {
				respondRaw(w, body)
				return
			}

			// 错误响应不包装，直接返回 HttpError
			if isHttpError(data) {
				respondRaw(w, body)
				return
			}

			// 成功响应包装
			ResOKJson(w, Response{Code: "", Message: "success", Data: data})
		}
	}
}

// ResOKJson 成功响应
func ResOKJson(w http.ResponseWriter, data any) {
	if data == nil {
		data = map[string]any{}
	}
	httpx.WriteJson(w, http.StatusOK, data)
}

// ResErrJson 通用错误响应
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

// AbortResponseWithCode 中止请求（自定义状态码）
func AbortResponseWithCode(w http.ResponseWriter, code int, err error) {
	w.WriteHeader(code)
	AbortResponse(w, err)
}

// AbortResponse 中止请求
func AbortResponse(w http.ResponseWriter, err error) {
	var code = agerrors.Code(err)
	if err == nil {
		code = agcodes.CodeNotAuthorized
	}
	// 写入错误响应（保持已设置的状态码）
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(HttpError{
		Code:        code.GetErrorCode(),
		Description: code.GetDescription(),
		Solution:    code.GetSolution(),
		Cause:       code.GetCause(),
		Detail:      code.GetErrorDetails(),
	})
}
