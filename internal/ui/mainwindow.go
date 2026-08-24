package ui

import (
	"context"
	"fmt"
	"image/color"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"githubbaidu/internal/article"
	"githubbaidu/internal/config"
	"githubbaidu/internal/github"
	"githubbaidu/internal/publisher"
)

const keyThemeVariant = "themeVariant"

// initialTheme 根据已保存的偏好或系统设置决定启动时使用的主题变体，
// 并把它设为当前应用主题（后续所有 currentTheme() 调用据此取色）。
func initialTheme(prefs fyne.Preferences) *Theme {
	variant := fyne.CurrentApp().Settings().ThemeVariant()
	switch prefs.String(keyThemeVariant) {
	case "light":
		variant = theme.VariantLight
	case "dark":
		variant = theme.VariantDark
	}
	t := NewTheme(variant)
	fyne.CurrentApp().Settings().SetTheme(t)
	return t
}

// themedRect 是随主题切换需要重新取色的结构背景矩形。
type themedRect struct {
	rect *canvas.Rectangle
	name fyne.ThemeColorName
}

// tableCell 是发布队列表格里一个单元格的可复用对象：文件名/仓库路径列显示
// label，状态列显示 badge，同一时刻只显示其中一个。
type tableCell struct {
	label *widget.Label
	badge *badge
}

// Build 装配并返回主窗口内容。
func Build(w fyne.Window, prefs fyne.Preferences) fyne.CanvasObject {
	initialTheme(prefs)
	cfg := config.Load(prefs)

	var chrome []themedRect
	trackRect := func(name fyne.ThemeColorName) *canvas.Rectangle {
		r := canvas.NewRectangle(currentTheme().Color(name, currentTheme().Variant))
		chrome = append(chrome, themedRect{rect: r, name: name})
		return r
	}

	// ---- 设置区（字段与原实现一致） ----
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
	settingsCard := widget.NewCard("发布设置", "", settingsForm)

	// ---- 队列表 ----
	var arts []article.Article
	var rowState *RowState
	cellPool := map[fyne.CanvasObject]*tableCell{}

	table := widget.NewTable(
		func() (int, int) { return len(arts) + 1, 3 },
		func() fyne.CanvasObject {
			lbl := widget.NewLabel("cell")
			b := newBadge()
			root := container.NewStack(lbl, b.object())
			cellPool[root] = &tableCell{label: lbl, badge: b}
			return root
		},
		func(id widget.TableCellID, o fyne.CanvasObject) {
			cell := cellPool[o]
			if id.Row == 0 {
				cell.badge.object().Hide()
				cell.label.Show()
				cell.label.TextStyle = fyne.TextStyle{Bold: true}
				cell.label.SetText([]string{"文件名", "仓库路径", "状态"}[id.Col])
				return
			}
			a := arts[id.Row-1]
			switch id.Col {
			case 0:
				cell.badge.object().Hide()
				cell.label.Show()
				cell.label.TextStyle = fyne.TextStyle{}
				cell.label.SetText(a.Title)
			case 1:
				cell.badge.object().Hide()
				cell.label.Show()
				cell.label.TextStyle = fyne.TextStyle{}
				cell.label.SetText(a.RepoPath)
			case 2:
				cell.label.Hide()
				state := "待发布"
				if rowState != nil {
					state = rowState.Status(id.Row - 1)
				}
				cell.badge.set(currentTheme(), state)
				cell.badge.object().Show()
			}
		},
	)
	table.SetColumnWidth(0, 220)
	table.SetColumnWidth(1, 280)
	table.SetColumnWidth(2, 220)

	logPanel := NewLogPanel(currentTheme())
	pill := newStatusPill()

	elapsedLabel := widget.NewLabel("00:00")
	totalLabel := widget.NewLabel("0")
	successLabel := widget.NewLabel("0")
	failLabel := widget.NewLabel("0")
	pendingLabel := widget.NewLabel("0")
	var startTime time.Time

	refreshCounts := func() {
		total := len(arts)
		success, fail := 0, 0
		if rowState != nil {
			for i := 0; i < total; i++ {
				switch {
				case rowState.Status(i) == "成功":
					success++
				case strings.HasPrefix(rowState.Status(i), "失败"):
					fail++
				}
			}
		}
		totalLabel.SetText(strconv.Itoa(total))
		successLabel.SetText(strconv.Itoa(success))
		failLabel.SetText(strconv.Itoa(fail))
		pendingLabel.SetText(strconv.Itoa(total - success - fail))
	}

	intervalSummary := widget.NewLabel("0 秒")
	retriesSummary := widget.NewLabel("0 次")
	queueSummary := widget.NewLabel("0 篇")
	refreshRightPanel := func() {
		c := readConfig()
		intervalSummary.SetText(fmt.Sprintf("%d 秒", c.IntervalSec))
		retriesSummary.SetText(fmt.Sprintf("%d 次", c.Retries))
		queueSummary.SetText(fmt.Sprintf("%d 篇", len(arts)))
	}
	intervalEntry.OnChanged = func(string) { refreshRightPanel() }
	retriesEntry.OnChanged = func(string) { refreshRightPanel() }

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
		refreshCounts()
		refreshRightPanel()
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
		refreshCounts()
		refreshRightPanel()
		table.Refresh()
	})

	// ---- 控制区 ----
	var cancelFn context.CancelFunc
	var isRunning bool
	startBtn := widget.NewButton("开始发布", nil)
	startBtn.Importance = widget.HighImportance
	stopBtn := widget.NewButton("停止", nil)
	stopBtn.Importance = widget.DangerImportance
	stopBtn.Disable()
	exportBtn := widget.NewButton("导出链接列表", nil)

	setRunning := func(running bool) {
		isRunning = running
		if running {
			startBtn.Disable()
			stopBtn.Enable()
		} else {
			startBtn.Enable()
			stopBtn.Disable()
		}
		total, done, failed := 0, 0, 0
		if rowState != nil {
			total = len(arts)
			done = rowState.Done()
			for i := 0; i < total; i++ {
				if strings.HasPrefix(rowState.Status(i), "失败") {
					failed++
				}
			}
		}
		pill.repaint(running, done, total, failed)
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
		startTime = time.Now()
		elapsedLabel.SetText("00:00")
		setRunning(true)
		logPanel.AppendInfo(fmt.Sprintf("开始发布 %d 篇到 %s/%s", len(arts), c.Owner, c.Repo))

		ctx, cancel := context.WithCancel(context.Background())
		cancelFn = cancel

		go func() {
			client := github.New(c.Token)
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
				fyne.Do(func() { logPanel.AppendInfo("已自动创建仓库 " + c.Repo) })
			}

			p := publisher.New(github.NewAdapter(client))
			runErr := p.Run(ctx, c, arts, func(e publisher.Event) {
				fyne.Do(func() {
					rowState.Apply(e)
					progress.SetText(rowState.Progress())
					refreshCounts()
					table.Refresh()
					logPanel.AppendEvent(e)
					elapsedLabel.SetText(formatElapsed(time.Since(startTime)))
					total := len(arts)
					failed := 0
					for i := 0; i < total; i++ {
						if strings.HasPrefix(rowState.Status(i), "失败") {
							failed++
						}
					}
					pill.repaint(true, rowState.Done(), total, failed)
				})
			})
			fyne.Do(func() {
				elapsedLabel.SetText(formatElapsed(time.Since(startTime)))
				if runErr != nil {
					logPanel.AppendInfo("已停止: " + runErr.Error())
				} else {
					logPanel.AppendInfo("全部完成")
				}
				setRunning(false)
			})
		}()
	}

	stopBtn.OnTapped = func() {
		if cancelFn != nil {
			cancelFn()
			logPanel.AppendInfo("正在停止…")
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

	saveBtn := widget.NewButton("保存配置", func() {
		readConfig().Save(prefs)
		logPanel.AppendInfo("配置已保存")
	})

	// ---- 日志面板工具条 ----
	autoScrollCheck := widget.NewCheck("自动滚动", func(on bool) { logPanel.SetAutoScroll(on) })
	autoScrollCheck.SetChecked(true)
	copyLogBtn := widget.NewButton("复制", func() {
		fyne.CurrentApp().Clipboard().SetContent(logPanel.Text())
	})
	exportLogBtn := widget.NewButton("导出", func() {
		dialog.ShowFileSave(func(wc fyne.URIWriteCloser, err error) {
			if err != nil || wc == nil {
				return
			}
			defer wc.Close()
			_, _ = wc.Write([]byte(logPanel.Text()))
		}, w)
	})
	clearLogBtn := widget.NewButton("清空", func() { logPanel.Clear() })

	logHeaderBg := trackRect(theme.ColorNameMenuBackground)
	logHeader := container.NewStack(logHeaderBg, container.NewPadded(container.NewHBox(
		widget.NewLabelWithStyle("运行日志 RUNTIME LOG", fyne.TextAlignLeading, fyne.TextStyle{}),
		layout.NewSpacer(),
		autoScrollCheck, copyLogBtn, exportLogBtn, clearLogBtn,
	)))
	logBg := canvas.NewRectangle(LogBackground)
	logArea := container.NewStack(logBg, logPanel.CanvasObject())
	logSection := container.NewBorder(logHeader, nil, nil, nil, logArea)

	// ---- 状态栏 ----
	statusBarBg := trackRect(theme.ColorNameMenuBackground)
	statusBar := container.NewStack(statusBarBg, container.NewPadded(container.NewHBox(
		widget.NewLabel("总数"), totalLabel,
		widget.NewLabel("成功"), successLabel,
		widget.NewLabel("失败"), failLabel,
		widget.NewLabel("待发"), pendingLabel,
		layout.NewSpacer(),
		widget.NewLabel("已用时"), elapsedLabel,
	)))

	// ---- 发布页 / 帮助页 ----
	toolbar := container.NewHBox(addFolderBtn, addFileBtn, clearBtn, layout.NewSpacer(), saveBtn)
	publishView := container.NewBorder(container.NewVBox(toolbar, settingsCard), nil, nil, nil, table)
	helpView := BuildHelp()
	contentStack := container.NewStack(publishView, helpView)
	helpView.Hide()

	navPublish := widget.NewButton("发布", nil)
	navHelp := widget.NewButton("帮助", nil)
	selectNav := func(active *widget.Button) {
		for _, b := range []*widget.Button{navPublish, navHelp} {
			if b == active {
				b.Importance = widget.HighImportance
			} else {
				b.Importance = widget.LowImportance
			}
			b.Refresh()
		}
	}
	navPublish.OnTapped = func() {
		helpView.Hide()
		publishView.Show()
		selectNav(navPublish)
	}
	navHelp.OnTapped = func() {
		publishView.Hide()
		helpView.Show()
		selectNav(navHelp)
	}
	selectNav(navPublish)

	navBg := trackRect(theme.ColorNameMenuBackground)
	nav := container.NewStack(navBg, container.NewPadded(container.NewVBox(navPublish, navHelp)))

	// ---- 右侧参数栏 ----
	rightBg := trackRect(theme.ColorNameMenuBackground)
	right := container.NewStack(rightBg, container.NewPadded(container.NewVBox(
		widget.NewLabelWithStyle("运行参数", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewGridWithColumns(2, widget.NewLabel("发布间隔"), intervalSummary),
		container.NewGridWithColumns(2, widget.NewLabel("失败重试"), retriesSummary),
		container.NewGridWithColumns(2, widget.NewLabel("队列篇数"), queueSummary),
		widget.NewSeparator(),
		startBtn, stopBtn,
		widget.NewSeparator(),
		saveBtn, clearBtn, exportBtn,
	)))

	// ---- 标题栏 ----
	iconBg := canvas.NewRectangle(currentTheme().accentColor())
	iconBg.CornerRadius = 6
	iconText := canvas.NewText("G", color.White)
	iconText.TextStyle = fyne.TextStyle{Bold: true}
	iconText.Alignment = fyne.TextAlignCenter
	iconBox := container.New(layout.NewGridWrapLayout(fyne.NewSize(24, 24)), container.NewStack(iconBg, container.NewCenter(iconText)))
	appName := widget.NewLabelWithStyle("GitHub 文章发布器", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	var themeToggleBtn *widget.Button
	repaintChrome := func() {
		for _, c := range chrome {
			c.rect.FillColor = currentTheme().Color(c.name, currentTheme().Variant)
			c.rect.Refresh()
		}
		iconBg.FillColor = currentTheme().accentColor()
		iconBg.Refresh()
		table.Refresh()

		total, done, failed := len(arts), 0, 0
		if rowState != nil {
			done = rowState.Done()
			for i := 0; i < total; i++ {
				if strings.HasPrefix(rowState.Status(i), "失败") {
					failed++
				}
			}
		}
		pill.repaint(isRunning, done, total, failed)

		if currentTheme().Variant == theme.VariantDark {
			themeToggleBtn.SetText("☀ 浅色")
		} else {
			themeToggleBtn.SetText("🌙 深色")
		}
	}

	themeToggleBtn = widget.NewButton("", func() {
		next := theme.VariantLight
		if currentTheme().Variant == theme.VariantLight {
			next = theme.VariantDark
		}
		variantName := "light"
		if next == theme.VariantDark {
			variantName = "dark"
		}
		prefs.SetString(keyThemeVariant, variantName)
		fyne.CurrentApp().Settings().SetTheme(NewTheme(next))
	})
	fyne.CurrentApp().Settings().AddListener(func(fyne.Settings) { repaintChrome() })
	repaintChrome()

	titleBarBg := trackRect(theme.ColorNameMenuBackground)
	titleBar := container.NewStack(titleBarBg, container.NewPadded(container.NewHBox(
		iconBox, appName, layout.NewSpacer(), pill.object(), themeToggleBtn,
	)))

	refreshRightPanel()
	refreshCounts()

	body := container.NewBorder(nil, nil, nav, right, contentStack)
	return container.NewBorder(titleBar, container.NewVBox(logSection, statusBar), nil, nil, body)
}
