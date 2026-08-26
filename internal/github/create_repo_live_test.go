package github

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"
)

var dotcomUserRe = regexp.MustCompile(`dotcom_user=([^;]+)`)

// ownerFromCookie 从会话 Cookie 串里取 dotcom_user 作为 owner；取不到返回空。
func ownerFromCookie(cookie string) string {
	m := dotcomUserRe.FindStringSubmatch(cookie)
	if len(m) == 2 {
		return m[1]
	}
	return ""
}

// TestCreateRepo_Live 真实新建一个仓库——【写操作】，双重开关守卫：
//   - 需 GH_TEST_COOKIE
//   - 还需 GH_TEST_ALLOW_MUTATE=1 显式放行
//
// owner 优先取 GH_TEST_OWNER，否则从 Cookie 的 dotcom_user 推断。仓库名用时间戳避免撞名。
// 跑法：
//
//	GH_TEST_COOKIE='...' GH_TEST_ALLOW_MUTATE=1 go test -run TestCreateRepo_Live ./internal/github/
func TestCreateRepo_Live(t *testing.T) {
	if testing.Short() {
		t.Skip("集成测试需真实网络与有效 Cookie，-short 模式跳过")
	}
	if os.Getenv("GH_TEST_ALLOW_MUTATE") != "1" {
		t.Skip("写操作测试默认跳过：设 GH_TEST_ALLOW_MUTATE=1 才真建仓库")
	}

	cookie := testCookie(t)

	owner := os.Getenv("GH_TEST_OWNER")
	if owner == "" {
		owner = ownerFromCookie(cookie)
	}
	if owner == "" {
		t.Skip("无法确定 owner：设 GH_TEST_OWNER 或让 Cookie 带 dotcom_user")
	}

	var opts []Option
	if p := testProxy(); p != "" {
		t.Logf("走代理: %s", p)
		opts = append(opts, WithProxy(p))
	}

	c, err := New(cookie, opts...)
	if err != nil {
		t.Fatalf("构造客户端失败: %v", err)
	}

	name := fmt.Sprintf("claude-test-repo-%d", time.Now().Unix())
	resp, err := c.CreateRepo(context.Background(), CreateRepoParams{
		Owner:      owner,
		Name:       name,
		Visibility: "public",
	})
	if err != nil {
		t.Fatalf("CreateRepo 失败: %v", err)
	}

	if !strings.Contains(resp.Data.Redirect, owner+"/"+name) {
		t.Errorf("Redirect = %q, 期望含 %q", resp.Data.Redirect, owner+"/"+name)
	}
	t.Logf("已建仓库 %s/%s；跳转 %s", owner, name, resp.Data.Redirect)
}
