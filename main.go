package main

import (
	"fmt"
	"image/color"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

var (
	red         color.RGBA
	green       color.RGBA
	orange      color.RGBA
	directory   string
	configs     []Config
	application fyne.App
	window      fyne.Window
	menu        *container.AppTabs
)

type Config struct {
	Name string
	File string
}

type ButtonCallback func()

func main() {
	red = color.RGBA{R: 156, G: 21, B: 21}
	green = color.RGBA{R: 47, G: 156, B: 17}
	orange = color.RGBA{R: 156, G: 106, B: 25}

	application = app.New()
	window = application.NewWindow("Wireguard GUI")
	directory = "/etc/wireguard/"
	menu = container.NewAppTabs()

	err := initConfigs()
	if err != nil {
		log.Fatalln(err)
	}

	initMenu()

	content := container.New(layout.NewVBoxLayout(), menu)

	window.SetContent(content)
	window.Resize(fyne.NewSize(900, 400))
	window.ShowAndRun()
}

func initMenu() {
	tabs := make([]fyne.Container, len(configs))

	for i, config := range configs {
		tabs[i] = *createTab(config)
	}

	for i, config := range configs {
		menu.Append(
			container.NewTabItem(
				config.Name,
				&tabs[i],
			),
		)
	}
}

func initConfigs() error {
	err := filepath.WalkDir(directory, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		basename := string(info.Name())

		if !strings.HasSuffix(basename, ".conf") {
			return nil
		}

		if strings.Contains(strings.ReplaceAll(path, directory, ""), "/") {
			return nil
		}

		configs = append(configs, Config{
			Name: strings.ReplaceAll(basename, ".conf", ""),
			File: path,
		})

		return nil
	})

	return err
}

func toggleNotice(notice *canvas.Text, isVisible bool) {
	notice.Hidden = !isVisible
	notice.Refresh()
}

func updateNotice(notice *canvas.Text, text string, c color.Color, isVisible, isFlash bool) {
	notice.Text = text
	notice.Color = c
	notice.Hidden = !isVisible
	notice.Refresh()

	if isFlash {
		go func() {
			time.Sleep(2 * time.Second)
			toggleNotice(notice, false)
		}()
	}

	log.Println(text)
}

func WgUp(config Config, notice *canvas.Text) {
	updateNotice(notice, fmt.Sprintf("Interface is starting"), orange, true, false)
	exec.Command("wg-quick", "up", config.Name).Output()
	updateNotice(notice, fmt.Sprintf("Interface is up"), green, true, true)
}

func WgDown(config Config, notice *canvas.Text) {
	updateNotice(notice, fmt.Sprintf("Interface is stopping"), orange, true, false)
	exec.Command("wg-quick", "down", config.Name).Output()
	updateNotice(notice, fmt.Sprintf("Interface is down"), green, true, true)

	go func() {
		time.Sleep(2 * time.Second)
		notice.Hidden = true
		notice.Refresh()
	}()
}

func WgRestart(config Config, notice *canvas.Text) {
	WgDown(config, notice)
	WgUp(config, notice)
}

func lintConfiguration(configuration string) string {
	configuration = strings.TrimSpace(configuration)
	configuration = fmt.Sprintf("%s\n", configuration)

	return configuration
}

func updateTextareaConfiguration(textarea *widget.Entry, content string) {
	textarea.SetText(content)
	textarea.OnChanged(content)
}

func updateConfigFile(config Config, content string) error {
	return os.WriteFile(config.File, []byte(content), 600)
}

func createTextarea() *widget.Entry {
	textarea := widget.NewMultiLineEntry()
	textarea.OnChanged = func(text string) {
		textarea.SetMinRowsVisible(strings.Count(text, "\n"))
		textarea.Refresh()
	}

	return textarea
}

func createColoredButton(label string, c color.Color, callback ButtonCallback) *fyne.Container {
	return container.NewMax(
		canvas.NewRectangle(c),
		widget.NewButton(label, callback),
	)
}

func createNotice() *canvas.Text {
	notice := canvas.NewText("", color.White)
	notice.TextStyle.Bold = true

	return notice
}

func createMargin() *canvas.Text {
	text := canvas.NewText("", color.Transparent)
	text.TextSize = 5

	return text
}

func createTab(config Config) *fyne.Container {
	data, err := os.ReadFile(config.File)
	if err != nil {
		log.Fatalln(err)
	}

	notice := createNotice()

	buttonStart := createColoredButton("Start", green, func() {
		WgUp(config, notice)
	})

	buttonStop := createColoredButton("Stop", red, func() {
		WgDown(config, notice)
	})

	buttonRestart := createColoredButton("Restart", orange, func() {
		WgRestart(config, notice)
	})

	top := container.New(
		layout.NewVBoxLayout(),
		createMargin(),
		container.New(
			layout.NewHBoxLayout(),
			notice,
			layout.NewSpacer(),
			buttonStart,
			buttonStop,
			buttonRestart,
		),
		createMargin(),
	)

	textareaConfiguration := createTextarea()
	updateTextareaConfiguration(textareaConfiguration, string(data))

	form := &widget.Form{
		Items: []*widget.FormItem{
			{
				Text:   "Configuration",
				Widget: textareaConfiguration,
			},
		},
		OnSubmit: func() {
			toggleNotice(notice, false)

			configuration := lintConfiguration(textareaConfiguration.Text)
			updateTextareaConfiguration(textareaConfiguration, configuration)
			err := updateConfigFile(config, configuration)

			if err != nil {
				updateNotice(notice, fmt.Sprintf("Error while updating: %s", err), red, true, false)
			} else {
				updateNotice(notice, fmt.Sprintf("Configuration updated"), green, true, true)
			}
		},
		SubmitText: "Save",
	}

	content := container.New(
		layout.NewVBoxLayout(),
		top,
		form,
	)

	return content
}
