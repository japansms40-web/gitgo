package github

import (
	"context"
	"fmt"
)

// CreateRepoWithFileParams 是「建仓库并写入首个文件」编排的入参。
type CreateRepoWithFileParams struct {
	// Owner 仓库属主登录名。
	Owner string
	// RepoName 新仓库名。
	RepoName string
	// Visibility 仓库可见性，"public"/"private"；空则 "public"。
	Visibility string
	// RepoDescription 仓库描述；可空。
	RepoDescription string

	// Branch 首次提交的分支；空则 "main"（新仓库默认分支）。
	Branch string
	// Filename 首个文件名（含路径）。
	Filename string
	// Content 文件内容。
	Content string
	// CommitMessage 提交信息；空则用 Filename。
	CommitMessage string
	// FileDescription 扩展提交说明；可空。
	FileDescription string
}

// CreateRepoWithFileResult 汇总编排各步的响应。
type CreateRepoWithFileResult struct {
	// Repo 建仓库的响应。
	Repo *CreateRepoResponse
	// File 建文件的响应。
	File *CreateFileResponse
}

// CreateRepoWithFile 一步到位：先建一个空仓库，再在其默认分支上创建首个文件并提交。
// 新仓库无历史，首次提交不需要父 commit（CreateFile 的 BaseCommit 留空）。
//
// 若仓库已建成但建文件失败，返回的 error 会点明这一点（此时仓库已存在、为空）。
func (c *Client) CreateRepoWithFile(ctx context.Context, p CreateRepoWithFileParams) (*CreateRepoWithFileResult, error) {
	if p.Owner == "" || p.RepoName == "" {
		return nil, fmt.Errorf("github: create repo-with-file owner/name 不能为空")
	}
	if p.Filename == "" {
		return nil, fmt.Errorf("github: create repo-with-file filename 不能为空")
	}

	branch := p.Branch
	if branch == "" {
		branch = "main"
	}
	message := p.CommitMessage
	if message == "" {
		message = p.Filename
	}

	repoResp, err := c.CreateRepo(ctx, CreateRepoParams{
		Owner:       p.Owner,
		Name:        p.RepoName,
		Visibility:  p.Visibility,
		Description: p.RepoDescription,
	})
	if err != nil {
		return nil, fmt.Errorf("github: create repo-with-file 建仓库失败: %w", err)
	}

	fileResp, err := c.CreateFile(ctx, CreateFileParams{
		Owner:       p.Owner,
		Repo:        p.RepoName,
		Branch:      branch,
		Filename:    p.Filename,
		Content:     p.Content,
		Message:     message,
		Description: p.FileDescription,
		BaseCommit:  "", // 新空仓库首次提交无父 commit
	})
	if err != nil {
		return nil, fmt.Errorf("github: create repo-with-file 仓库已建成但建文件失败: %w", err)
	}

	return &CreateRepoWithFileResult{Repo: repoResp, File: fileResp}, nil
}
