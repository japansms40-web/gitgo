package proxycheck

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestCheckURL_Direct 直连（空代理）拨测本地 httptest 服务，应拿到状态码、无错误。
func TestCheckURL_Direct(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	res, err := checkURL(context.Background(), "", srv.URL)
	if err != nil {
		t.Fatalf("直连拨测应成功，得到错误: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, 期望 200", res.StatusCode)
	}
	if res.Latency <= 0 {
		t.Errorf("Latency = %v, 期望 > 0", res.Latency)
	}
}

// TestCheckURL_BadProxy 指向一个没人监听的 socks5 代理，应返回错误（拨不通）。
func TestCheckURL_BadProxy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	_, err := checkURL(context.Background(), "socks5://127.0.0.1:1", srv.URL)
	if err == nil {
		t.Fatalf("坏代理应返回错误，实际 nil")
	}
}

// TestCheckURL_AnyStatusIsReachable 目标返回 4xx 也算「连通」（只有网络层失败才判不通）。
func TestCheckURL_AnyStatusIsReachable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	res, err := checkURL(context.Background(), "", srv.URL)
	if err != nil {
		t.Fatalf("拿到 HTTP 响应即算连通，不应报错: %v", err)
	}
	if res.StatusCode != http.StatusForbidden {
		t.Errorf("StatusCode = %d, 期望 403", res.StatusCode)
	}
}
