package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config 保存发布任务的运行参数。不依赖任何 UI 类型。
// json tag 同时决定磁盘持久化格式与前端绑定收到的字段名。
type Config struct {
	Threads          int  `json:"threads"`          // 并发线程数
	IntervalSec      int  `json:"intervalSec"`      // 同一账号相邻两次发布尝试之间的等待秒数
	PerAccountCount  int  `json:"perAccountCount"`  // 单个账号最多发布多少次
	FailSwitchCount  int  `json:"failSwitchCount"`  // 账号连续失败达到此次数就换号
	CycleRounds      int  `json:"cycleRounds"`      // 账号池整体循环轮数
	RoundIntervalSec int  `json:"roundIntervalSec"` // 相邻两轮之间的等待秒数
	KeywordSlots     int  `json:"keywordSlots"`     // 内容里的关键词插入位数量
	CreateRepo       bool `json:"createRepo"`       // 处理账号前是否先建仓库/空间
}

// Normalize 纠正非法数值。
func (c *Config) Normalize() {
	if c.Threads < 1 {
		c.Threads = 1
	}
	if c.IntervalSec < 0 {
		c.IntervalSec = 0
	}
	if c.PerAccountCount < 1 {
		c.PerAccountCount = 1
	}
	if c.FailSwitchCount < 1 {
		c.FailSwitchCount = 1
	}
	if c.CycleRounds < 1 {
		c.CycleRounds = 1
	}
	if c.RoundIntervalSec < 0 {
		c.RoundIntervalSec = 0
	}
	if c.KeywordSlots < 0 {
		c.KeywordSlots = 0
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

func defaults() Config {
	return Config{
		Threads:          1,
		IntervalSec:      1,
		PerAccountCount:  1,
		FailSwitchCount:  3,
		CycleRounds:      1,
		RoundIntervalSec: 1,
	}
}

// Load 从 path 指向的 JSON 文件读取配置；文件不存在时返回带默认值的空配置。
func Load(path string) Config {
	c := defaults()
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
