package github

import (
	"context"
	"os"
	"testing"
)

// testProxy 决定集成测试走哪个代理：优先 GITHUB_TEST_PROXY，否则回退标准的
// HTTPS_PROXY / HTTP_PROXY 环境变量（大小写都认）。github.com 常需代理才能连。
func testProxy() string {
	for _, k := range []string{
		"GITHUB_TEST_PROXY",
		"https_proxy", "HTTPS_PROXY",
		"http_proxy", "HTTP_PROXY",
	} {
		if v := os.Getenv(k); v != "" {
			return v
		}
	}
	return ""
}

// testCookie 从 GH_TEST_COOKIE 读集成测试用的浏览器会话 Cookie。
// 绝不写死进仓库（本仓库公开，Cookie 泄露即会话被劫持）；没设就跳过集成测试。
func testCookie(t *testing.T) string {
	t.Helper()
	cookie := os.Getenv("GH_TEST_COOKIE")
	if cookie == "" {
		t.Skip("设 GH_TEST_COOKIE=<浏览器会话 Cookie 整串> 以跑集成测试")
	}
	return cookie
}

// TestListRepos_Live 真实打 github.com 的集成测试：用 GH_TEST_COOKIE 里的会话 Cookie 构造客户端，
// 请求 GET /repos?q=owner:@me&page=1，断言能拿到分页数据。
//
// 跑法：GH_TEST_COOKIE='_gh_sess=...; user_session=...' go test -run Live ./internal/github/
// 未设 GH_TEST_COOKIE 或 `go test -short` 时跳过（只留离线解析测试）。
// 代理取 GITHUB_TEST_PROXY 或系统 HTTPS_PROXY/HTTP_PROXY。
func TestListRepos_Live(t *testing.T) {
	if testing.Short() {
		t.Skip("集成测试需真实网络与有效 Cookie，-short 模式跳过")
	}

	cookie := testCookie(t)

	var opts []Option
	if p := testProxy(); p != "" {
		t.Logf("走代理: %s", p)
		opts = append(opts, WithProxy(p))
	}

	c, err := New(cookie, opts...)
	if err != nil {
		t.Fatalf("构造客户端失败: %v", err)
	}

	resp, err := c.ListRepos(context.Background(), 1)
	if err != nil {
		t.Fatalf("ListRepos 失败（Cookie 可能已过期，请更换 GH_TEST_COOKIE）: %v", err)
	}

	route := resp.Payload.ReposFinderPageRoute
	if route.PageCount < 1 {
		t.Errorf("PageCount = %d, 期望 >= 1", route.PageCount)
	}
	t.Logf("取到 %d 个仓库，共 %d 页", route.RepositoryCount, route.PageCount)
}
