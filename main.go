package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	startWebServerText = "Start Web Server"
	stopWebServerText  = "Stop Web Server"
)

var bp = "localhost:8080"

func main() {
	// get initial working directory
	wd, err := getInitialWorkingDirectory()
	if err != nil {
		log.Fatal(err)
	}

	// bound variables
	workingDirectory := binding.BindString(&wd)
	boundPort := binding.BindString(&bp)

	// content size of window
	var contentSize fyne.Size

	// create http server reference
	var server *http.Server

	// create Fyne Window
	a := app.New()
	a.SetIcon(resourceIconPng)
	w := a.NewWindow("Go Fyne Webserver")
	w.SetMaster()
	w.CenterOnScreen()

	// create widgets
	lbl1 := widget.NewLabel("Serving files from")
	lbl2 := widget.NewLabelWithData(workingDirectory)
	lbl2.TextStyle = fyne.TextStyle{Monospace: true}
	lbl3 := widget.NewLabel("Open your browser from:")
	lbl3.Hide()
	lbl4 := newTappableLabelWithData(boundPort)
	lbl4.TextStyle = fyne.TextStyle{Monospace: true}
	lbl4.Hide()

	// button - main functionality here...
	var runServerButton *widget.Button
	var changeDirButton *widget.Button
	runServerButton = widget.NewButtonWithIcon(startWebServerText, theme.MediaPlayIcon(), func() {
		if server != nil {
			// server is running? shut it down
			if err := server.Close(); err != nil {
				log.Fatal(err)
			}
			server = nil

			// update button
			runServerButton.SetText(startWebServerText)
			runServerButton.SetIcon(theme.MediaPlayIcon())

			// enable change dir
			changeDirButton.Enable()

			// hide info
			lbl3.Hide()
			lbl4.Hide()
		} else {
			// get port
			port := getOpenPort()
			if port == "" {
				log.Fatalf("Could not open any port in the range 8080 to 9090.")
			}
			port = "localhost:" + port
			_ = boundPort.Set(port)

			// create static file server which will be used to serve the content
			mux := http.NewServeMux()

			workDir, _ := workingDirectory.Get()
			fileServer := http.FileServer(http.Dir(workDir))
			mux.Handle("/", fileServer)
			server = &http.Server{Addr: port, Handler: mux}

			// serve in function
			go func() {
				if err := server.ListenAndServe(); err != nil {
					if !strings.Contains(err.Error(), "Server closed") {
						log.Fatal(err)
					}
				}
			}()

			// update button
			runServerButton.SetText(stopWebServerText)
			runServerButton.SetIcon(theme.MediaStopIcon())

			// disable change dir
			changeDirButton.Disable()

			// show info
			lbl3.Show()
			lbl4.Show()
		}
	})

	// change directory button
	changeDirButton = widget.NewButton("Change Directory", func() {
		d := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
			if err == nil && uri != nil {
				workingDirectory.Set(uri.Path())
			}
			w.Resize(contentSize)
		}, w)
		d.Show()
		workDir, _ := workingDirectory.Get()
		uri, _ := storage.ParseURI("file://" + workDir)
		listableUri, _ := storage.ListerForURI(uri)
		d.SetLocation(listableUri)
		w.Resize(d.MinSize().AddWidthHeight(300, 200))
	})

	// Set content and run
	content := container.NewVBox(lbl1, lbl2, lbl3, lbl4, runServerButton, changeDirButton)
	w.SetContent(content)
	contentSize = content.Size()

	w.ShowAndRun()
}

// getInitialWorkingDirectory will return the initial directory to serve files from - or an error
func getInitialWorkingDirectory() (workingDirectory string, err error) {
	// get the current directory
	workingDirectory, err = os.Getwd()
	if err == nil {
		workingDirectory = guessPossibleSubdirectories(workingDirectory)
	}

	return
}

// guessPossibleSubdirectories will try to find a subpath of our base which makes sense
func guessPossibleSubdirectories(baseDirectory string) string {
	for _, subDir := range []string{
		filepath.Join(".output", "public"), // nuxt 3
		".output",
		"dist", // vue et al
		"static",
	} {
		if fi, err := os.Stat(filepath.Join(baseDirectory, subDir)); err == nil && fi.IsDir() {
			return filepath.Join(baseDirectory, subDir)
		}
	}
	return baseDirectory
}

// getOpenPort will return an open port in the range 8080 to 9090
func getOpenPort() string {
	for port := 8080; port <= 9090; port++ {
		strPort := strconv.Itoa(port)
		ln, err := net.Listen("tcp", ":"+strPort)
		if ln != nil {
			_ = ln.Close()
		}
		if err == nil {
			return strPort
		}
	}
	return ""
}

type tappableLabel struct {
	widget.Label
}

func newTappableLabelWithData(data binding.String) *tappableLabel {
	label := &tappableLabel{}
	label.ExtendBaseWidget(label)
	label.SetText("")
	label.Bind(data)

	return label
}

func (t *tappableLabel) Tapped(_ *fyne.PointEvent) {
	_ = exec.Command("open", "http://"+bp).Run()
}

func (b *tappableLabel) Cursor() desktop.Cursor {
	return desktop.PointerCursor
}
