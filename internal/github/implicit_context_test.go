package github

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestGetImplicitContext_DecodeResponse 离线验证响应结构体：拿
// 11_GET_github-copilot-chat-implicit-context-GhHnCQtQMat.txt 里抠出的真实响应体喂给结构体，
// 断言关键字段解得对。永远稳定、不依赖网络。
func TestGetImplicitContext_DecodeResponse(t *testing.T) {
	body, err := os.ReadFile(filepath.Join("testdata", "implicit_context_response.json"))
	if err != nil {
		t.Fatalf("读取 testdata 失败: %v", err)
	}

	var ctx ImplicitContext
	if err := json.Unmarshal(body, &ctx); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	if ctx.Type != "file" {
		t.Errorf("Type = %q, 期望 %q", ctx.Type, "file")
	}
	if ctx.RepoID != 1340931205 {
		t.Errorf("RepoID = %d, 期望 1340931205", ctx.RepoID)
	}
	if ctx.RepoOwner != "GhHnCQtQMatj02" {
		t.Errorf("RepoOwner = %q, 期望 %q", ctx.RepoOwner, "GhHnCQtQMatj02")
	}
	if ctx.RepoName != "mvrjra" {
		t.Errorf("RepoName = %q, 期望 %q", ctx.RepoName, "mvrjra")
	}
	if ctx.Ref != "main" {
		t.Errorf("Ref = %q, 期望 %q", ctx.Ref, "main")
	}
	if ctx.CommitOID != "bda8b1344eb3bf4e04726a3631e51504360ca600" {
		t.Errorf("CommitOID = %q", ctx.CommitOID)
	}
	if ctx.Path != nil {
		t.Errorf("Path = %v, 期望 nil（抓包为 null）", *ctx.Path)
	}
}
