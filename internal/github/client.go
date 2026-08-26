// Package github 复刻 GitHub 网页版（github.com，非 api.github.com）内部 JSON 接口的调用。
// 所有 HTTP 走项目统一入口 internal/httpclient（gohttpkit），靠浏览器 Cookie 会话鉴权。
//
// 当前只实现仓库列表 GET /repos；Client 是薄 wrapper，后续接口直接往上挂方法即可。
package github

import (
	"github.com/japansms40-web/gohttpkit/httpx"

	"gitmd/internal/httpclient"
)

// baseURL 是 GitHub 网页版域名。这些接口挂在 github.com 上，不是 api.github.com。
const baseURL = "https://github.com"

// webUserAgent 复刻抓包同款 Chrome UA，尽量贴近真实浏览器指纹。
const webUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/147.0.0.0 Safari/537.36"

// Client 是 GitHub 网页接口客户端。薄薄包住 gohttpkit 的 *httpx.Client，端点方法挂在其上。
type Client struct {
	http *httpx.Client
}

// LastStatusCode 返回本客户端最近一次请求的响应状态码（0 表示尚无响应，如网络层失败）。
// 用于调用方按状态码分类（如账号验活时 401/403 判坏号）。
func (c *Client) LastStatusCode() int {
	return c.http.SnapshotResponseStatusCode()
}

// Option 调整客户端构造参数（函数式选项，便于后续加代理/超时等而不破坏 New 签名）。
type Option func(*httpclient.Config)

// WithProxy 让客户端走代理，支持 socks5:// / http:// / https://（可带 user:pass@）。
// github.com 常需代理才能连，空串等于直连。
func WithProxy(proxyURL string) Option {
	return func(cfg *httpclient.Config) { cfg.ProxyURL = proxyURL }
}

// New 用给定的浏览器会话 Cookie 构造客户端。cookie 传抓包/浏览器里那一整串
// （形如 "_gh_sess=...; user_session=...; dotcom_user=..."）。
//
// Cookie 作为入参而非写死进包，生产代码不落明文；测试里才用抓包常量。
func New(cookie string, opts ...Option) (*Client, error) {
	cfg := httpclient.Config{
		BaseURL: baseURL,
		// key 全小写：gohttpkit 直写底层 map，大写会成为指纹差异。
		Headers: map[string]string{
			"accept":                "application/json",
			"content-type":          "application/json",
			"x-requested-with":      "XMLHttpRequest",
			"github-is-react":       "true",
			"github-verified-fetch": "true",
			"x-github-app-type":     "dataRouter",
			"user-agent":            webUserAgent,
			"accept-language":       "zh-CN,zh;q=0.9",
			"referer":               baseURL + "/repos",
			"cookie":                cookie,
		},
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	c, err := httpclient.New(cfg)
	if err != nil {
		return nil, err
	}
	return &Client{http: c}, nil
}
