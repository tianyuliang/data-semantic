package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/kweaver-ai/dsg/services/apps/data-semantic/internal/pkg/hydra"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/internal/pkg/usermgm"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"

	errcode "github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/errorx"
)

// UserInfo 用户信息
type UserInfo struct {
	ID        string             // 用户 ID
	Name      string             // 用户名称
	OrgInfos  []*usermgm.DepInfo // 组织信息
	UserType  int                // 用户类型 (TokenTypeUser 或 TokenTypeClient)
}

// context key 常量（与 idrm-go-common/interception 兼容）
const (
	InfoName = "info"                              // 用户信息 context key
	Token    = "token"                             // Bearer token context key
)

// BearerToken context key
const contextKeyBearerToken = "GoCommon/interception.BearerToken"

// TokenType Token 类型
const TokenType = "token_type"

// TokenType 常量
const (
	TokenTypeUser   = iota // 用户 Token
	TokenTypeClient        // 应用 Token
)

// NewContextWithBearerToken 生成一个包含 BearerToken 的 context.Context
func NewContextWithBearerToken(parent context.Context, t string) context.Context {
	return context.WithValue(parent, contextKeyBearerToken, t)
}

// JWTAuth JWT 认证中间件
func JWTAuth(hydraClient *hydra.Client, userMgmClient *usermgm.Client) rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			var err error
			newCtx := r.Context()

			// 提取 token
			tokenID := r.Header.Get("Authorization")
			token := strings.TrimPrefix(tokenID, "Bearer ")
			if tokenID == "" || token == "" {
				AbortResponseWithCode(w, http.StatusUnauthorized, errcode.Desc(errcode.NotAuthentication))
				return
			}

			// 调用 Hydra introspect
			info, err := hydraClient.Introspect(newCtx, token)
			if err != nil {
				logx.WithContext(r.Context()).Error("Token Introspect", "error", err)
				AbortResponseWithCode(w, http.StatusBadRequest, errcode.Desc(errcode.HydraException))
				return
			}
			if !info.Active {
				AbortResponseWithCode(w, http.StatusUnauthorized, errcode.Desc(errcode.AuthenticationFailure))
				return
			}

			// 保存 Bearer token 用于身份认证
			newCtx = NewContextWithBearerToken(newCtx, token)

			// 处理用户 Token
			newCtx, err = handleUserToken(newCtx, userMgmClient, &info, tokenID)
			if err != nil {
				logx.WithContext(r.Context()).Error("handleUserToken", "error", err)
				AbortResponseWithCode(w, http.StatusBadRequest, err)
				return
			}

			next(w, r.WithContext(newCtx))
		}
	}
}

// handleUserToken 处理用户 Token
func handleUserToken(ctx context.Context, userMgm *usermgm.Client, info *hydra.TokenIntrospectInfo, tokenID string) (context.Context, error) {
	name, _, depInfos, err := userMgm.GetUserNameByUserID(ctx, info.VisitorID)
	if err != nil {
		return ctx, errcode.Desc(errcode.GetUserInfoFailure)
	}

	userInfo := &UserInfo{
		ID:       info.VisitorID,
		Name:     name,
		OrgInfos: depInfos,
		UserType: TokenTypeUser,
	}

	ctx = context.WithValue(ctx, InfoName, userInfo)
	ctx = context.WithValue(ctx, Token, tokenID)
	ctx = context.WithValue(ctx, TokenType, TokenTypeUser)

	return ctx, nil
}
