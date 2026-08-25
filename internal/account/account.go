// Package account 是 dfclientkit/account 在本项目内的别名层，保留
// githubbaidu/internal/account 这个导入路径不变，方便调用方少改代码；同时提供
// 本项目专属的默认存盘路径。
package account

import (
	"path/filepath"

	dfaccount "dfclientkit/account"
	"dfclientkit/appconfig"
)

type Account = dfaccount.Account

const (
	StatusPending = dfaccount.StatusPending
	StatusRunning = dfaccount.StatusRunning
	StatusSuccess = dfaccount.StatusSuccess
	StatusFailed  = dfaccount.StatusFailed
)

var ParseImportText = dfaccount.ParseImportText

const appName = "ghpublisher"

// DefaultPath 返回账号队列在当前系统上的默认存放路径。
func DefaultPath() (string, error) {
	dir, err := appconfig.DefaultDir(appName)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "accounts.json"), nil
}

func Load(path string) []Account             { return dfaccount.Load(path) }
func Save(path string, list []Account) error { return dfaccount.Save(path, list) }
