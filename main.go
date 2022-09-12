package main

import (
	"flag"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"image/color"
	"log"
	"os"
	"strings"
)

func main() {
	application := app.New()
	window := application.NewWindow("Wireguard GUI (deblan)")

	iface := flag.String("iface", "wg0", "wireguard iface")
	filename := flag.String("filename", "/etc/wireguard/wg0.conf", "wireguard iface")
	flag.Parse()

	data, err := os.ReadFile(*filename)

	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}
	_ = iface

	configuration := string(data)

	textareaConfiguration := widget.NewMultiLineEntry()
	textareaConfiguration.SetText(configuration)
	textareaConfiguration.SetMinRowsVisible(strings.Count(configuration, "\n"))

	notice := canvas.NewText("", color.White)

	form := &widget.Form{
		Items: []*widget.FormItem{
			{
				Text:   "Configuration",
				Widget: textareaConfiguration,
			},
		},
		OnSubmit: func() {
			notice.Text = ""
			notice.Refresh()

			configuration := textareaConfiguration.Text
			err := os.WriteFile(*filename, []byte(configuration), 600)

			if err != nil {
				log.Println(err)
				notice.Text = fmt.Sprintf("Error: %s", err)
				notice.Color = color.RGBA{R: 255}
				notice.Refresh()
			} else {
				notice.Text = "Configuration updated!"
				notice.Color = color.RGBA{G: 255}
			}

			notice.Refresh()
		},
		SubmitText: "Save",
	}

	content := container.New(layout.NewVBoxLayout(), form, notice)

	window.SetContent(content)
	window.Resize(fyne.NewSize(900, 400))
	window.ShowAndRun()
}
