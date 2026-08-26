package proxycheck

import (
	"context"
	"os"
	"testing"
)

// liveProxy 取集成测试用的代理：优先 GITHUB_TEST_PROXY，否则系统 HTTPS_PROXY/HTTP_PROXY。
func liveProxy() string {
	for _, k := range []string{"GITHUB_TEST_PROXY", "https_proxy", "HTTPS_PROXY", "http_proxy", "HTTP_PROXY"} {
		if v := os.Getenv(k); v != "" {
			return v
		}
	}
	return ""
}

// TestCheck_Live 真走代理拨测 github.com。-short 或无代理时跳过。
func TestCheck_Live(t *testing.T) {
	if testing.Short() {
		t.Skip("集成测试需真实网络，-short 跳过")
	}
	proxy := liveProxy()
	if proxy == "" {
		t.Skip("无代理环境变量，跳过（github.com 直连多半不通）")
	}

	res, err := Check(context.Background(), proxy)
	if err != nil {
		t.Fatalf("走代理 %s 拨测 github.com 失败: %v", proxy, err)
	}
	if res.StatusCode == 0 {
		t.Errorf("StatusCode = 0，期望拿到真实状态码")
	}
	t.Logf("代理 %s → github.com HTTP %d · %v", proxy, res.StatusCode, res.Latency.Round(1e6))
}
