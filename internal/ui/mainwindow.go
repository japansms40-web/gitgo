package ui

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"githubbaidu/internal/article"
	"githubbaidu/internal/config"
	"githubbaidu/internal/github"
	"githubbaidu/internal/publisher"
)

// Build 装配并返回主窗口内容。
func Build(w fyne.Window, prefs fyne.Preferences) fyne.CanvasObject {
	cfg := config.Load(prefs)

	// ---- 设置区 ----
	tokenEntry := widget.NewPasswordEntry()
	tokenEntry.SetText(cfg.Token)
	ownerEntry := widget.NewEntry()
	ownerEntry.SetText(cfg.Owner)
	repoEntry := widget.NewEntry()
	repoEntry.SetText(cfg.Repo)
	branchEntry := widget.NewEntry()
	branchEntry.SetText(cfg.Branch)
	dirEntry := widget.NewEntry()
	dirEntry.SetText(cfg.Dir)
	intervalEntry := widget.NewEntry()
	intervalEntry.SetText(strconv.Itoa(cfg.IntervalSec))
	retriesEntry := widget.NewEntry()
	retriesEntry.SetText(strconv.Itoa(cfg.Retries))
	autoCreateCheck := widget.NewCheck("仓库不存在时自动创建", nil)
	autoCreateCheck.SetChecked(cfg.AutoCreate)
	tokenStatus := widget.NewLabel("")

	readConfig := func() config.Config {
		iv, _ := strconv.Atoi(strings.TrimSpace(intervalEntry.Text))
		rt, _ := strconv.Atoi(strings.TrimSpace(retriesEntry.Text))
		c := config.Config{
			Token:       strings.TrimSpace(tokenEntry.Text),
			Owner:       strings.TrimSpace(ownerEntry.Text),
			Repo:        strings.TrimSpace(repoEntry.Text),
			Branch:      strings.TrimSpace(branchEntry.Text),
			Dir:         strings.TrimSpace(dirEntry.Text),
			AutoCreate:  autoCreateCheck.Checked,
			IntervalSec: iv,
			Retries:     rt,
		}
		c.Normalize()
		return c
	}

	validateBtn := widget.NewButton("验证 Token", func() {
		c := readConfig()
		if c.Token == "" {
			dialog.ShowError(fmt.Errorf("请先填写 Token"), w)
			return
		}
		tokenStatus.SetText("验证中…")
		go func() {
			login, err := github.New(c.Token).ValidateToken(context.Background())
			fyne.Do(func() {
				if err != nil {
					tokenStatus.SetText("无效: " + err.Error())
				} else {
					tokenStatus.SetText("✓ 已登录: " + login)
				}
			})
		}()
	})

	settingsForm := widget.NewForm(
		widget.NewFormItem("Token", container.NewBorder(nil, nil, nil, validateBtn, tokenEntry)),
		widget.NewFormItem("", tokenStatus),
		widget.NewFormItem("Owner", ownerEntry),
		widget.NewFormItem("Repo", repoEntry),
		widget.NewFormItem("分支", branchEntry),
		widget.NewFormItem("目标目录", dirEntry),
		widget.NewFormItem("发布间隔(秒)", intervalEntry),
		widget.NewFormItem("失败重试次数", retriesEntry),
		widget.NewFormItem("", autoCreateCheck),
	)

	// ---- 队列表 ----
	var arts []article.Article
	var rowState *RowState

	table := widget.NewTable(
		func() (int, int) { return len(arts) + 1, 3 }, // +1 表头
		func() fyne.CanvasObject { return widget.NewLabel("cell") },
		func(id widget.TableCellID, o fyne.CanvasObject) {
			lbl := o.(*widget.Label)
			if id.Row == 0 {
				lbl.SetText([]string{"文件名", "仓库路径", "状态"}[id.Col])
				lbl.TextStyle.Bold = true
				return
			}
			lbl.TextStyle.Bold = false
			a := arts[id.Row-1]
			switch id.Col {
			case 0:
				lbl.SetText(a.Title)
			case 1:
				lbl.SetText(a.RepoPath)
			case 2:
				if rowState != nil {
					lbl.SetText(rowState.Status(id.Row - 1))
				} else {
					lbl.SetText("待发布")
				}
			}
		},
	)
	table.SetColumnWidth(0, 200)
	table.SetColumnWidth(1, 260)
	table.SetColumnWidth(2, 220)

	logBox := widget.NewMultiLineEntry()
	logBox.Wrapping = fyne.TextWrapWord
	appendLog := func(s string) { logBox.SetText(logBox.Text + s + "\n") }

	progress := widget.NewLabel("0/0")

	reloadQueue := func(paths []string) {
		c := readConfig()
		list, err := article.ScanPaths(paths, c.Dir)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		arts = list
		rowState = nil
		progress.SetText(fmt.Sprintf("0/%d", len(arts)))
		table.Refresh()
	}

	addFolderBtn := widget.NewButton("选择文件夹", func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil || uri == nil {
				return
			}
			reloadQueue([]string{uri.Path()})
		}, w)
	})
	addFileBtn := widget.NewButton("添加文件", func() {
		dialog.ShowFileOpen(func(rc fyne.URIReadCloser, err error) {
			if err != nil || rc == nil {
				return
			}
			defer rc.Close()
			reloadQueue([]string{rc.URI().Path()})
		}, w)
	})
	clearBtn := widget.NewButton("清空", func() {
		arts = nil
		rowState = nil
		progress.SetText("0/0")
		table.Refresh()
	})

	// ---- 控制区 ----
	var cancelFn context.CancelFunc
	startBtn := widget.NewButton("开始发布", nil)
	stopBtn := widget.NewButton("停止", nil)
	stopBtn.Disable()
	exportBtn := widget.NewButton("导出链接列表", nil)

	setRunning := func(running bool) {
		if running {
			startBtn.Disable()
			stopBtn.Enable()
		} else {
			startBtn.Enable()
			stopBtn.Disable()
		}
	}

	startBtn.OnTapped = func() {
		c := readConfig()
		if err := c.Validate(); err != nil {
			dialog.ShowError(err, w)
			return
		}
		if len(arts) == 0 {
			dialog.ShowError(fmt.Errorf("队列为空，请先添加文件"), w)
			return
		}
		c.Save(prefs)

		rowState = NewRowState(len(arts))
		table.Refresh()
		setRunning(true)
		appendLog(fmt.Sprintf("开始发布 %d 篇到 %s/%s", len(arts), c.Owner, c.Repo))

		ctx, cancel := context.WithCancel(context.Background())
		cancelFn = cancel

		go func() {
			client := github.New(c.Token)
			// 仓库检查/自动创建
			exists, err := client.RepoExists(ctx, c.Owner, c.Repo)
			if err != nil {
				fyne.Do(func() {
					dialog.ShowError(err, w)
					setRunning(false)
				})
				return
			}
			if !exists {
				if !c.AutoCreate {
					fyne.Do(func() {
						dialog.ShowError(fmt.Errorf("仓库 %s/%s 不存在（可勾选自动创建）", c.Owner, c.Repo), w)
						setRunning(false)
					})
					return
				}
				if err := client.CreateRepo(ctx, c.Repo); err != nil {
					fyne.Do(func() {
						dialog.ShowError(fmt.Errorf("创建仓库失败: %w", err), w)
						setRunning(false)
					})
					return
				}
				fyne.Do(func() { appendLog("已自动创建仓库 " + c.Repo) })
			}

			p := publisher.New(github.NewAdapter(client))
			runErr := p.Run(ctx, c, arts, func(e publisher.Event) {
				fyne.Do(func() {
					rowState.Apply(e)
					progress.SetText(rowState.Progress())
					table.Refresh()
					switch e.Kind {
					case publisher.EventSuccess:
						appendLog(fmt.Sprintf("✓ %s → %s", e.Title, e.URL))
					case publisher.EventFailure:
						appendLog(fmt.Sprintf("✗ %s 失败: %v", e.Title, e.Err))
					case publisher.EventRetry:
						appendLog(fmt.Sprintf("↻ %s 重试: %v", e.Title, e.Err))
					}
				})
			})
			fyne.Do(func() {
				if runErr != nil {
					appendLog("已停止: " + runErr.Error())
				} else {
					appendLog("全部完成")
				}
				setRunning(false)
			})
		}()
	}

	stopBtn.OnTapped = func() {
		if cancelFn != nil {
			cancelFn()
			appendLog("正在停止…")
		}
	}

	exportBtn.OnTapped = func() {
		if rowState == nil {
			dialog.ShowInformation("提示", "还没有发布结果", w)
			return
		}
		urls := rowState.SuccessURLs()
		if len(urls) == 0 {
			dialog.ShowInformation("提示", "暂无成功链接", w)
			return
		}
		dialog.ShowFileSave(func(wc fyne.URIWriteCloser, err error) {
			if err != nil || wc == nil {
				return
			}
			defer wc.Close()
			_, _ = wc.Write([]byte(strings.Join(urls, "\n")))
		}, w)
	}

	// ---- 布局 ----
	queueBar := container.NewHBox(addFolderBtn, addFileBtn, clearBtn)
	controlBar := container.NewHBox(startBtn, stopBtn, exportBtn, progress)
	top := container.NewVBox(widget.NewLabelWithStyle("设置", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), settingsForm, queueBar)
	bottom := container.NewVBox(controlBar, widget.NewLabel("日志"), logBox)
	return container.NewBorder(top, bottom, nil, nil, table)
}
