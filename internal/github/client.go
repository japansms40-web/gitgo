package github

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// Client 是对 GitHub REST API 的最小封装，仅用标准库。
type Client struct {
	token   string
	baseURL string
	http    *http.Client
}

// New 创建默认客户端。
func New(token string) *Client {
	return &Client{
		token:   token,
		baseURL: "https://api.github.com",
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) newRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}

func bytesReader(b []byte) *bytes.Reader { return bytes.NewReader(b) }

// apiError 携带状态码，便于上层区分 404/403 等。
type apiError struct {
	StatusCode int
	Message    string
}

func (e *apiError) Error() string {
	return fmt.Sprintf("github api %d: %s", e.StatusCode, e.Message)
}

func decodeMessage(body []byte) string {
	var m struct {
		Message string `json:"message"`
	}
	_ = json.Unmarshal(body, &m)
	return m.Message
}

// ValidateToken 调用 GET /user，返回登录名。
func (c *Client) ValidateToken(ctx context.Context) (string, error) {
	req, err := c.newRequest(ctx, http.MethodGet, "/user", nil)
	if err != nil {
		return "", err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", &apiError{resp.StatusCode, decodeMessage(b)}
	}
	var u struct {
		Login string `json:"login"`
	}
	if err := json.Unmarshal(b, &u); err != nil {
		return "", err
	}
	return u.Login, nil
}

// RepoExists 调用 GET /repos/{owner}/{repo}，200=存在，404=不存在。
func (c *Client) RepoExists(ctx context.Context, owner, repo string) (bool, error) {
	req, err := c.newRequest(ctx, http.MethodGet, "/repos/"+owner+"/"+repo, nil)
	if err != nil {
		return false, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	switch resp.StatusCode {
	case 200:
		return true, nil
	case 404:
		return false, nil
	default:
		return false, &apiError{resp.StatusCode, decodeMessage(b)}
	}
}

// CreateRepo 在当前用户下创建公开仓库并 auto_init（生成初始 commit，使默认分支存在）。
func (c *Client) CreateRepo(ctx context.Context, repo string) error {
	payload, _ := json.Marshal(map[string]any{
		"name":      repo,
		"private":   false,
		"auto_init": true,
	})
	req, err := c.newRequest(ctx, http.MethodPost, "/user/repos", bytesReader(payload))
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 201 {
		return &apiError{resp.StatusCode, decodeMessage(b)}
	}
	return nil
}

// GetFileSHA 返回文件当前 sha；文件不存在返回空字符串且 err 为 nil。
func (c *Client) GetFileSHA(ctx context.Context, owner, repo, path, branch string) (string, error) {
	p := "/repos/" + owner + "/" + repo + "/contents/" + path + "?ref=" + branch
	req, err := c.newRequest(ctx, http.MethodGet, p, nil)
	if err != nil {
		return "", err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	switch resp.StatusCode {
	case 200:
		var f struct {
			SHA string `json:"sha"`
		}
		if err := json.Unmarshal(b, &f); err != nil {
			return "", err
		}
		return f.SHA, nil
	case 404:
		return "", nil
	default:
		return "", &apiError{resp.StatusCode, decodeMessage(b)}
	}
}

// PutParams 是创建/更新单个文件的参数。
type PutParams struct {
	Owner, Repo, Path, Branch string
	Message                   string
	Content                   []byte
	SHA                       string // 空=新建；非空=更新
}

// PutFile 创建或更新文件；遇 403/429+Retry-After 会等待后重试一次。
// 返回文件的 html_url。
func (c *Client) PutFile(ctx context.Context, p PutParams) (string, error) {
	body := map[string]any{
		"message": p.Message,
		"content": base64.StdEncoding.EncodeToString(p.Content),
		"branch":  p.Branch,
	}
	if p.SHA != "" {
		body["sha"] = p.SHA
	}
	payload, _ := json.Marshal(body)
	path := "/repos/" + p.Owner + "/" + p.Repo + "/contents/" + p.Path

	for attempt := 0; attempt < 2; attempt++ {
		req, err := c.newRequest(ctx, http.MethodPut, path, bytesReader(payload))
		if err != nil {
			return "", err
		}
		resp, err := c.http.Do(req)
		if err != nil {
			return "", err
		}
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode == 200 || resp.StatusCode == 201 {
			var r struct {
				Content struct {
					HTMLURL string `json:"html_url"`
				} `json:"content"`
			}
			_ = json.Unmarshal(b, &r)
			return r.Content.HTMLURL, nil
		}
		// 限流：等待 Retry-After 后重试一次
		if (resp.StatusCode == 403 || resp.StatusCode == 429) && attempt == 0 {
			wait := 60
			if ra := resp.Header.Get("Retry-After"); ra != "" {
				if n, e := strconv.Atoi(ra); e == nil {
					wait = n
				}
			}
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(time.Duration(wait) * time.Second):
			}
			continue
		}
		return "", &apiError{resp.StatusCode, decodeMessage(b)}
	}
	return "", &apiError{StatusCode: 0, Message: "重试后仍失败"}
}
