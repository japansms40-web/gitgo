// Package httpclient 是 gitgo 统一的 HTTP 客户端入口：项目内所有对外 HTTP 调用都基于
// gohttpkit（github.com/japansms40-web/gohttpkit）的 httpx 构建，不要再裸用 net/http。
//
// 默认拦截器链已包含：结构化日志、四种压缩解码、状态码与响应头缓存、网络层重试+指数退避。
// 默认链不替你判业务死活（非 2xx 的响应体原样返回，状态码走 SnapshotResponseStatusCode 读）；
// 需要「非 2xx→error」「业务错误归类」时，改用 httpx.APIChain(classify) 或自加拦截器。
//
// 用法：
//
//	c, err := httpclient.New(httpclient.Config{
//	    BaseURL: "https://api.example.com",
//	    Headers: map[string]string{"accept": "application/json"},
//	})
//	if err != nil {
//	    return err
//	}
//	body, err := c.Get(ctx, "/v1/ping", nil)
//	status := c.SnapshotResponseStatusCode()
//
// 需要更细的控制（自定义重试、代理 Transport、会话回写 OnResponseHeaders、自定义 HeaderProvider
// 等）时，直接用 httpx.New(httpx.Options{...})，本包只是最常见场景的便捷构造。
package httpclient

import (
	"time"

	"github.com/japansms40-web/gohttpkit/httpx"
)

// Config 是项目侧构造 HTTP 客户端的最小参数集。字段留空即取 gohttpkit 默认。
type Config struct {
	// BaseURL 基础域名，如 "https://api.example.com"。相对 path 会拼在它后面。
	BaseURL string
	// Headers 候选请求头（key 请用小写）。实际发哪些由每次请求的 HeaderWhitelist 决定：
	// 传 nil 白名单则发全量候选头。
	Headers map[string]string
	// ProxyURL 代理地址，支持 socks5:// / http:// / https://（可带 user:pass@）。空 = 直连。
	ProxyURL string
	// Timeout 整请求超时（连接 + 重定向 + 读体全程）。0 = 用 gohttpkit 默认（30s，可被 env 覆盖）。
	Timeout time.Duration
}

// New 按项目约定构造一个 *httpx.Client：以 StaticHeaders 承载基础域名与候选头，
// 其余走 gohttpkit 默认（含默认拦截器链与默认重试）。
func New(cfg Config) (*httpx.Client, error) {
	return httpx.New(httpx.Options{
		Headers: httpx.StaticHeaders{
			Base:    cfg.BaseURL,
			Headers: cfg.Headers,
		},
		ProxyURL: cfg.ProxyURL,
		Timeout:  cfg.Timeout,
	})
}
