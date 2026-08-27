package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestBackupLinksFile 验证新一轮发布清空前，backupLinksFile 会把上一轮的
// 查看链接.txt 复制一份到 查看链接历史/ 子目录（带时间戳），非空才备、失败不误报。
func TestBackupLinksFile(t *testing.T) {
	ts := time.Date(2026, 8, 27, 15, 30, 0, 0, time.UTC)

	t.Run("非空链接文件复制到历史子目录", func(t *testing.T) {
		dir := t.TempDir()
		want := "https://a\nhttps://b\n"
		if err := os.WriteFile(filepath.Join(dir, LinksFileName), []byte(want), 0o644); err != nil {
			t.Fatal(err)
		}

		dst, err := backupLinksFile(dir, ts)
		if err != nil {
			t.Fatalf("备份不该出错: %v", err)
		}
		if dst == "" {
			t.Fatal("非空文件应产生备份路径")
		}

		gotBackup, err := os.ReadFile(dst)
		if err != nil {
			t.Fatalf("读备份文件失败: %v", err)
		}
		if string(gotBackup) != want {
			t.Errorf("备份内容 = %q, 期望 %q", gotBackup, want)
		}
		// 时间戳命名 + 落在历史子目录下
		if wantName := "查看链接-20260827-153000.txt"; filepath.Base(dst) != wantName {
			t.Errorf("备份文件名 = %q, 期望 %q", filepath.Base(dst), wantName)
		}
		if filepath.Base(filepath.Dir(dst)) != LinksBackupDir {
			t.Errorf("备份应落在 %q 子目录, 得到 %q", LinksBackupDir, dst)
		}
		// 备份只复制、不动原文件（清空由调用方另做）
		if src, _ := os.ReadFile(filepath.Join(dir, LinksFileName)); string(src) != want {
			t.Errorf("原文件被改动: %q", src)
		}
	})

	t.Run("空文件不备份", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, LinksFileName), []byte{}, 0o644); err != nil {
			t.Fatal(err)
		}
		dst, err := backupLinksFile(dir, ts)
		if err != nil || dst != "" {
			t.Fatalf("空文件应跳过备份, 得到 dst=%q err=%v", dst, err)
		}
		if _, err := os.Stat(filepath.Join(dir, LinksBackupDir)); !os.IsNotExist(err) {
			t.Error("空文件不该建历史子目录")
		}
	})

	t.Run("文件不存在不备份也不报错", func(t *testing.T) {
		dir := t.TempDir()
		dst, err := backupLinksFile(dir, ts)
		if err != nil || dst != "" {
			t.Fatalf("缺文件应跳过备份, 得到 dst=%q err=%v", dst, err)
		}
	})
}
