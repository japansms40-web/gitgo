package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// Config 保存一次发布所需的全部设置。不依赖任何 UI 类型。
// json tag 同时决定磁盘持久化格式与前端绑定收到的字段名。
type Config struct {
	Token       string `json:"token"`
	Owner       string `json:"owner"`
	Repo        string `json:"repo"`
	Branch      string `json:"branch"`
	Dir         string `json:"dir"`         // 仓库内目标目录，如 "posts"
	AutoCreate  bool   `json:"autoCreate"`  // 仓库不存在时自动创建
	IntervalSec int    `json:"intervalSec"` // 每篇之间的等待秒数
	Retries     int    `json:"retries"`     // 单篇失败重试次数
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

// DefaultPath 返回配置文件在当前系统上的默认存放路径
// （macOS: ~/Library/Application Support/ghpublisher/config.json，
// Windows: %AppData%/ghpublisher/config.json）。
func DefaultPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "ghpublisher", "config.json"), nil
}

// Load 从 path 指向的 JSON 文件读取配置；文件不存在时返回带默认值的空配置。
func Load(path string) Config {
	c := Config{Branch: "main", Dir: "posts", IntervalSec: 1, Retries: 2}
	b, err := os.ReadFile(path)
	if err != nil {
		return c
	}
	_ = json.Unmarshal(b, &c)
	return c
}

// Save 把配置写入 path 指向的 JSON 文件，按需创建父目录。
func (c Config) Save(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o600)
}
