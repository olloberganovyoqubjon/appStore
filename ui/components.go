package ui

// Foydalanuvchi interfeysi komponentlari uchun paket
import (
	"bytes"           // Baytlar bilan ishlash uchun
	"encoding/base64" // Base64 dekodlash uchun
	"fmt"             // Formatlash va chop etish uchun
	"image"           // Rasm ma'lumotlari bilan ishlash uchun
	_ "image/png"     // PNG formatini qo'llab-quvvatlash uchun
	"path/filepath"   // Fayl yo'llarini boshqarish uchun
	"strings"         // Stringlar bilan ishlash uchun
	"time"            // Vaqt bilan ishlash uchun

	"main/models"   // Loyiha modellarini ishlatish uchun
	"main/services" // Xizmatlar paketini ishlatish uchun
	"main/storage"  // Saqlash paketini ishlatish uchun

	"fyne.io/fyne/v2"           // Fyne asosiy paketi
	"fyne.io/fyne/v2/canvas"    // Grafik elementlar uchun
	"fyne.io/fyne/v2/container" // Konteynerlar uchun
	"fyne.io/fyne/v2/dialog"    // Dialog oynalari uchun
	"fyne.io/fyne/v2/layout"    // Tartibni boshqarish uchun
	"fyne.io/fyne/v2/theme"     // Temalar va ikonkalarni ishlatish uchun
	"fyne.io/fyne/v2/widget"    // Vidjetlar uchun
)

// UI ni sozlash funksiyasi
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
	contentContainer := container.NewGridWrap(fyne.NewSize(150, 150))

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

// Dastur kartasini yaratish funksiyasi
func createSoftwareCard(software models.Software, descriptionLabel *widget.Label, myWindow fyne.Window) fyne.CanvasObject {
	// Dastur nomini qalin va markazlashtirilgan yorliq sifatida yaratish
	title := widget.NewLabelWithStyle(software.Name, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	// Ikonka uchun o'zgaruvchi
	var iconImage *canvas.Image
	if software.Icon != "" { // Agar ikonka mavjud bo'lsa
		// Base64 dan dekodlash
		decoded, err := base64.StdEncoding.DecodeString(software.Icon)
		if err != nil { // Agar dekodlashda xatolik bo'lsa
			// Xato ikonkasini ishlatish
			iconImage = canvas.NewImageFromResource(theme.ErrorIcon())
		} else { // Dekodlash muvaffaqiyatli bo'lsa
			// Baytlardan rasmni dekodlash
			img, _, err := image.Decode(bytes.NewReader(decoded))
			if err != nil { // Agar rasm dekodlashda xatolik bo'lsa
				// Xato ikonkasini ishlatish
				iconImage = canvas.NewImageFromResource(theme.ErrorIcon())
			} else { // Rasm muvaffaqiyatli dekodlansa
				// Rasmni Fyne tasviriga aylantirish
				iconImage = canvas.NewImageFromImage(img)
				// Tasvirni o'lchamga moslashtirish
				iconImage.FillMode = canvas.ImageFillContain
				// Minimal o'lchamni belgilash
				iconImage.SetMinSize(fyne.NewSize(80, 80))
			}
		}
	}
	if iconImage == nil { // Agar ikonka hali aniqlanmagan bo'lsa
		// Standart fayl ikonkasini ishlatish
		iconImage = canvas.NewImageFromResource(theme.FileIcon())
	}

	// Yuklash tugmasi (ikonka bilan)
	downloadButton := widget.NewButtonWithIcon("", theme.DownloadIcon(), nil)
	// Tugma ahamiyatini past darajaga qo'yish
	downloadButton.Importance = widget.LowImportance

	// Cheksiz progress bar (yuklanayotganlik ko'rsatkichi)
	progressBar := widget.NewProgressBarInfinite()
	// Dastlab progress barni yashirish
	progressBar.Hide()

	// Yangilash tugmasi
	updateButton := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), nil)
	// Dastlab yangilash tugmasini yashirish
	updateButton.Hide()

	// Ma'lumot tugmasi
	infoButton := widget.NewButtonWithIcon("", theme.InfoIcon(), func() {
		// Tavsifni ko'rsatish
		descriptionLabel.SetText("📌 " + software.Description)
	})
	// Tugma ahamiyatini past darajaga qo'yish
	infoButton.Importance = widget.LowImportance

	// Versiya yorlig'i (qalin va kursiv)
	version := widget.NewLabelWithStyle("v: "+software.Version, fyne.TextAlignTrailing, fyne.TextStyle{Bold: true, Italic: true})

	// Yuklash tugmasi bosilganda
	downloadButton.OnTapped = func() {
		// Papka tanlash dialogini ko'rsatish
		dialog.ShowFolderOpen(func(folder fyne.ListableURI, err error) {
			if err != nil || folder == nil { // Agar xatolik bo'lsa yoki papka tanlanmasa
				return // Funksiyadan chiqish
			}

			// Tanlangan papka yo'lini saqlash
			services.DownloadPath = folder.Path()
			// Yuklash URL ni shakllantirish
			fileURL := fmt.Sprintf("http://localhost:8080/appStore/download/%s", software.ID)
			// Faylning mahalliy yo'lini yaratish
			filePath := filepath.Join(services.DownloadPath, software.MainFile)

			// Yuklash tugmasini yashirish
			downloadButton.Hide()
			// Progress barni ko'rsatish
			progressBar.Show()

			// Goroutinada yuklash (UI bloklanmasligi uchun)
			go func() {
				// Faylni yuklash
				err := services.DownloadFile(fileURL, filePath)
				if err != nil { // Agar xatolik bo'lsa
					// Xatolikni konsolga chop etish
					fmt.Println("Xatolik:", err)
					return // Goroutindan chiqish
				}

				// Yuklangan dastur ma'lumotlarini tayyorlash
				softwareData := models.DownloadedSoftware{
					ID:           software.ID,
					Name:         software.Name,
					Version:      software.Version,
					FilePath:     filePath,
					DownloadDate: time.Now().Format("2006-01-02 15:04:05"), // Hozirgi vaqt
				}

				// Ma'lumotni JSON faylga saqlash
				storage.SaveDownloadedSoftware(softwareData, "downloaded_software.json")

				// Progress barni yashirish
				progressBar.Hide()
				// Yangilash tugmasini ko'rsatish
				updateButton.Show()

				// Yuklash tugaganligi haqida bildirishnoma yuborish
				fyne.CurrentApp().SendNotification(fyne.NewNotification("Yuklandi", filePath))
			}()
		}, myWindow)
	}

	// Versiya va ma'lumot tugmasi uchun gorizontal konteyner
	versionContainer := container.NewHBox(
		infoButton,         // Ma'lumot tugmasi
		layout.NewSpacer(), // Bo'sh joy
		version,            // Versiya yorlig'i
	)

	// Tugmalar uchun gorizontal konteyner
	buttonContainer := container.NewHBox(
		progressBar,    // Progress bar
		downloadButton, // Yuklash tugmasi
		updateButton,   // Yangilash tugmasi
	)

	// Sarlavha va tugmalar konteyneri
	titleContainer := container.NewBorder(nil, nil, nil, buttonContainer, title)

	// Kartaning asosiy kontenti (vertikal)
	content := container.NewVBox(
		titleContainer,                 // Sarlavha va tugmalar
		container.NewCenter(iconImage), // Ikonkani markazlashtirish
		versionContainer,               // Versiya konteyneri
	)

	// Kartani ajratuvchi chiziq
	border := widget.NewSeparator()
	// Kartani chiziq va kontent bilan birlashtirish
	card := container.NewStack(border, content)

	// Kartani qaytarish
	return card
}
