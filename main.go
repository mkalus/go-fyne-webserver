package main

import (
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	startWebServerText = "Start Web Server"
	stopWebServerText  = "Stop Web Server"
)

func main() {
	// get the current directory
	workingDirectory, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	workingDirectory = filepath.Join(workingDirectory, "static")

	// create static file server which will be used to serve the content
	fileServer := http.FileServer(http.Dir("./static"))
	http.Handle("/", fileServer)

	// create http server reference
	var server *http.Server

	// create Fyne Window
	a := app.New()
	a.SetIcon(resourceIconPng)
	w := a.NewWindow("Go Fyne Webserver")
	w.SetMaster()
	w.CenterOnScreen()

	// explaining text
	text := widget.NewRichTextFromMarkdown("Serving files from\n\n`" + workingDirectory + "`\n\nOnce server runs, click here to open it:\n\n[http://localhost:8080/](http://localhost:8080/)\n\n")

	// button - main functionality here...
	var button *widget.Button
	button = widget.NewButtonWithIcon(startWebServerText, theme.MediaPlayIcon(), func() {
		if server != nil {
			// server is running? shut it down
			button.SetText(startWebServerText)
			button.SetIcon(theme.MediaPlayIcon())
			if err := server.Close(); err != nil {
				log.Fatal(err)
			}
			server = nil
		} else {
			// not running - start go function to start it
			go func() {
				button.SetText(stopWebServerText)
				button.SetIcon(theme.MediaStopIcon())
				server = &http.Server{Addr: ":8080", Handler: nil}
				if err := server.ListenAndServe(); err != nil {
					if !strings.Contains(err.Error(), "Server closed") {
						log.Fatal(err)
					}
				}
			}()
		}
	})

	// Set content and run
	content := container.NewVBox(text, button)
	w.SetContent(content)

	w.ShowAndRun()
}
