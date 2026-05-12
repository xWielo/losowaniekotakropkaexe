package main

import (
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var imageExts = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
}

func listCats() []string {
	entries, err := os.ReadDir("koty")
	if err != nil {
		return nil
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() && imageExts[strings.ToLower(filepath.Ext(e.Name()))] {
			files = append(files, filepath.Join("koty", e.Name()))
		}
	}
	return files
}

func main() {
	a := app.New()
	w := a.NewWindow("Losowanie Kota")
	w.Resize(fyne.NewSize(660, 580))

	cats := listCats()

	imgCanvas := canvas.NewImageFromFile("")
	imgCanvas.FillMode = canvas.ImageFillContain
	imgCanvas.SetMinSize(fyne.NewSize(620, 440))

	statusLabel := widget.NewLabelWithStyle(
		"Kliknij przycisk, żeby wylosować kota!",
		fyne.TextAlignCenter,
		fyne.TextStyle{},
	)

	var btn *widget.Button
	btn = widget.NewButtonWithIcon("Losuj kota!", theme.MediaPlayIcon(), func() {
		if len(cats) == 0 {
			statusLabel.SetText("⚠  Brak zdjęć w folderze koty/")
			return
		}
		btn.Disable()
		statusLabel.SetText("Losuję...")

		go func() {
			time.Sleep(time.Second)
			picked := cats[rand.Intn(len(cats))]
			imgCanvas.File = picked
			imgCanvas.Refresh()
			statusLabel.SetText("🎉  " + filepath.Base(picked))
			btn.Enable()
		}()
	})
	btn.Importance = widget.HighImportance

	if len(cats) == 0 {
		statusLabel.SetText("⚠  Brak zdjęć w folderze koty/")
		btn.Disable()
	}

	bottom := container.NewVBox(
		statusLabel,
		container.New(layout.NewCenterLayout(), btn),
	)

	content := container.NewBorder(nil, bottom, nil, nil, imgCanvas)
	w.SetContent(content)
	w.ShowAndRun()
}
