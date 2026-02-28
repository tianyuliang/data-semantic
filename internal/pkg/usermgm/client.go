package usermgm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// DepInfo 部门信息
type DepInfo struct {
	OrgCode string `json:"org_code"`
	OrgName string `json:"org_name"`
}

// Client 用户管理服务客户端
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient 创建用户管理服务客户端
func NewClient(url string, timeout time.Duration) *Client {
	if url == "" {
		return nil
	}
	return &Client{
		baseURL: url,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// GetUserNameByUserID 通过用户ID获取用户名
func (c *Client) GetUserNameByUserID(ctx context.Context, userID string) (name string, isNormalUser bool, depInfos []*DepInfo, err error) {
	if c == nil {
		// 未配置 UserManagement 服务时返回默认值
		return userID, true, nil, nil
	}

	url := fmt.Sprintf("%s/api/v1/users/%s/info", c.baseURL, userID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", false, nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", false, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", false, nil, fmt.Errorf("UserManagement service returned status %d", resp.StatusCode)
	}

	var result struct {
		Name       string     `json:"name"`
		NormalUser bool       `json:"is_normal_user"`
		DepInfos   []*DepInfo `json:"dep_infos"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", false, nil, err
	}

	return result.Name, result.NormalUser, result.DepInfos, nil
}
