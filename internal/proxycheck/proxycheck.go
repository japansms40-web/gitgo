// Package proxycheck 提供一个「代理连通性拨测」：走给定代理请求 github.com，拿到任何 HTTP
// 响应即算连通（只有网络层失败才判不通）。用于代理设置页的「测试连通」。
package proxycheck

import (
	"context"
	"time"

	"github.com/japansms40-web/gohttpkit/httpx"
)

// githubTarget 是拨测目标：github.com 上的轻量端点（robots.txt 仅 ~2KB），
// 用它探连通比拉整个首页（数百 KB）快得多，拿到 200 即证明代理能通 github.com。
const githubTarget = "https://github.com/robots.txt"

// defaultTimeout 单次拨测超时；测试是即时反馈，不重试。
const defaultTimeout = 15 * time.Second

// Result 是一次拨测的结果。
type Result struct {
	// StatusCode 目标返回的 HTTP 状态码（能拿到就说明连通）。
	StatusCode int
	// Latency 请求往返耗时。
	Latency time.Duration
}

// Check 走 proxyURL（socks5:// / http:// / https://，可带 user:pass@；空=直连）拨测 github.com。
// 拿到任何 HTTP 响应返回 Result、err 为 nil；网络层拨不通返回 err。
func Check(ctx context.Context, proxyURL string) (Result, error) {
	return checkURL(ctx, proxyURL, githubTarget)
}

// checkURL 是 Check 的可测内核：target 可注入，便于用 httptest 做单测。
func checkURL(ctx context.Context, proxyURL, target string) (Result, error) {
	client, err := httpx.New(httpx.Options{
		Headers:  httpx.StaticHeaders{},
		ProxyURL: proxyURL,
		Timeout:  defaultTimeout,
		Retry:    httpx.NoRetry(), // 拨测一次即反馈，别 3 次重试拖慢
	})
	if err != nil {
		return Result{}, err
	}

	start := time.Now()
	_, err = client.Get(ctx, target, nil)
	latency := time.Since(start)
	if err != nil {
		return Result{}, err
	}
	return Result{StatusCode: client.SnapshotResponseStatusCode(), Latency: latency}, nil
}
