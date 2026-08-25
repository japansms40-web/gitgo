// Package config 保存发布任务的运行参数，通用部分委托给 dfclientkit。
package config

import (
	"dfclientkit/appconfig"
	"dfclientkit/taskrunner"
)

const appName = "ghpublisher"

// Config 保存发布任务的运行参数。不依赖任何 UI 类型。
type Config struct {
	taskrunner.RunConfig
	KeywordSlots int `json:"keywordSlots"` // 内容里的关键词插入位数量
}

// Normalize 纠正非法数值。
func (c *Config) Normalize() {
	c.RunConfig.Normalize()
	if c.KeywordSlots < 0 {
		c.KeywordSlots = 0
	}
}

func defaults() Config {
	return Config{
		RunConfig: taskrunner.RunConfig{
			Threads:          1,
			IntervalSec:      1,
			PerAccountCount:  1,
			FailSwitchCount:  3,
			CycleRounds:      1,
			RoundIntervalSec: 1,
		},
	}
}

// Load 读取磁盘上保存的任务参数；不存在则返回默认值。
func Load() Config {
	return appconfig.Load(appName, defaults())
}

// Save 把任务参数写入磁盘。
func Save(cfg Config) error {
	return appconfig.Save(appName, cfg)
}
