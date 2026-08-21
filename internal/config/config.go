package config

import (
	"errors"

	"fyne.io/fyne/v2"
)

// Config 保存一次发布所需的全部设置。不依赖任何 UI 类型。
type Config struct {
	Token       string
	Owner       string
	Repo        string
	Branch      string
	Dir         string // 仓库内目标目录，如 "posts"
	AutoCreate  bool   // 仓库不存在时自动创建
	IntervalSec int    // 每篇之间的等待秒数
	Retries     int    // 单篇失败重试次数
}

// Validate 检查必填项。
func (c Config) Validate() error {
	if c.Token == "" {
		return errors.New("Token 不能为空")
	}
	if c.Owner == "" {
		return errors.New("仓库 owner 不能为空")
	}
	if c.Repo == "" {
		return errors.New("仓库名不能为空")
	}
	return nil
}

// Normalize 补默认值并纠正非法数值。
func (c *Config) Normalize() {
	if c.Branch == "" {
		c.Branch = "main"
	}
	if c.Dir == "" {
		c.Dir = "posts"
	}
	if c.IntervalSec < 0 {
		c.IntervalSec = 0
	}
	if c.Retries < 0 {
		c.Retries = 0
	}
}

const (
	keyToken    = "token"
	keyOwner    = "owner"
	keyRepo     = "repo"
	keyBranch   = "branch"
	keyDir      = "dir"
	keyAuto     = "autoCreate"
	keyInterval = "intervalSec"
	keyRetries  = "retries"
)

// Load 从 Fyne Preferences 读取（Token 也持久化，方便重复使用）。
func Load(p fyne.Preferences) Config {
	c := Config{
		Token:       p.String(keyToken),
		Owner:       p.String(keyOwner),
		Repo:        p.String(keyRepo),
		Branch:      p.StringWithFallback(keyBranch, "main"),
		Dir:         p.StringWithFallback(keyDir, "posts"),
		AutoCreate:  p.Bool(keyAuto),
		IntervalSec: p.IntWithFallback(keyInterval, 1),
		Retries:     p.IntWithFallback(keyRetries, 2),
	}
	return c
}

// Save 写回 Fyne Preferences。
func (c Config) Save(p fyne.Preferences) {
	p.SetString(keyToken, c.Token)
	p.SetString(keyOwner, c.Owner)
	p.SetString(keyRepo, c.Repo)
	p.SetString(keyBranch, c.Branch)
	p.SetString(keyDir, c.Dir)
	p.SetBool(keyAuto, c.AutoCreate)
	p.SetInt(keyInterval, c.IntervalSec)
	p.SetInt(keyRetries, c.Retries)
}
