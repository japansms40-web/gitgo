package github

import (
	"encoding/json"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCreateFile_DecodeResponse 离线验证响应结构体：拿 16_POST_..._create-main.txt 里抠出的
// 真实响应体喂给结构体，断言关键字段解得对。
func TestCreateFile_DecodeResponse(t *testing.T) {
	body, err := os.ReadFile(filepath.Join("testdata", "create_file_response.json"))
	if err != nil {
		t.Fatalf("读取 testdata 失败: %v", err)
	}

	var resp CreateFileResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	if resp.Data.Redirect != "https://github.com/VKMW57by18/abyuds/tree/main" {
		t.Errorf("Redirect = %q", resp.Data.Redirect)
	}
	if resp.Data.Message != nil {
		t.Errorf("Message = %v, 期望 nil（抓包为 null）", *resp.Data.Message)
	}
	if !strings.HasPrefix(resp.Data.CommitQuorumPollPath, "/VKMW57by18/abyuds/check_commit_quorum/") {
		t.Errorf("CommitQuorumPollPath = %q", resp.Data.CommitQuorumPollPath)
	}
}

// TestBuildCreateFileForm 离线验证 multipart 构造：不发网络，只断言生成的 content-type 带
// multipart 边界、body 里含所有必需字段与固定默认值。
func TestBuildCreateFileForm(t *testing.T) {
	p := CreateFileParams{
		Owner:       "octocat",
		Repo:        "hello",
		Branch:      "main",
		Filename:    "note.md",
		Content:     "# 标题\n正文内容",
		Message:     "add note",
		Description: "扩展说明",
		BaseCommit:  "abc123",
	}

	body, contentType, err := buildCreateFileForm(p)
	if err != nil {
		t.Fatalf("buildCreateFileForm 失败: %v", err)
	}

	// content-type 必须是 multipart/form-data 且带 boundary。
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		t.Fatalf("content-type 解析失败: %v", err)
	}
	if mediaType != "multipart/form-data" {
		t.Errorf("mediaType = %q, 期望 multipart/form-data", mediaType)
	}
	if params["boundary"] == "" {
		t.Errorf("content-type 缺 boundary: %q", contentType)
	}

	s := string(body)
	// 必需字段与固定默认值都要在 body 里。
	wants := []string{
		`name="message"`, "add note",
		`name="placeholder_message"`, // = message
		`name="description"`, "扩展说明",
		`name="commit-choice"`, "direct",
		`name="target_branch"`, "main",
		`name="commit"`, "abc123",
		`name="same_repo"`, "1",
		`name="content_changed"`, "true",
		`name="filename"`, "note.md",
		`name="new_filename"`,
		`name="value"`, "正文内容",
	}
	for _, w := range wants {
		if !strings.Contains(s, w) {
			t.Errorf("multipart body 缺片段: %q", w)
		}
	}
}
