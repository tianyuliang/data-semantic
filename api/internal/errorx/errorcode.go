package errorx

import (
	"fmt"
	"regexp"

	"github.com/kweaver-ai/idrm-go-frame/core/errorx/agcodes"
	"github.com/kweaver-ai/idrm-go-frame/core/errorx/agerrors"
)

// Model Name
const (
	publicModelName = "Public"
)

var Success = map[string]string{
	"code":        "0",
	"description": "成功",
	"solution":    "",
}

// Public error
const (
	publicPreCoder = publicModelName + "."

	PublicInternalError         = publicPreCoder + "InternalError"
	PublicInvalidParameter      = publicPreCoder + "InvalidParameter"
	PublicInvalidParameterJson  = publicPreCoder + "InvalidParameterJson"
	PublicInvalidParameterValue = publicPreCoder + "InvalidParameterValue"
	PublicDatabaseError         = publicPreCoder + "DatabaseError"
	PublicRequestParameterError = publicPreCoder + "RequestParameterError"
	PublicUniqueIDError         = publicPreCoder + "PublicUniqueIDError"

	NotAuthentication                     = publicPreCoder + "NotAuthentication"
	HydraException                        = publicPreCoder + "HydraException"
	AuthServiceException                  = publicPreCoder + "AuthServiceException"
	AuthenticationFailure                 = publicPreCoder + "AuthenticationFailure"
	GetUserInfoFailure                    = publicPreCoder + "GetUserInfoFailure"
	GetAppInfoFailure                     = publicPreCoder + "GetAppInfoFailure"
	GetProtonAppInfoFailure               = publicPreCoder + "GetProtonAppInfoFailure"
	AuthorizationFailure                  = publicPreCoder + "AuthorizationFailure"
	PermissionCheckFailure                = publicPreCoder + "PermissionCheckFailure"
	AccessTypeNotSupport                  = publicPreCoder + "AccessTypeNotSupport"
	AccessControlClientTokenMustHasUserId = publicPreCoder + "AccessControlClientTokenMustHasUserId"

	ContextNotHaveToken    = publicPreCoder + "ContextNotHaveToken"
	ContextNotHaveUserInfo = publicPreCoder + "ContextNotHaveUserInfo"
	CallAfSailorError      = publicPreCoder + "CallAfSailorError"

	PublicResourceNotFoundError = publicPreCoder + "ResourceNotFoundError"
)

var publicErrorMap = ErrorCode{
	PublicInternalError: {
		Description: "内部错误",
		Cause:       "",
		Solution:    "",
	},
	PublicInvalidParameter: {
		Description: "参数值校验不通过",
		Cause:       "",
		Solution:    "请使用请求参数构造规范化的请求字符串。详细信息参见产品 API 文档",
	},
	PublicInvalidParameterJson: {
		Description: "参数值校验不通过：json格式错误",
		Solution:    "请使用请求参数构造规范化的请求字符串，详细信息参见产品 API 文档",
	},
	PublicInvalidParameterValue: {
		Description: "参数值[param]校验不通过",
		Cause:       "",
		Solution:    "请使用请求参数构造规范化的请求字符串。详细信息参见产品 API 文档",
	},
	PublicDatabaseError: {
		Description: "数据库异常",
		Cause:       "",
		Solution:    "请检查数据库状态",
	},
	PublicRequestParameterError: {
		Description: "请求参数格式错误",
		Cause:       "输入请求参数格式或内容有问题",
		Solution:    "请输入正确格式的请求参数",
	},
	PublicUniqueIDError: {
		Description: "ID生成失败",
		Cause:       "",
		Solution:    "",
	},
	NotAuthentication: {
		Description: "无用户登录信息",
		Cause:       "",
		Solution:    "",
	},
	HydraException: {
		Description: "授权服务异常",
		Cause:       "",
		Solution:    "",
	},
	AuthServiceException: {
		Description: "授权服务异常",
	},
	AuthenticationFailure: {
		Description: "用户登录已过期",
		Cause:       "",
		Solution:    "",
	},
	GetUserInfoFailure: {
		Description: "获取用户信息失败",
		Cause:       "",
		Solution:    "",
	},
	GetAppInfoFailure: {
		Description: "获取应用信息失败",
		Cause:       "",
		Solution:    "",
	},
	GetProtonAppInfoFailure: {
		Description: "获取部署控制台应用信息失败",
		Cause:       "",
		Solution:    "",
	},
	AuthorizationFailure: {
		Description: "暂无权限，您可联系系统管理员配置",
		Cause:       "",
		Solution:    "",
	},
	PermissionCheckFailure: {
		Description: "暂无[permission_name]权限，您可联系管理员配置",
	},
	AccessTypeNotSupport: {
		Description: "暂不支持的访问类型",
		Cause:       "",
		Solution:    "",
	},
	AccessControlClientTokenMustHasUserId: {
		Description: "客户端token必须携带userId",
		Cause:       "",
		Solution:    "请重试",
	},
	ContextNotHaveToken: {
		Description: "上下文中没有令牌",
		Cause:       "",
		Solution:    "请重试",
	},
	ContextNotHaveUserInfo: {
		Description: "上下文中没有用户信息",
		Cause:       "",
		Solution:    "请重试",
	},
	CallAfSailorError: {
		Description: "请求认知助手服务失败",
		Solution:    "请检查配置及服务",
	},
	PublicResourceNotFoundError: {
		Description: "资源不存在",
		Solution:    "请检查参数和数据",
	},
}

type ErrorCodeFullInfo struct {
	Code        string      `json:"code"`
	Description string      `json:"description"`
	Cause       string      `json:"cause"`
	Solution    string      `json:"solution"`
	Detail      interface{} `json:"detail,omitempty"`
}

type ErrorCodeInfo struct {
	Description string
	Cause       string
	Solution    string
}

type ErrorCode map[string]ErrorCodeInfo

var errorCodeMap ErrorCode

func IsErrorCode(err error) bool {
	_, ok := err.(*agerrors.Error)
	return ok
}

func RegisterErrorCode(errCodes ...ErrorCode) {
	if errorCodeMap == nil {
		// errorCodeMap init
		errorCodeMap = ErrorCode{}
	}

	for _, m := range errCodes {
		for k := range m {
			if _, ok := errorCodeMap[k]; ok {
				// error code is not allowed to repeat
				panic(fmt.Sprintf("error code is not allowed to repeat, code: %s", k))
			}

			errorCodeMap[k] = m[k]
		}
	}
}

func init() {
	RegisterErrorCode(publicErrorMap)
}

func Desc(errCode string, args ...any) error {
	return newCoder(errCode, nil, args...)
}
func WithDetail(errCode string, detail map[string]any) error {
	return newCoder(errCode, detail)
}
func Detail(errCode string, err any, args ...any) error {
	return newCoder(errCode, err, args...)
}

func New(errorCode, description, cause, solution string, detail interface{}, errLink string) error {
	coder := agcodes.New(errorCode, description, cause, solution, detail, errLink)
	return agerrors.NewCode(coder)
}

// ManualNew 手动new一个错误
func ManualNew(errCode string, err ErrorCodeInfo) error {
	coder := agcodes.New(errCode, err.Description, err.Cause, err.Solution, err, "")
	return agerrors.NewCode(coder)
}

func newCoder(errCode string, err any, args ...any) error {
	errInfo, ok := errorCodeMap[errCode]
	if !ok {
		errInfo = errorCodeMap[PublicInternalError]
		errCode = PublicInternalError
	}

	desc := errInfo.Description
	if len(args) > 0 {
		desc = FormatDescription(desc, args...)
	}
	if err == nil {
		err = struct{}{}
	}

	coder := agcodes.New(errCode, desc, errInfo.Cause, errInfo.Solution, err, "")
	return agerrors.NewCode(coder)
}

// FormatDescription replace the placeholder in coder.Description
// Example:
// Description: call service [service_name] api [api_name] error,
// args:  configuration-center, create
// =>
// Description: call service [configuration-center] api [create] error,
func FormatDescription(s string, args ...interface{}) string {
	if len(args) <= 0 {
		return s
	}
	re, _ := regexp.Compile("\\[\\w+\\]")
	result := re.ReplaceAll([]byte(s), []byte("[%v]"))
	return fmt.Sprintf(string(result), args...)
}
