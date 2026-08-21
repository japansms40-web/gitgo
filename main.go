package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"githubbaidu/internal/ui"
)

func main() {
	a := app.NewWithID("com.ghpublisher.app")
	w := a.NewWindow("GitHub 文章发布器")
	w.SetContent(ui.Build(w, a.Preferences()))
	w.Resize(fyne.NewSize(760, 640))
	w.ShowAndRun()
}
