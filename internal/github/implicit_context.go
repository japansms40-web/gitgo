package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// GetImplicitContext 拉取 GitHub 网页版 Copilot 聊天的「隐式上下文」：给定仓库 owner/repo 与
// 当前所在的页面路径 contextPath（如 "/owner/repo/new/main"），返回该页面对应的仓库上下文。
// 复刻 docs/flows_1_requests/11_GET_github-copilot-chat-implicit-context-GhHnCQtQMat.txt：
//
//	GET /github-copilot/chat/implicit-context/{owner}/{repo}/{urlencode(contextPath)}
//
// contextPath 会整段 URL 编码进最后一节（每个 "/" 变 "%2F"），gohttpkit 用完整 URL 构造请求、
// 预编码路径原样发出。非 2xx（如 Cookie 过期）返回带状态码的 error。
func (c *Client) GetImplicitContext(ctx context.Context, owner, repo, contextPath string) (*ImplicitContext, error) {
	if owner == "" || repo == "" {
		return nil, fmt.Errorf("github: implicit-context owner/repo 不能为空")
	}

	path := "/github-copilot/chat/implicit-context/" +
		url.PathEscape(owner) + "/" + url.PathEscape(repo) + "/" + url.PathEscape(contextPath)

	body, err := c.http.Get(ctx, path, nil)
	if err != nil {
		return nil, fmt.Errorf("github: implicit-context 请求失败: %w", err)
	}

	if code := c.http.SnapshotResponseStatusCode(); code < 200 || code >= 300 {
		return nil, fmt.Errorf("github: implicit-context 返回非 2xx 状态码 %d: %s", code, bodySnippet(body))
	}

	var resp ImplicitContext
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("github: implicit-context 解析响应失败: %w", err)
	}
	return &resp, nil
}
