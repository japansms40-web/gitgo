package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// ListRepos 拉取当前登录用户（owner:@me）的仓库列表第 page 页。
// 复刻 docs/flows_1_requests/01_GET_repos.txt：GET /repos?q=owner:@me&page=N。
//
// page < 1 会归一到 1。非 2xx（如 Cookie 过期返回的 401/403）返回带状态码的 error——
// gohttpkit 默认链不替你判非 2xx，这里显式校验。
func (c *Client) ListRepos(ctx context.Context, page int) (*ReposResponse, error) {
	if page < 1 {
		page = 1
	}

	params := url.Values{
		"q":    {"owner:@me"},
		"page": {strconv.Itoa(page)},
	}

	body, err := c.http.Get(ctx, "/repos", params)
	if err != nil {
		return nil, fmt.Errorf("github: list repos 请求失败: %w", err)
	}

	if code := c.http.SnapshotResponseStatusCode(); code < 200 || code >= 300 {
		return nil, fmt.Errorf("github: list repos 返回非 2xx 状态码 %d: %s", code, bodySnippet(body))
	}

	var resp ReposResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("github: list repos 解析响应失败: %w", err)
	}
	return &resp, nil
}

// bodySnippet 截取响应体前若干字节，便于在错误里排查（避免把整页 HTML 打进日志）。
func bodySnippet(body []byte) string {
	const max = 200
	if len(body) > max {
		return string(body[:max]) + "…"
	}
	return string(body)
}
