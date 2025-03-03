package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Software - API dan keladigan dastur ma'lumotlari
type Software struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	MainFile    string `json:"mainFile"`
	Icon        string `json:"icon"`
}

type ResponseData struct {
	Message string     `json:"message"`
	Object  []Software `json:"object"`
}

var downloadPath = "C:/Downloads" // Default yuklab olish papkasi

func fetchAPIData(url string) ([]Software, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data ResponseData
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, err
	}

	return data.Object, nil
}

// Faylni yuklab olish
func downloadFile(url, filePath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func main() {
	myApp := app.NewWithID("men-go-fyne")
	myWindow := myApp.NewWindow("Men Go Fyne")
	myWindow.Resize(fyne.NewSize(800, 500))

	label := widget.NewLabel("Ma'lumot yuklanmoqda...")
	descriptionLabel := widget.NewLabel("") // Pastki tavsif maydoni

	softwares, err := fetchAPIData("http://localhost:8080/appStore/getAllSoftware")
	if err != nil {
		label.SetText(fmt.Sprintf("Xatolik: %v", err))
		myWindow.SetContent(container.NewVBox(label))
		myWindow.ShowAndRun()
		return
	}

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Dastur nomi yoki tavsif bo'yicha qidirish...")

	contentContainer := container.NewGridWrap(fyne.NewSize(150, 150))
	updateSoftwareList := func(query string) {
		contentContainer.Objects = nil
		var filteredSoftwares []fyne.CanvasObject
		query = strings.ToLower(query)
		for _, software := range softwares {
			if query == "" || strings.Contains(strings.ToLower(software.Name), query) || strings.Contains(strings.ToLower(software.Description), query) {
				card := createSoftwareCard(software, descriptionLabel, myWindow)
				filteredSoftwares = append(filteredSoftwares, card)
			}
		}
		contentContainer.Objects = filteredSoftwares
		contentContainer.Refresh()
	}

	updateSoftwareList("")

	searchEntry.OnChanged = func(query string) {
		updateSoftwareList(query)
	}

	mainContainer := container.NewVBox(
		container.NewPadded(searchEntry),
		contentContainer,
		widget.NewSeparator(),
		descriptionLabel, // Tavsif ekranning pastki qismida chiqadi
	)

	myWindow.SetContent(mainContainer)
	myWindow.ShowAndRun()
}

func createSoftwareCard(software Software, descriptionLabel *widget.Label, myWindow fyne.Window) fyne.CanvasObject {
	title := widget.NewLabelWithStyle(software.Name, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	var iconImage *canvas.Image
	if software.Icon != "" {
		decoded, err := base64.StdEncoding.DecodeString(software.Icon)
		if err != nil {
			iconImage = canvas.NewImageFromResource(theme.ErrorIcon())
		} else {
			img, _, err := image.Decode(bytes.NewReader(decoded))
			if err != nil {
				iconImage = canvas.NewImageFromResource(theme.ErrorIcon())
			} else {
				iconImage = canvas.NewImageFromImage(img)
				iconImage.FillMode = canvas.ImageFillContain
				iconImage.SetMinSize(fyne.NewSize(80, 80))
			}
		}
	}
	if iconImage == nil {
		iconImage = canvas.NewImageFromResource(theme.FileIcon())
	}

	// Tugmalar
	downloadButton := widget.NewButtonWithIcon("", theme.DownloadIcon(), nil)
	downloadButton.Importance = widget.LowImportance

	progressBar := widget.NewProgressBarInfinite()
	progressBar.Hide()

	updateButton := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), nil)
	updateButton.Hide()

	infoButton := widget.NewButtonWithIcon("", theme.InfoIcon(), func() {
		descriptionLabel.SetText("📌 " + software.Description)
	})
	infoButton.Importance = widget.LowImportance

	version := widget.NewLabelWithStyle("v: "+software.Version, fyne.TextAlignTrailing, fyne.TextStyle{Bold: true, Italic: true})

	downloadButton.OnTapped = func() {
		dialog.ShowFolderOpen(func(folder fyne.ListableURI, err error) {
			if err != nil || folder == nil {
				return
			}

			downloadPath = folder.Path()
			fileURL := fmt.Sprintf("http://localhost:8080/appStore/download/%s", software.ID)
			filePath := filepath.Join(downloadPath, software.MainFile)

			// Yuklanish indikatorini ko'rsatish
			downloadButton.Hide()
			progressBar.Show()

			go func() {
				err := downloadFile(fileURL, filePath)

				if err != nil {
					fmt.Println("Xatolik yuz berdi:", err)
					return
				}

				fyne.CurrentApp().SendNotification(fyne.NewNotification("Yuklandi", filePath))

				// UI yangilash
				fyne.CurrentApp().Driver().CanvasForObject(downloadButton).Refresh(downloadButton)

				progressBar.Hide()
				updateButton.Show()
			}()
		}, myWindow)
	}

	versionContainer := container.NewHBox(
		infoButton,
		layout.NewSpacer(),
		version,
	)

	buttonContainer := container.NewHBox(
		progressBar,
		downloadButton,
		updateButton,
	)

	titleContainer := container.NewBorder(nil, nil, nil, buttonContainer, title)
	content := container.NewVBox(
		titleContainer,
		container.NewCenter(iconImage),
		versionContainer,
	)

	border := widget.NewSeparator()
	card := container.NewStack(border, content)

	return card
}
