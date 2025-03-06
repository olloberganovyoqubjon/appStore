package ui

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	_ "image/png"
	"os"
	"os/exec"
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
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func SetupUI(myWindow fyne.Window) {
	// Yuklanish xabari uchun yorliq
	label := widget.NewLabel("Ma'lumot yuklanmoqda...") // "Ma'lumot yuklanmoqda..." matnli yangi yorliq yaratadi

	// Dastur tavsifi uchun bo'sh yorliq
	descriptionLabel := widget.NewLabel("") // Bo‘sh matnli yorliq yaratadi, keyinchalik dastur tavsifi uchun ishlatiladi

	// API dan dasturlarni olish
	softwares, err := services.FetchAPIData("http://localhost:8080/appStore/getAllSoftware") // API dan dasturlar ro‘yxatini oladi
	if err != nil {                                                                          // Agar xatolik bo'lsa
		// Xatolik xabarini yorliqqa yozish
		label.SetText(fmt.Sprintf("Xatolik: %v", err)) // Xatolik haqida xabar yorliqqa yoziladi
		// Oynaga faqat xatolik xabarini joylashtirish
		myWindow.SetContent(container.NewVBox(label)) // Faqat xatolik xabari ko‘rsatilgan oynani yangilaydi
		return                                        // Funksiyadan chiqish
	}

	// Qidiruv maydoni yaratish
	searchEntry := widget.NewEntry() // Yangi matn kiritish maydoni (input) yaratadi
	// Placeholder matn qo'shish
	searchEntry.SetPlaceHolder("Dastur nomi yoki tavsif bo'yicha qidirish...") // Qidiruv maydoniga placeholder matn qo‘shadi
	searchEntry.Resize(fyne.NewSize(400, searchEntry.MinSize().Height))        // Qidiruv maydonining o‘lchamini 400xstandart balandlikka o‘zgartiradi

	// Qayta yuklash tugmasi
	reloadButton := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), func() { // Ikonkali (refresh) tugma yaratadi
		SetupUI(myWindow) // Tugma bosilganda UI ni qayta yuklaydi
	})

	// Qidiruv maydoni va tugmani joylashtirish
	searchContainer := container.NewBorder(nil, nil, nil, reloadButton, searchEntry) // Qidiruv maydoni va tugmani o‘ng tarafda joylashtiradi

	// Dastur kartalari uchun grid konteyner (150x150 o'lchamli)
	contentContainer := container.NewGridWrap(fyne.NewSize(150, 140)) // 150x140 o‘lchamdagi grid konteyner yaratadi

	// Dastur ro'yxatini yangilash funksiyasi
	updateSoftwareList := func(query string) { // Qidiruv so‘roviga asoslangan dastur ro‘yxatini yangilaydi
		// Avvalgi ob'ektlarni tozalash
		contentContainer.Objects = nil // Grid ichidagi barcha obyektlarni o‘chiradi
		// Filtlangan dasturlarni saqlash uchun massiv
		var filteredSoftwares []fyne.CanvasObject // Filtlangan dastur kartalari uchun massiv
		// Qidiruv so'zini kichik harflarga aylantirish
		query = strings.ToLower(query) // Qidiruv matnini kichik harflarga o‘zgartiradi
		// Har bir dastur uchun tsikl
		for _, software := range softwares { // Har bir dastur bo‘yicha tsikl yuritadi
			// Agar qidiruv bo'sh yoki nom/tavsifda so'z bo'lsa
			if query == "" || strings.Contains(strings.ToLower(software.Name), query) || strings.Contains(strings.ToLower(software.Description), query) {
				// Dastur kartasini yaratish
				card := createSoftwareCard(software, descriptionLabel, myWindow) // Dastur uchun kartani yaratadi
				// Kartani ro'yxatga qo'shish
				filteredSoftwares = append(filteredSoftwares, card) // Kartani filtlangan ro‘yxatga qo‘shadi
			}
		}
		// Yangi kartalarni konteynerga qo'shish
		contentContainer.Objects = filteredSoftwares // Gridga yangi kartalarni joylashtiradi
		// Konteynerni yangilash
		contentContainer.Refresh() // Gridni yangilaydi
	}

	// Boshlang'ich ro'yxatni ko'rsatish (qidiruvsiz)
	updateSoftwareList("") // Dastlab bo‘sh qidiruv bilan ro‘yxatni ko‘rsatadi

	// Qidiruv maydonida matn o'zgarsa
	searchEntry.OnChanged = func(query string) { // Qidiruv maydonidagi matn o‘zgarsa ishlaydi
		// Ro'yxatni yangilash
		updateSoftwareList(query) // Yangi qidiruv so‘roviga asosan ro‘yxatni yangilaydi
	}

	// Asosiy vertikal konteyner
	mainContainer := container.NewVBox( // Vertikal tartibdagi asosiy konteyner yaratadi
		container.NewPadded(searchContainer), // Qidiruv maydoni + tugmani padding bilan qo‘shadi
		contentContainer,                     // Dastur kartalari gridini qo‘shadi
		widget.NewSeparator(),                // Ajratuvchi chiziq qo‘shadi
		descriptionLabel,                     // Tavsif yorlig‘ini qo‘shadi
	)

	// Oynaga asosiy konteynerni joylashtirish
	myWindow.SetContent(mainContainer) // Oynaning asosiy tarkibini yangilaydi
}

func createSoftwareCard(software models.Software, descriptionLabel *widget.Label, myWindow fyne.Window) fyne.CanvasObject {
	// Dastur nomini yorliq sifatida yaratish
	title := widget.NewLabelWithStyle(software.Name, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}) // Dastur nomini qalin va markazda ko‘rsatadi

	// Ikonka tasvirini tayyorlash
	var iconImage *canvas.Image // Ikonka uchun o‘zgaruvchi
	if software.Icon != "" {    // Agar dasturda ikonka bo‘lsa
		decoded, err := base64.StdEncoding.DecodeString(software.Icon) // Base64 dan ikonka ma’lumotlarini dekod qiladi
		if err != nil {                                                // Agar dekodlashda xatolik bo‘lsa
			iconImage = canvas.NewImageFromResource(theme.ErrorIcon()) // Xato ikonini ko‘rsatadi
		} else { // Dekodlash muvaffaqiyatli bo‘lsa
			img, _, err := image.Decode(bytes.NewReader(decoded)) // Dekodlangan baytlardan tasvir yaratadi
			if err != nil {                                       // Tasvirni dekodlashda xatolik bo‘lsa
				iconImage = canvas.NewImageFromResource(theme.ErrorIcon()) // Xato ikonini ko‘rsatadi
			} else { // Tasvir muvaffaqiyatli yaratilsa
				iconImage = canvas.NewImageFromImage(img)    // Tasvirdan ikonka obyektini yaratadi
				iconImage.FillMode = canvas.ImageFillContain // Tasvirni konteynerga moslashtiradi
				iconImage.SetMinSize(fyne.NewSize(100, 100)) // Ikonka o‘lchamini 100x100 ga sozlaydi
			}
		}
	}
	if iconImage == nil { // Agar ikonka hali ham bo‘sh bo‘lsa
		iconImage = canvas.NewImageFromResource(theme.FileIcon()) // Standart fayl ikonkasini ishlatadi
	}

	// Tugmalar va progress bar yaratish
	deleteButton := widget.NewButtonWithIcon("", theme.DeleteIcon(), nil) // "O‘chirish" tugmasi (faqat ikonka)
	deleteButton.Importance = widget.LowImportance                        // Tugma muhimligini past darajaga qo‘yadi
	deleteButton.Hide()                                                   // Tugmani yashiradi

	downloadButton := widget.NewButtonWithIcon("", theme.DownloadIcon(), nil) // "Yuklash" tugmasi (faqat ikonka)
	downloadButton.Importance = widget.LowImportance                          // Tugma muhimligini past darajaga qo‘yadi
	downloadButton.Hide()                                                     // Tugmani yashiradi

	openButton := widget.NewButtonWithIcon("", theme.ComputerIcon(), nil) // "Ochish" tugmasi (faqat ikonka)
	openButton.Importance = widget.LowImportance                          // Tugma muhimligini past darajaga qo‘yadi
	openButton.Hide()                                                     // Tugmani yashiradi

	progressBar := widget.NewProgressBarInfinite() // Cheksiz progress bar yaratadi
	progressBar.Hide()                             // Progress barni yashiradi

	updateButton := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), nil) // "Yangilash" tugmasi (faqat ikonka)
	updateButton.Hide()                                                        // Tugmani yashiradi

	infoButton := widget.NewButtonWithIcon("", theme.InfoIcon(), func() { // "Ma’lumot" tugmasi (faqat ikonka)
		descriptionLabel.SetText("📌 " + software.Description) // Tugma bosilganda tavsifni ko‘rsatadi
	})
	infoButton.Importance = widget.LowImportance // Tugma muhimligini past darajaga qo‘yadi

	// JSON fayldan yuklangan dasturlarni o‘qish
	downloadedSoftwares, err := storage.LoadDownloadedSoftware("downloaded_software.json") // Yuklangan dasturlarni JSON dan o‘qiydi
	if err != nil {
		dialog.ShowError(fmt.Errorf("JSON o'qishda xatolik: %v", err), myWindow) // Agar xatolik bo‘lsa
	}

	// Dastur holatini tekshirish
	var isDownloaded bool                            // Dastur yuklanganligini tekshirish uchun flag
	var isLatestVersion bool                         // Dastur oxirgi versiyada ekanligini tekshirish uchun flag
	for _, downloaded := range downloadedSoftwares { // Yuklangan dasturlar bo‘yicha tsikl
		if downloaded.ID == software.ID { // Agar dastur ID si mos kelsa
			isDownloaded = true                         // Dastur yuklangan deb belgilanadi
			if downloaded.Version == software.Version { // Agar versiyalar bir xil bo‘lsa
				isLatestVersion = true // Oxirgi versiya deb belgilanadi
			}
			break // Tsikldan chiqadi
		}
	}

	// Tugmalarni holatga qarab ko‘rsatish
	if !isDownloaded { // Agar dastur yuklanmagan bo‘lsa
		downloadButton.Show() // Faqat "Yuklash" tugmasini ko‘rsatadi
	} else { // Agar dastur yuklangan bo‘lsa
		deleteButton.Show()   // "O‘chirish" tugmasini ko‘rsatadi
		openButton.Show()     // "Ochish" tugmasini ko‘rsatadi
		if !isLatestVersion { // Agar versiya eskirgan bo‘lsa
			updateButton.Show() // "Yangilash" tugmasini ko‘rsatadi
		}
	}

	// O‘chirish tugmasi funksiyasi
	deleteButton.OnTapped = func() {
		dialog.ShowConfirm("O'chirish", "Ushbu dasturni ro'yxatdan o'chirishni xohlaysizmi?", func(confirmed bool) {
			if confirmed {
				softwareData, err := storage.GetSoftwareByID(software.ID, "downloaded_software.json")
				if err != nil {
					if errors.Is(err, storage.ErrSoftwareNotFound) {
						dialog.ShowError(fmt.Errorf("dastur topilmadi: %s", software.ID), myWindow)
					} else {
						dialog.ShowError(fmt.Errorf("ma'lumot olishda xatolik: %v", err), myWindow)
					}
					return
				}

				if softwareData.DirPath != "" {
					err = os.RemoveAll(softwareData.DirPath)
					if err != nil {
						dialog.ShowError(fmt.Errorf("papkani o'chirishda xatolik: %v", err), myWindow)
						return
					}
				}

				err = storage.DeleteSoftware(software.ID, "downloaded_software.json")
				if err != nil {
					switch {
					case errors.Is(err, storage.ErrFileOpenFailed):
						dialog.ShowError(fmt.Errorf("faylni ochishda xatolik: %v", err), myWindow)
					case errors.Is(err, storage.ErrFileCreateFailed):
						dialog.ShowError(fmt.Errorf("faylni qayta yaratishda xatolik: %v", err), myWindow)
					case errors.Is(err, storage.ErrJSONDecodeFailed):
						dialog.ShowError(fmt.Errorf("JSON o'qishda xatolik: %v", err), myWindow)
					case errors.Is(err, storage.ErrJSONEncodeFailed):
						dialog.ShowError(fmt.Errorf("JSON yozishda xatolik: %v", err), myWindow)
					case errors.Is(err, storage.ErrSoftwareNotFound):
						dialog.ShowError(fmt.Errorf("dastur topilmadi: %s", software.ID), myWindow)
					default:
						dialog.ShowError(fmt.Errorf("noma'lum xatolik: %v", err), myWindow)
					}
					return
				}

				services.RemoveShortcut(softwareData.Name)
				deleteButton.Hide()
				updateButton.Hide()
				downloadButton.Show()
				openButton.Hide()
				fyne.CurrentApp().SendNotification(fyne.NewNotification("O'chirildi", software.Name+" ro'yxatdan o'chirildi"))
			}
		}, myWindow)
	}

	// Yuklash tugmasi funksiyasi
	downloadButton.OnTapped = func() { // "Yuklash" tugmasi bosilganda ishlaydi
		folder := services.GetUserLocalPath()                      // Foydalanuvchi mahalliy papkasini oladi
		services.DownloadPath = filepath.Join(folder, software.ID) // Yuklash yo‘lini dastur ID si bilan birlashtiradi

		if _, err := os.Stat(services.DownloadPath); err == nil { // Agar papka mavjud bo‘lsa
			err = os.RemoveAll(services.DownloadPath) // Papkani o‘chiradi
			if err != nil {
				dialog.ShowError(fmt.Errorf("papkani o'chirishda xatolik: %v", err), myWindow) // Agar xatolik bo‘lsa
				return                                                                         // Funksiyadan chiqadi
			}
		}

		err = os.Mkdir(services.DownloadPath, 0755) // Yangi papka yaratadi (ruxsatlar: 0755)
		if err != nil {                             // Agar xatolik bo‘lsa
			dialog.ShowError(fmt.Errorf("papka yaratishda xatolik: %v", err), myWindow)
			return // Funksiyadan chiqadi
		}

		fileURL := fmt.Sprintf("http://localhost:8080/appStore/download/%s", software.ID) // Yuklash URL sini yaratadi

		downloadButton.Hide() // "Yuklash" tugmasini yashiradi
		deleteButton.Hide()   // "O‘chirish" tugmasini yashiradi
		updateButton.Hide()   // "Yangilash" tugmasini yashiradi
		openButton.Hide()     // "Ochish" tugmasini yashiradi
		progressBar.Show()    // Progress barni ko‘rsatadi

		go func() { // Goroutine ishlatib, fon rejimida yuklashni amalga oshiradi
			mainFilePath, iconFilePath, err := services.DownloadFile(fileURL, services.DownloadPath) // Faylni yuklaydi
			if err != nil {                                                                          // Agar xatolik bo‘lsa
				progressBar.Hide()                                                                                     // Progress barni yashiradi
				downloadButton.Show()                                                                                  // "Yuklash" tugmasini qayta ko‘rsatadi
				fyne.CurrentApp().SendNotification(fyne.NewNotification("Xatolik", "Yuklashda xatolik: "+err.Error())) // Xatolik xabarini ko‘rsatadi
				return                                                                                                 // Funksiyadan chiqadi
			}

			softwareData := models.DownloadedSoftware{ // Yuklangan dastur ma’lumotlarini tayyorlaydi
				ID:           software.ID,                              // Dastur ID si
				Name:         software.Name,                            // Dastur nomi
				Version:      software.Version,                         // Dastur versiyasi
				DirPath:      services.DownloadPath,                    // Yuklash yo‘li
				MainFile:     filepath.Base(mainFilePath),              // Asosiy fayl nomi
				IconPath:     filepath.Base(iconFilePath),              // Ikonka fayl nomi
				DownloadDate: time.Now().Format("2006-01-02 15:04:05"), // Yuklash sanasi
			}

			err = storage.SaveDownloadedSoftware(softwareData, "downloaded_software.json")
			if err != nil {
				switch {
				case errors.Is(err, storage.ErrFileCreateFailed):
					dialog.ShowError(fmt.Errorf("faylni yaratishda xatolik: %v", err), myWindow)
				case errors.Is(err, storage.ErrJSONEncodeFailed):
					dialog.ShowError(fmt.Errorf("JSON yozishda xatolik: %v", err), myWindow)
				case errors.Is(err, storage.ErrFileOpenFailed):
					dialog.ShowError(fmt.Errorf("faylni ochishda xatolik: %v", err), myWindow)
				case errors.Is(err, storage.ErrJSONDecodeFailed):
					dialog.ShowError(fmt.Errorf("JSON o'qishda xatolik: %v", err), myWindow)
				default:
					dialog.ShowError(fmt.Errorf("yuklashda noma'lum xatolik: %v", err), myWindow)
				}
				progressBar.Hide()
				downloadButton.Show()
				return
			}
			progressBar.Hide()                                                                               // Progress barni yashiradi
			deleteButton.Show()                                                                              // "O‘chirish" tugmasini ko‘rsatadi
			openButton.Show()                                                                                // "Ochish" tugmasini ko‘rsatadi
			fyne.CurrentApp().SendNotification(fyne.NewNotification("Yuklandi va o'rnatildi", mainFilePath)) // Muvaffaqiyat xabarini ko‘rsatadi

			programPath := filepath.Join(softwareData.DirPath, softwareData.MainFile) // Dastur yo‘lini birlashtiradi
			err = services.CreateShortcut(programPath, softwareData.Name)             // Dastur yorlig‘ini yaratadi
			if err != nil {
				dialog.ShowError(fmt.Errorf("dastur o'rnatilishida xatolik: %v", err), myWindow) // Agar xatolik bo‘lsa
			}
		}()
	}

	// Yangilash tugmasi funksiyasi
	updateButton.OnTapped = func() { // "Yangilash" tugmasi bosilganda ishlaydi
		data, err := storage.GetSoftwareByID(software.ID, "downloaded_software.json") // Dastur ma’lumotlarini oladi
		if err != nil {                                                               // Agar xatolik bo‘lsa
			dialog.ShowError(fmt.Errorf("dastur ma’lumotlarini olishd xatolik: %v", err), myWindow) // Xatolikni konsolga chiqaradi
			data = &models.DownloadedSoftware{}                                                     // Bo‘sh obyekt yaratadi
		}

		softwareData := *data // Ma’lumotlarni dereference qiladi (ko‘rsatkichdan qiymatga)

		softwareData = models.DownloadedSoftware{ // Yangi ma’lumotlar bilan to‘ldiradi
			ID:           software.ID,                              // Dastur ID si
			Name:         software.Name,                            // Dastur nomi
			Version:      software.Version,                         // Yangi versiya
			DirPath:      softwareData.DirPath,                     // Avvalgi papka yo‘li
			MainFile:     software.MainFile,                        // Asosiy fayl
			DownloadDate: time.Now().Format("2006-01-02 15:04:05"), // Yangilash sanasi
		}

		// Papkani o‘chirish
		if softwareData.DirPath != "" { // Agar papka mavjud bo‘lsa
			err = os.RemoveAll(softwareData.DirPath) // Papkani o‘chiradi
			if err != nil {                          // Agar xatolik bo‘lsa
				dialog.ShowError(fmt.Errorf("papka o'chirishda xatolik: %v", err), myWindow)
				return // Funksiyadan chiqadi
			}
		}

		folder := services.GetUserLocalPath()                      // Foydalanuvchi mahalliy papkasini oladi
		services.DownloadPath = filepath.Join(folder, software.ID) // Yangi yuklash yo‘lini yaratadi

		err = os.Mkdir(services.DownloadPath, 0755) // Yangi papka yaratadi
		if err != nil {                             // Agar xatolik bo‘lsa
			dialog.ShowError(fmt.Errorf("papka yaratishda xatolik: %v", err), myWindow)
			return // Funksiyadan chiqadi
		}

		services.RemoveShortcut(softwareData.Name) // Avvalgi yorliqni o‘chiradi

		err = storage.SaveDownloadedSoftware(softwareData, "downloaded_software.json")
		if err != nil {
			switch {
			case errors.Is(err, storage.ErrFileCreateFailed):
				dialog.ShowError(fmt.Errorf("faylni yaratishda xatolik: %v", err), myWindow)
			case errors.Is(err, storage.ErrJSONEncodeFailed):
				dialog.ShowError(fmt.Errorf("JSON yozishda xatolik: %v", err), myWindow)
			case errors.Is(err, storage.ErrFileOpenFailed):
				dialog.ShowError(fmt.Errorf("faylni ochishda xatolik: %v", err), myWindow)
			case errors.Is(err, storage.ErrJSONDecodeFailed):
				dialog.ShowError(fmt.Errorf("JSON o'qishda xatolik: %v", err), myWindow)
			default:
				dialog.ShowError(fmt.Errorf("yuklashda noma'lum xatolik: %v", err), myWindow)
			}
			progressBar.Hide()
			downloadButton.Show()
			return
		}
		fileURL := fmt.Sprintf("http://localhost:8080/appStore/download/%s", software.ID) // Yuklash URL sini yaratadi

		downloadButton.Hide()                                                    // "Yuklash" tugmasini yashiradi
		deleteButton.Hide()                                                      // "O‘chirish" tugmasini yashiradi
		updateButton.Hide()                                                      // "Yangilash" tugmasini yashiradi
		openButton.Hide()                                                        // "Ochish" tugmasini yashiradi
		progressBar.Show()                                                       // Progress barni ko‘rsatadi
		mainFile, _, err := services.DownloadFile(fileURL, softwareData.DirPath) // Faylni yuklaydi
		go func() {                                                              // Goroutine ishlatib, fon rejimida yangilashni amalga oshiradi
			if err != nil { // Agar xatolik bo‘lsa
				fmt.Println("Xatolik:", err) // Xatolikni konsolga chiqaradi
				progressBar.Hide()           // Progress barni yashiradi
				deleteButton.Show()          // "O‘chirish" tugmasini ko‘rsatadi
				updateButton.Show()          // "Yangilash" tugmasini qayta ko‘rsatadi
				dialog.ShowError(fmt.Errorf("faylni yuklashda xatolik: %v", err), myWindow)
				return // Funksiyadan chiqadi
			}

			programPath := filepath.Join(softwareData.DirPath, softwareData.MainFile) // Dastur yo‘lini birlashtiradi
			err = services.CreateShortcut(programPath, softwareData.Name)             // Yangi yorliq yaratadi
			if err != nil {
				dialog.ShowError(fmt.Errorf("dastur o'rnatilishida xatolik: %v", err), myWindow) // Agar xatolik bo‘lsa
			}

			progressBar.Hide()  // Progress barni yashiradi
			openButton.Show()   // "Ochish" tugmasini ko‘rsatadi
			deleteButton.Show() // "O‘chirish" tugmasini ko‘rsatadi

			filePath := filepath.Join(services.DownloadPath, mainFile)                       // Yuklangan fayl yo‘lini birlashtiradi
			fyne.CurrentApp().SendNotification(fyne.NewNotification("Yangilandi", filePath)) // Muvaffaqiyat xabarini ko‘rsatadi
		}()
	}

	// "Ochish" tugmasi funksiyasi
	openButton.OnTapped = func() { // "Ochish" tugmasi bosilganda ishlaydi
		softwareData, err := storage.GetSoftwareByID(software.ID, "downloaded_software.json") // Dastur ma’lumotlarini oladi
		if err != nil {                                                                       // Agar xatolik bo‘lsa
			fmt.Println("", err) // Xatolikni konsolga chiqaradi
			dialog.ShowError(fmt.Errorf("dastur ma'lumotlarini olishda xatolik: %v", err), myWindow)
			return // Funksiyadan chiqadi
		}

		filePath := filepath.Join(softwareData.DirPath, softwareData.MainFile) // Dastur fayl yo‘lini birlashtiradi
		if _, err := os.Stat(filePath); os.IsNotExist(err) {                   // Agar fayl mavjud bo‘lmasa
			fyne.CurrentApp().SendNotification(fyne.NewNotification("Xatolik", "Dastur fayli topilmadi")) // Xatolik xabarini ko‘rsatadi
			return                                                                                        // Funksiyadan chiqadi
		}

		// Dasturni ishga tushirish
		cmd := exec.Command(filePath) // Dasturni ishga tushirish uchun buyruq tayyorlaydi
		err = cmd.Start()             // Dasturni ishga tushiradi
		if err != nil {               // Agar xatolik bo‘lsa
			dialog.ShowError(fmt.Errorf("dasturni ochishda xatolik: %v", err), myWindow) // Funksiyadan chiqadi
			return
		}
	}

	// Tugmalar konteyneri
	buttonContainer := container.NewVBox( // Tugmalarni vertikal tartibda joylashtiradi
		progressBar,    // Progress bar
		deleteButton,   // "O‘chirish" tugmasi
		downloadButton, // "Yuklash" tugmasi
		updateButton,   // "Yangilash" tugmasi
		openButton,     // "Ochish" tugmasi
		infoButton,     // "Ma’lumot" tugmasi
	)

	// Kartaning asosiy kontenti
	content := container.NewWithoutLayout( // Tartibsiz konteyner yaratadi
		title,           // Dastur nomi
		iconImage,       // Ikonka tasviri
		buttonContainer, // Tugmalar konteyneri
	)

	// title ning joylashuvini belgilaymiz
	// iconImage ning joylashuvi va o‘lchamini belgilaymiz
	iconImage.Move(fyne.NewPos(7, 35))       // Ikonkani (7, 35) koordinatasiga joylashtiradi
	iconImage.Resize(fyne.NewSize(100, 100)) // Ikonka o‘lchamini 100x100 ga o‘zgartiradi

	// title ning joylashuvini belgilaymiz
	title.Move(fyne.NewPos(50, 5)) // Nomni (50, 5) koordinatasiga joylashtiradi

	// buttonContainer ning joylashuvi va o‘lchamini belgilaymiz
	buttonContainer.Move(fyne.NewPos(120, 5))    // Tugmalarni (120, 5) koordinatasiga joylashtiradi
	buttonContainer.Resize(fyne.NewSize(30, 30)) // Tugmalar konteynerini 30x30 o‘lchamga sozlaydi

	// Kartani ajratuvchi chiziq bilan birlashtirish
	border := widget.NewSeparator()             // Ajratuvchi chiziq yaratadi
	card := container.NewStack(border, content) // Chiziq va kontentni birlashtiradi

	return card // Tayyor kartani qaytaradi
}
