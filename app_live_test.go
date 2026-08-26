package main

import (
	"context"
	"os"
	"testing"
)

// TestCheckAccount_Live 真实验活：用 GH_TEST_COOKIE 的会话 + GH_TEST_PROXY/系统代理，
// 走 CheckAccount → ListRepos，断言判为活号。只读操作，-short 或无 Cookie 时跳过。
func TestCheckAccount_Live(t *testing.T) {
	if testing.Short() {
		t.Skip("集成测试需真实网络与有效 Cookie，-short 跳过")
	}
	ck := os.Getenv("GH_TEST_COOKIE")
	if ck == "" {
		t.Skip("设 GH_TEST_COOKIE 以跑账号验活集成测试")
	}
	proxy := ""
	for _, k := range []string{"GH_TEST_PROXY", "GITHUB_TEST_PROXY", "https_proxy", "HTTPS_PROXY", "http_proxy", "HTTP_PROXY"} {
		if v := os.Getenv(k); v != "" {
			proxy = v
			break
		}
	}

	a := &App{ctx: context.Background()}
	res := a.CheckAccount(ck, proxy)
	if !res.Ok {
		t.Fatalf("期望活号，得到 %+v", res)
	}
	t.Logf("验活结果: %+v", res)
}
