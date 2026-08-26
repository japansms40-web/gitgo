package github

import (
	"context"
	"encoding/json"
	"fmt"
)

// CreateRepoParams 是「新建仓库」的入参。仅暴露常用字段，其余固定字段复刻抓包默认值。
type CreateRepoParams struct {
	// Owner 仓库属主登录名（对应 body.owner）。
	Owner string
	// Name 仓库名。
	Name string
	// Visibility 可见性，"public" 或 "private"；空则按 "public"。
	Visibility string
	// Description 仓库描述；可空。
	Description string
}

// createRepoBody 是 POST /repositories 的 JSON body，结构对着 scratch 抓包。
type createRepoBody struct {
	Owner                string         `json:"owner"`
	TemplateRepositoryID string         `json:"template_repository_id"`
	IncludeAllBranches   string         `json:"include_all_branches"`
	Repository           createRepoRepo `json:"repository"`
	Metrics              map[string]any `json:"metrics"`
}

type createRepoRepo struct {
	Name              string `json:"name"`
	Visibility        string `json:"visibility"`
	Description       string `json:"description"`
	AutoInit          string `json:"auto_init"`
	LicenseTemplate   string `json:"license_template"`
	GitignoreTemplate string `json:"gitignore_template"`
}

// CreateRepoResponse 是建仓库接口的响应（由 TestCreateRepo_Live 实测抓到）。
type CreateRepoResponse struct {
	Data struct {
		// Redirect 建成后跳转的仓库路径，如 "/owner/name"。
		Redirect string `json:"redirect"`
	} `json:"data"`
}

// buildCreateRepoBody 拼出建仓库的 JSON body。metrics 用抓包默认值填死（纯遥测，服务端不据此判死）。
func buildCreateRepoBody(p CreateRepoParams) ([]byte, error) {
	visibility := p.Visibility
	if visibility == "" {
		visibility = "public"
	}
	body := createRepoBody{
		Owner:                p.Owner,
		TemplateRepositoryID: "",
		IncludeAllBranches:   "0",
		Repository: createRepoRepo{
			Name:              p.Name,
			Visibility:        visibility,
			Description:       p.Description,
			AutoInit:          "0",
			LicenseTemplate:   "",
			GitignoreTemplate: "",
		},
		Metrics: map[string]any{
			"user_filtered_dropdown":                    false,
			"user_set_template":                         false,
			"user_changed_default_owner":                false,
			"user_changed_owner_after_setting_template": false,
			"created_from_organization":                 false,
			"prepopulated_template":                     false,
			"owner_has_marketplace_apps":                false,
			"user_interacted_with_marketplace_apps":     false,
			"user_is_admin":                             false,
			"submit_clicked_count":                      1,
			"submit_errors":                             []string{},
			"clicked_suggested_repo_name":               false,
			"used_suggested_repo_name":                  false,
			"submitted_using_v2":                        true,
		},
	}
	return json.Marshal(body)
}

// CreateRepo 新建一个仓库：POST /repositories（JSON body，Cookie 会话鉴权）。
// 复用 Client 现有最小请求头（含 content-type: application/json）。
//
// 非 2xx（如重名、Cookie 过期）返回带状态码的 error。
func (c *Client) CreateRepo(ctx context.Context, p CreateRepoParams) (*CreateRepoResponse, error) {
	if p.Owner == "" || p.Name == "" {
		return nil, fmt.Errorf("github: create repo owner/name 不能为空")
	}

	body, err := buildCreateRepoBody(p)
	if err != nil {
		return nil, fmt.Errorf("github: create repo 构造 body 失败: %w", err)
	}

	respBody, err := c.http.PostJSON(ctx, "/repositories", body)
	if err != nil {
		return nil, fmt.Errorf("github: create repo 请求失败: %w", err)
	}

	if code := c.http.SnapshotResponseStatusCode(); code < 200 || code >= 300 {
		return nil, fmt.Errorf("github: create repo 返回非 2xx 状态码 %d: %s", code, bodySnippet(respBody))
	}

	var resp CreateRepoResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("github: create repo 解析响应失败: %w", err)
	}
	return &resp, nil
}
