// Package config 把内容生成参数持久化到本机配置文件，通用读写委托给 dfclientkit。
package config

import (
	"dfclientkit/appconfig"

	"gitmd/internal/contentgen"
)

const appName = "gitmd"

func defaults() contentgen.Options {
	return contentgen.Options{
		Count:            5,
		KeywordOrder:     contentgen.OrderSequential,
		KeywordTransform: contentgen.TransformNone,
	}
}

// Load 读取磁盘上保存的生成参数；不存在则返回默认值。
func Load() contentgen.Options {
	opts := appconfig.Load(appName, defaults())
	opts.Normalize()
	return opts
}

// Save 把生成参数写入磁盘。
func Save(opts contentgen.Options) error {
	opts.Normalize()
	return appconfig.Save(appName, opts)
}
