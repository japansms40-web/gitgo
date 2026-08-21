package github

import (
	"context"

	"githubbaidu/internal/publisher"
)

// Adapter 把 *Client 适配成 publisher.GitHubClient。
type Adapter struct {
	c *Client
}

// NewAdapter 包装 Client。
func NewAdapter(c *Client) *Adapter { return &Adapter{c: c} }

func (a *Adapter) GetFileSHA(ctx context.Context, owner, repo, path, branch string) (string, error) {
	return a.c.GetFileSHA(ctx, owner, repo, path, branch)
}

func (a *Adapter) PutFile(ctx context.Context, p publisher.PutParams) (string, error) {
	return a.c.PutFile(ctx, PutParams{
		Owner: p.Owner, Repo: p.Repo, Path: p.Path, Branch: p.Branch,
		Message: p.Message, Content: p.Content, SHA: p.SHA,
	})
}
