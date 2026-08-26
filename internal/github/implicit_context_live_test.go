package github

import (
	"context"
	"testing"
)

// TestGetImplicitContext_Live 真实打 github.com 的集成测试：先用 ListRepos 取一个真实仓库，
// 再请求该仓库「新建文件」页的 Copilot 隐式上下文，断言回显的 owner/repo 与请求一致。
// 自包含，只需 GH_TEST_COOKIE（+ 可选代理）。
//
// 跑法：GH_TEST_COOKIE='...' go test -run TestGetImplicitContext_Live ./internal/github/
func TestGetImplicitContext_Live(t *testing.T) {
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

	ctx := context.Background()

	repos, err := c.ListRepos(ctx, 1)
	if err != nil {
		t.Fatalf("ListRepos 失败（Cookie 可能已过期）: %v", err)
	}
	list := repos.Payload.ReposFinderPageRoute.Repositories
	if len(list) == 0 {
		t.Skip("该账号无仓库，跳过隐式上下文集成测试")
	}
	repo := list[0]
	owner, name := repo.Owner, repo.Name

	// 「新建文件」页的页面路径，形如 /owner/repo/new/main。
	contextPath := "/" + owner + "/" + name + "/new/main"

	ic, err := c.GetImplicitContext(ctx, owner, name, contextPath)
	if err != nil {
		t.Fatalf("GetImplicitContext 失败: %v", err)
	}

	if ic.RepoOwner != owner {
		t.Errorf("RepoOwner = %q, 期望 %q", ic.RepoOwner, owner)
	}
	if ic.RepoName != name {
		t.Errorf("RepoName = %q, 期望 %q", ic.RepoName, name)
	}
	t.Logf("owner=%s repo=%s repoID=%d ref=%s commit=%s", ic.RepoOwner, ic.RepoName, ic.RepoID, ic.Ref, ic.CommitOID)
}
