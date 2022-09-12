package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"image/color"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Config struct {
	Name string
	File string
}

func main() {
	application := app.New()
	window := application.NewWindow("Wireguard GUI (deblan)")
	configs := []Config{}
	directory := "/etc/wireguard/"

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

	if err != nil {
		log.Fatalln(err)
	}

	menu := container.NewAppTabs()
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

	content := container.New(layout.NewVBoxLayout(), menu)

	window.SetContent(content)
	window.Resize(fyne.NewSize(900, 400))
	window.ShowAndRun()
}

func WgUp(config Config, notice *canvas.Text) {
	notice.Hidden = false
	notice.Text = fmt.Sprintf("Interface is starting")
	notice.Color = color.RGBA{R: 200, G: 165, B: 76}
	notice.Refresh()

	exec.Command("wg-quick", "up", config.Name).Output()

	notice.Text = fmt.Sprintf("Interface is up")
	notice.Color = color.RGBA{R: 111, G: 200, B: 72}
	notice.Refresh()

	go func() {
		time.Sleep(2 * time.Second)
		notice.Hidden = true
		notice.Refresh()
	}()
}

func WgDown(config Config, notice *canvas.Text) {
	notice.Refresh()

	notice.Hidden = false
	notice.Text = fmt.Sprintf("Interface is stopping")
	notice.Color = color.RGBA{R: 200, G: 165, B: 76}
	notice.Refresh()

	exec.Command("wg-quick", "down", config.Name).Output()

	notice.Text = fmt.Sprintf("Interface is down")
	notice.Color = color.RGBA{R: 111, G: 200, B: 72}
	notice.Refresh()

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

func createTab(config Config) *fyne.Container {
	notice := canvas.NewText("", color.White)
	notice.TextStyle.Bold = true

	r1 := canvas.NewText("", color.Transparent)
	r2 := canvas.NewText("", color.Transparent)
	r1.TextSize = 5
	r2.TextSize = 5

	data, err := os.ReadFile(config.File)

	if err != nil {
		log.Fatalln(err)
	}

	buttonStart := widget.NewButton("Start", func() {
		WgUp(config, notice)
	})

	buttonStop := widget.NewButton("Stop", func() {
		WgDown(config, notice)
	})

	buttonRestart := widget.NewButton("Restart", func() {
		WgRestart(config, notice)
	})

	top := container.New(
		layout.NewVBoxLayout(),
		r1,
		container.New(
			layout.NewHBoxLayout(),
			buttonStart,
			buttonStop,
			buttonRestart,
			notice,
		),
		r2,
	)

	configuration := string(data)

	textareaConfiguration := widget.NewMultiLineEntry()
	textareaConfiguration.SetText(configuration)
	textareaConfiguration.SetMinRowsVisible(strings.Count(configuration, "\n"))

	textareaConfiguration.OnChanged = func(text string) {
		textareaConfiguration.SetMinRowsVisible(strings.Count(text, "\n"))
		textareaConfiguration.Refresh()
	}

	form := &widget.Form{
		Items: []*widget.FormItem{
			{
				Text:   "Configuration",
				Widget: textareaConfiguration,
			},
		},
		OnSubmit: func() {
			notice.Hidden = false
			configuration := fmt.Sprintf("%s\n", textareaConfiguration.Text)
			err := os.WriteFile(config.File, []byte(configuration), 600)

			if err != nil {
				log.Println(err)
				notice.Text = fmt.Sprintf("Error while updating: %s", err)
				notice.Color = color.RGBA{R: 200, G: 68, B: 68}
				notice.Refresh()
			} else {
				notice.Text = fmt.Sprintf("Configuration updated")
				notice.Color = color.RGBA{R: 111, G: 200, B: 72}

				go func() {
					time.Sleep(2 * time.Second)
					notice.Hidden = true
					notice.Refresh()
				}()
			}

			notice.Refresh()
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
