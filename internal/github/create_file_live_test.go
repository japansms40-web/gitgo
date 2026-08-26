package github

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

// TestCreateFile_Live 真实在仓库里创建文件并提交——这是【写操作】，双重开关守卫：
//   - 需 GH_TEST_COOKIE（有效会话）
//   - 还需 GH_TEST_ALLOW_MUTATE=1 显式放行（默认 go test 绝不发写请求，避免往仓库塞垃圾 commit）
//
// 跑法：
//
//	GH_TEST_COOKIE='...' GH_TEST_ALLOW_MUTATE=1 go test -run TestCreateFile_Live ./internal/github/
//
// 流程：ListRepos 取真实仓库 → GetImplicitContext 拿分支 HEAD 的 CommitOID →
// CreateFile 用时间戳文件名创建，断言返回跳转 URL 指向该仓库分支树页。
func TestCreateFile_Live(t *testing.T) {
	if testing.Short() {
		t.Skip("集成测试需真实网络与有效 Cookie，-short 模式跳过")
	}
	if os.Getenv("GH_TEST_ALLOW_MUTATE") != "1" {
		t.Skip("写操作测试默认跳过：设 GH_TEST_ALLOW_MUTATE=1 才真在仓库建文件提交")
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
		t.Fatalf("ListRepos 失败: %v", err)
	}
	list := repos.Payload.ReposFinderPageRoute.Repositories
	if len(list) == 0 {
		t.Skip("该账号无仓库，跳过创建文件集成测试")
	}
	repo := list[0]
	owner, name := repo.Owner, repo.Name
	const branch = "main"

	ic, err := c.GetImplicitContext(ctx, owner, name, "/"+owner+"/"+name+"/new/"+branch)
	if err != nil {
		t.Fatalf("GetImplicitContext 失败: %v", err)
	}

	filename := fmt.Sprintf("claude-test-%d.md", time.Now().UnixNano())
	resp, err := c.CreateFile(ctx, CreateFileParams{
		Owner:      owner,
		Repo:       name,
		Branch:     branch,
		Filename:   filename,
		Content:    "# claude 集成测试\n由 CreateFile 集成测试创建，可删。",
		Message:    "test: CreateFile 集成测试",
		BaseCommit: ic.CommitOID,
	})
	if err != nil {
		t.Fatalf("CreateFile 失败: %v", err)
	}

	if !strings.Contains(resp.Data.Redirect, owner+"/"+name) {
		t.Errorf("Redirect = %q, 期望含 %q", resp.Data.Redirect, owner+"/"+name)
	}
	t.Logf("已在 %s/%s 创建 %s；跳转 %s", owner, name, filename, resp.Data.Redirect)
}
