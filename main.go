package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

const SEARCH_LOCATION string = "PATH"

func main() {
	myApp := app.NewWithID("55a15cad-0bbe-4536-9ff0-cad5f9648a72_SearchBox")

	window := myApp.NewWindow("Run Box")
	window.Resize(fyne.NewSize(500, 200))
	window.SetFixedSize(true)
	window.SetMaster()

	text := widget.NewLabel("Search Here:")
	text.TextStyle.Monospace = true
	text.TextStyle = fyne.TextStyle{Italic: true}

	textBox := widget.NewEntry()
	textBox.SetPlaceHolder("Keyword ...")
	textBox.MultiLine = false

	allFiles := loadFiles(myApp)

	filteredFiles := []string{}

	list := widget.NewList(func() int {
		return len(filteredFiles)
	}, func() fyne.CanvasObject {
		return widget.NewLabel("")
	}, func(lii widget.ListItemID, co fyne.CanvasObject) {
		// co.(*widget.Lable) = type assertion
		co.(*widget.Label).SetText(filteredFiles[lii])
	})

	textBox.OnChanged = func(s string) {
		files := []string{}
		if strings.Compare(s, "") == 0 {
			filteredFiles = []string{}
			list.Refresh()
		} else {
			for _, fileName := range allFiles {
				if strings.Contains(fileName, s) {
					files = append(files, fileName)
				}
			}
			filteredFiles = files
		}
		list.Refresh()
	}

	textBox.OnSubmitted = func(s string) {
		if len(filteredFiles) > 0 {
			openSelectedFile(myApp, filteredFiles[0])
		}
	}

	img, err := fyne.LoadResourceFromPath("./icons/settings_white_24dp.svg")
	if err != nil {
		handleError(myApp, err.Error(), func() {
			os.Exit(1)
		})
	}

	gridLayout := container.NewGridWithRows(2,
		container.NewVBox(
			container.NewBorder(
				nil,
				nil,
				text,
				widget.NewButtonWithIcon("", img, func() {
					openPrefWindow(myApp)
				}),
			),
			textBox),
		container.NewMax(list),
	)
	mainContainer := container.New(layout.NewPaddedLayout(), gridLayout)
	window.SetContent(mainContainer)

	window.Show()
	myApp.Run()
}

func openSelectedFile(myApp fyne.App, fileName string) {
	location := myApp.Preferences().StringWithFallback(SEARCH_LOCATION, "nil")
	if location == "nil" {
		handleError(myApp, "Search Folder path is empty.", nil)
	} else {
		path, err := exec.LookPath(location + "/" + fileName)
		if err != nil {
			handleError(myApp, err.Error(), nil)
		}
		err = exec.Command(path).Start()
		if err != nil {
			handleError(myApp, err.Error(), nil)
		} else {
			os.Exit(0)
		}
	}
}

func loadFiles(myApp fyne.App) []string {
	files := []string{}

	location := myApp.Preferences().StringWithFallback(SEARCH_LOCATION, "nil")
	if location == "nil" {
		handleError(myApp, "Search location is empty", nil)
		return []string{}
	}

	f, err := os.Open(location)
	if err != nil {
		handleError(myApp, err.Error(), nil)
	}

	fileInfo, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		handleError(myApp, err.Error(), func() {
			os.Exit(-1)
		})
	}

	for _, file := range fileInfo {
		if !file.IsDir() {
			files = append(files, file.Name())
		}
	}
	return files
}

func openPrefWindow(myApp fyne.App) {
	searchTextbox := widget.NewEntry()
	searchTextbox.MultiLine = false
	searchTextbox.SetPlaceHolder("Path..")

	// add saved location.
	searchTextbox.SetText(myApp.Preferences().StringWithFallback(SEARCH_LOCATION, ""))

	prefWindow := myApp.NewWindow("Settings")
	prefWindow.SetContent(
		container.NewPadded(
			container.NewGridWithRows(
				3,
				widget.NewLabel("Settings"),
				container.New(
					layout.NewFormLayout(),
					widget.NewLabel("search location:"),
					searchTextbox,
				),
				container.NewBorder(nil, nil, nil, widget.NewButton("Save", func() {
					_, err := os.Stat(searchTextbox.Text)
					if os.IsNotExist(err) {
						handleError(myApp, err.Error(), nil)
					} else {
						myApp.Preferences().SetString(SEARCH_LOCATION, searchTextbox.Text)
						prefWindow.Close()
					}
				})),
			),
		),
	)
	prefWindow.Resize(fyne.Size{Width: 500})
	prefWindow.SetFixedSize(true)
	prefWindow.Show()
}

func handleError(app fyne.App, message string, onClose func()) {
	img, err := fyne.LoadResourceFromPath("./icons/error_white_24dp.svg")
	if err != nil {
		fmt.Println("Could not load settings image")
		os.Exit(1)
	}

	icon := widget.NewIcon(img)

	errWindow := app.NewWindow("Error")
	errWindow.SetPadded(true)
	errWindow.SetFixedSize(true)
	errWindow.RequestFocus()
	errWindow.CenterOnScreen()
	errWindow.SetContent(
		container.NewGridWithRows(
			3,
			layout.NewSpacer(),
			container.NewVBox(
				icon,
				layout.NewSpacer(),
				widget.NewTextGridFromString(message),
			),
			layout.NewSpacer(),
		),
	)
	if onClose != nil {
		errWindow.SetOnClosed(onClose)
	}
	errWindow.Show()
}
