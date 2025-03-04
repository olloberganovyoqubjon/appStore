package ui

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"
	"time"

	"main/models"
	"main/services"
	"main/storage"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func SetupUI(myWindow fyne.Window) {
	// Yuklanish xabari uchun yorliq
	label := widget.NewLabel("Ma'lumot yuklanmoqda...")

	// Dastur tavsifi uchun bo'sh yorliq
	descriptionLabel := widget.NewLabel("")

	// API dan dasturlarni olish
	softwares, err := services.FetchAPIData("http://localhost:8080/appStore/getAllSoftware")
	if err != nil { // Agar xatolik bo'lsa
		// Xatolik xabarini yorliqqa yozish
		label.SetText(fmt.Sprintf("Xatolik: %v", err))
		// Oynaga faqat xatolik xabarini joylashtirish
		myWindow.SetContent(container.NewVBox(label))
		return // Funksiyadan chiqish
	}

	// Qidiruv maydoni yaratish
	searchEntry := widget.NewEntry()
	// Placeholder matn qo'shish
	searchEntry.SetPlaceHolder("Dastur nomi yoki tavsif bo'yicha qidirish...")

	// Dastur kartalari uchun grid konteyner (150x150 o'lchamli)
	contentContainer := container.NewGridWrap(fyne.NewSize(200, 150))

	// Dastur ro'yxatini yangilash funksiyasi
	updateSoftwareList := func(query string) {
		// Avvalgi ob'ektlarni tozalash
		contentContainer.Objects = nil
		// Filtlangan dasturlarni saqlash uchun massiv
		var filteredSoftwares []fyne.CanvasObject
		// Qidiruv so'zini kichik harflarga aylantirish
		query = strings.ToLower(query)
		// Har bir dastur uchun tsikl
		for _, software := range softwares {
			// Agar qidiruv bo'sh yoki nom/tavsifda so'z bo'lsa
			if query == "" || strings.Contains(strings.ToLower(software.Name), query) || strings.Contains(strings.ToLower(software.Description), query) {
				// Dastur kartasini yaratish
				card := createSoftwareCard(software, descriptionLabel, myWindow)
				// Kartani ro'yxatga qo'shish
				filteredSoftwares = append(filteredSoftwares, card)
			}
		}
		// Yangi kartalarni konteynerga qo'shish
		contentContainer.Objects = filteredSoftwares
		// Konteynerni yangilash
		contentContainer.Refresh()
	}

	// Boshlang'ich ro'yxatni ko'rsatish (qidiruvsiz)
	updateSoftwareList("")

	// Qidiruv maydonida matn o'zgarsa
	searchEntry.OnChanged = func(query string) {
		// Ro'yxatni yangilash
		updateSoftwareList(query)
	}

	// Asosiy vertikal konteyner
	mainContainer := container.NewVBox(
		// Qidiruv maydonini padding bilan qo'shish
		container.NewPadded(searchEntry),
		// Dastur kartalari konteyneri
		contentContainer,
		// Ajratuvchi chiziq
		widget.NewSeparator(),
		// Tavsif yorlig'i
		descriptionLabel,
	)

	// Oynaga asosiy konteynerni joylashtirish
	myWindow.SetContent(mainContainer)
}

// Dastur kartasini yaratish funksiyasi (yangilangan)
func createSoftwareCard(software models.Software, descriptionLabel *widget.Label, myWindow fyne.Window) fyne.CanvasObject {
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

	// Tugmalar va progress bar
	deleteButton := widget.NewButtonWithIcon("", theme.DeleteIcon(), nil) // Yangi "Delete" tugmasi
	deleteButton.Importance = widget.LowImportance
	deleteButton.Hide()

	downloadButton := widget.NewButtonWithIcon("", theme.DownloadIcon(), nil)
	downloadButton.Importance = widget.LowImportance
	downloadButton.Hide()

	progressBar := widget.NewProgressBarInfinite()
	progressBar.Hide()

	updateButton := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), nil)
	updateButton.Hide()

	infoButton := widget.NewButtonWithIcon("", theme.InfoIcon(), func() {
		descriptionLabel.SetText("📌 " + software.Description)
	})
	infoButton.Importance = widget.LowImportance

	version := widget.NewLabelWithStyle("v: "+software.Version, fyne.TextAlignTrailing, fyne.TextStyle{Bold: true, Italic: true})

	// JSON fayldan yuklangan dasturlarni o'qish
	downloadedSoftwares, err := storage.LoadDownloadedSoftware("downloaded_software.json")
	if err != nil {
		fmt.Println("JSON faylni o'qishda xatolik:", err)
	}

	// Dastur holatini tekshirish
	var isDownloaded bool
	var isLatestVersion bool
	for _, downloaded := range downloadedSoftwares {
		if downloaded.ID == software.ID {
			isDownloaded = true
			if downloaded.Version == software.Version {
				isLatestVersion = true
			}
			break
		}
	}

	// Tugmalarni holatga qarab ko'rsatish
	if !isDownloaded {
		downloadButton.Show() // JSON da bo'lmasa faqat yuklab olish tugmasi
	} else {
		deleteButton.Show() // JSON da bo'lsa o'chirish tugmasi
		if !isLatestVersion {
			updateButton.Show() // Versiya eskirgan bo'lsa yangilash tugmasi
		}
	}

	// O'chirish tugmasi funksiyasi
	deleteButton.OnTapped = func() {

		// Tasdiqlash dialogi
		dialog.ShowConfirm("O'chirish", "Ushbu dasturni ro'yxatdan o'chirishni xohlaysizmi?", func(confirmed bool) {
			if confirmed {
				// Dastur ma'lumotlarini olish
				softwareData, err := storage.GetSoftwareByID(software.ID, "downloaded_software.json")
				if err != nil {
					fmt.Println(err)
					fyne.CurrentApp().SendNotification(fyne.NewNotification("Xatolik", "Dastur topilmadi"))
					return
				}

				// Papkani o'chirish
				if softwareData.DirPath != "" {
					err = os.RemoveAll(softwareData.DirPath)
					if err != nil {
						fmt.Println("Papkani o'chirishda xatolik:", err)
						fyne.CurrentApp().SendNotification(fyne.NewNotification("Xatolik", "Papka o'chirilmadi"))
						return
					}
				}
				err = storage.DeleteSoftware(software.ID, "downloaded_software.json")
				if err != nil {
					fmt.Println("O'chirishda xatolik:", err)
					return
				}

				services.RemoveShortcut(softwareData.Name)

				// UI ni yangilash
				deleteButton.Hide()
				updateButton.Hide()
				downloadButton.Show()
				fyne.CurrentApp().SendNotification(fyne.NewNotification("O'chirildi", software.Name+" ro'yxatdan o'chirildi"))
			}
		}, myWindow)
	}

	// Yuklash tugmasi funksiyasi
	downloadButton.OnTapped = func() {
		// dialog.ShowFolderOpen(func(folder fyne.ListableURI, err error) {
		// 	if err != nil || folder == nil {
		// 		return
		// 	}

		folder := services.GetUserLocalPath()
		services.DownloadPath = filepath.Join(folder, software.ID)

		if _, err := os.Stat(services.DownloadPath); err == nil {
			err = os.RemoveAll(services.DownloadPath)
			if err != nil {
				fmt.Println("Papkani o'chirishda xatolik:", err)
				return
			}
		}

		err = os.Mkdir(services.DownloadPath, 0755)
		if err != nil {
			fmt.Println("Papka yaratishda xatolik:", err)
			return
		}

		fileURL := fmt.Sprintf("http://localhost:8080/appStore/download/%s", software.ID)

		downloadButton.Hide()
		deleteButton.Hide()
		updateButton.Hide()
		progressBar.Show()

		go func() {
			mainFilePath, iconFilePath, err := services.DownloadFile(fileURL, services.DownloadPath)
			if err != nil {
				progressBar.Hide()
				downloadButton.Show()
				fyne.CurrentApp().SendNotification(fyne.NewNotification("Xatolik", "Yuklashda xatolik: "+err.Error()))
				return
			}

			softwareData := models.DownloadedSoftware{
				ID:           software.ID,
				Name:         software.Name,
				Version:      software.Version,
				DirPath:      services.DownloadPath,
				MainFile:     filepath.Base(mainFilePath),
				IconPath:     filepath.Base(iconFilePath), // Ikonka yo'li saqlanadi
				DownloadDate: time.Now().Format("2006-01-02 15:04:05"),
			}

			err = storage.SaveDownloadedSoftware(softwareData, "downloaded_software.json")
			if err != nil {
				fmt.Println("Saqlashda xatolik:", err)
			}
			progressBar.Hide()
			deleteButton.Show()
			fyne.CurrentApp().SendNotification(fyne.NewNotification("Yuklandi va o'rnatildi", mainFilePath))

			programPath := filepath.Join(softwareData.DirPath, softwareData.MainFile)
			err = services.CreateShortcut(programPath, softwareData.Name)
			if err != nil {
				fmt.Println("Dastur o'rnatilishida xatolik: ", err)
			}
		}()

	}

	// Yangilash tugmasi funksiyasi
	updateButton.OnTapped = func() {
		data, err := storage.GetSoftwareByID(software.ID, "downloaded_software.json")
		if err != nil {
			fmt.Println(err)
			data = &models.DownloadedSoftware{} // Bo'sh ko'rsatkich
		}

		softwareData := *data // Dereferencing: * bilan qiymatni olish

		softwareData = models.DownloadedSoftware{
			ID:           software.ID,
			Name:         software.Name,
			Version:      software.Version,
			DirPath:      softwareData.DirPath,
			MainFile:     software.MainFile,
			DownloadDate: time.Now().Format("2006-01-02 15:04:05"),
		}

		// Papkani o'chirish
		if softwareData.DirPath != "" {
			err = os.RemoveAll(softwareData.DirPath)
			if err != nil {
				fmt.Println("Papkani o'chirishda xatolik:", err)
				fyne.CurrentApp().SendNotification(fyne.NewNotification("Xatolik", "Papka o'chirilmadi"))
				return
			}
		}

		services.RemoveShortcut(softwareData.Name)

		fmt.Println("softwareData: ", softwareData)
		storage.SaveDownloadedSoftware(softwareData, "downloaded_software.json")
		fileURL := fmt.Sprintf("http://localhost:8080/appStore/download/%s", software.ID)
		// filePath := filepath.Join(softwareData.DirPath, software.MainFile)

		downloadButton.Hide()
		deleteButton.Hide()
		updateButton.Hide()
		progressBar.Show()

		go func() {
			mainFile, _, err := services.DownloadFile(fileURL, softwareData.DirPath)
			if err != nil {
				fmt.Println("Xatolik:", err)
				progressBar.Hide()
				deleteButton.Show()
				updateButton.Show() // Xatolik bo'lsa qayta yangilash imkoni
				return
			}

			programPath := filepath.Join(softwareData.DirPath, softwareData.MainFile)
			err = services.CreateShortcut(programPath, softwareData.Name)
			if err != nil {
				fmt.Println("Dastur o'rnatilishida xatolik: ", err)
			}

			progressBar.Hide()
			deleteButton.Show() // Yangilash tugagach o'chirish tugmasi ko'rinadi

			filePath := filepath.Join(services.DownloadPath, mainFile)
			fyne.CurrentApp().SendNotification(fyne.NewNotification("Yangilandi", filePath))
		}()
	}

	// Versiya va ma'lumot tugmasi uchun konteyner
	versionContainer := container.NewHBox(
		infoButton,
		layout.NewSpacer(),
		version,
	)

	// Tugmalar uchun konteyner
	buttonContainer := container.NewHBox(
		progressBar,
		deleteButton, // O'chirish tugmasi qo'shildi
		downloadButton,
		updateButton,
	)

	// Sarlavha va tugmalar konteyneri
	titleContainer := container.NewBorder(nil, nil, nil, buttonContainer, title)

	// Kartaning asosiy kontenti
	content := container.NewVBox(
		titleContainer,
		container.NewCenter(iconImage),
		versionContainer,
	)

	// Kartani ajratuvchi chiziq bilan birlashtirish
	border := widget.NewSeparator()
	card := container.NewStack(border, content)

	return card
}
