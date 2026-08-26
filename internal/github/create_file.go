package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"

	"github.com/japansms40-web/gohttpkit/httpx"
)

// CreateFile 在 owner/repo 的 Branch 分支上创建一个新文件并直接提交（commit-choice=direct）。
// 复刻 docs/flows_1_requests/16_POST_..._create-main.txt：
//
//	POST /{owner}/{repo}/create/{branch}  (multipart/form-data)
//
// 典型用法是先 GetImplicitContext 拿到分支当前 CommitOID，作 BaseCommit 传入。
// 非 2xx（如 Cookie 过期、并发冲突）返回带状态码的 error。
func (c *Client) CreateFile(ctx context.Context, p CreateFileParams) (*CreateFileResponse, error) {
	if p.Owner == "" || p.Repo == "" || p.Branch == "" {
		return nil, fmt.Errorf("github: create file owner/repo/branch 不能为空")
	}
	if p.Filename == "" {
		return nil, fmt.Errorf("github: create file filename 不能为空")
	}

	body, contentType, err := buildCreateFileForm(p)
	if err != nil {
		return nil, fmt.Errorf("github: create file 构造表单失败: %w", err)
	}

	path := "/" + url.PathEscape(p.Owner) + "/" + url.PathEscape(p.Repo) +
		"/create/" + url.PathEscape(p.Branch)

	// 用 ExtraHeaders 把 content-type 覆盖成带 boundary 的 multipart（盖掉客户端默认的 application/json）。
	respBody, err := c.http.Do(ctx, httpx.RequestSpec{
		Method:       http.MethodPost,
		Path:         path,
		Body:         body,
		ExtraHeaders: map[string]string{"content-type": contentType},
	})
	if err != nil {
		return nil, fmt.Errorf("github: create file 请求失败: %w", err)
	}

	if code := c.http.SnapshotResponseStatusCode(); code < 200 || code >= 300 {
		return nil, fmt.Errorf("github: create file 返回非 2xx 状态码 %d: %s", code, bodySnippet(respBody))
	}

	var resp CreateFileResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("github: create file 解析响应失败: %w", err)
	}
	return &resp, nil
}

// buildCreateFileForm 用 mime/multipart 拼出创建文件的表单 body，返回 body 与带 boundary 的
// content-type。固定字段（commit-choice=direct 等）复刻抓包。
func buildCreateFileForm(p CreateFileParams) (body []byte, contentType string, err error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	fields := []struct{ name, value string }{
		{"message", p.Message},
		{"placeholder_message", p.Message},
		{"description", p.Description},
		{"commit-choice", "direct"},
		{"target_branch", p.Branch},
		{"quick_pull", ""},
		{"guidance_task", ""},
		{"commit", p.BaseCommit},
		{"same_repo", "1"},
		{"pr", ""},
		{"content_changed", "true"},
		{"filename", p.Filename},
		{"new_filename", p.Filename},
		{"value", p.Content},
	}
	for _, f := range fields {
		if err := w.WriteField(f.name, f.value); err != nil {
			return nil, "", err
		}
	}
	if err := w.Close(); err != nil {
		return nil, "", err
	}
	return buf.Bytes(), w.FormDataContentType(), nil
}
