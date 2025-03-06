package services

import (
	"archive/zip"     // ZIP arxivlar bilan ishlash uchun
	"encoding/base64" // Base64 kodlash/dekodlash uchun
	"encoding/json"   // JSON bilan ishlash uchun
	"fmt"             // Formatlash va xatoliklarni chop etish uchun
	"io"              // Fayl o‘qish/yozish uchun
	"main/models"     // `models` paketidagi tuzilmalarni ishlatish uchun
	"net/http"        // HTTP so‘rovlar uchun
	"os"              // Operatsion tizim bilan ishlash uchun (fayllar, papkalar)
	"os/exec"         // Tashqi buyruqlarni ishga tushirish uchun
	"path/filepath"   // Fayl yo‘llarini boshqarish uchun
)

// Yuklash javobi uchun tuzilma
type DownloadResponse struct {
	ID          string `json:"id"`          // Dastur ID si
	Name        string `json:"name"`        // Dastur nomi
	Description string `json:"description"` // Dastur tavsifi
	MainFile    string `json:"mainFile"`    // Asosiy ishga tushiriladigan fayl
	Version     string `json:"version"`     // Dastur versiyasi
	Icon        string `json:"icon"`        // Base64 kodlangan ikonka
	File        string `json:"file"`        // Base64 kodlangan ZIP fayl
}

var DownloadPath = "C:/Downloads" // Standart yuklash yo‘li (o‘zgaruvchi)

// Faylni va ikonani URL dan yuklab olish va ZIP ni ochish funksiyasi
func DownloadFile(url, dirPath string) (mainFilePath, iconFilePath string, err error) {
	resp, err := http.Get(url) // URL ga HTTP GET so‘rov yuboradi
	if err != nil {            // Agar xatolik bo‘lsa
		return "", "", fmt.Errorf("HTTP so'rovda xatolik: %v", err) // Xatolik xabarini qaytaradi
	}
	defer resp.Body.Close() // Funksiya tugagach, javobni yopadi

	if resp.StatusCode != http.StatusOK { // Agar serverdan 200 OK bo‘lmasa
		return "", "", fmt.Errorf("serverdan noto'g'ri javob: %s", resp.Status) // Xatolik xabarini qaytaradi
	}

	var downloadResp DownloadResponse                      // Yuklash javobi uchun o‘zgaruvchi
	err = json.NewDecoder(resp.Body).Decode(&downloadResp) // JSON ni dekod qiladi
	if err != nil {                                        // Agar xatolik bo‘lsa
		return "", "", fmt.Errorf("JSON dekodlashda xatolik: %v", err) // Xatolik xabarini qaytaradi
	}

	fileData, err := base64.StdEncoding.DecodeString(downloadResp.File) // ZIP faylni Base64 dan dekod qiladi
	if err != nil {                                                     // Agar xatolik bo‘lsa
		return "", "", fmt.Errorf("Base64 fayl dekodlashda xatolik: %v", err) // Xatolik xabarini qaytaradi
	}

	zipFileName := downloadResp.Name + ".zip"          // ZIP fayl nomi (dastur nomi + .zip)
	zipFilePath := filepath.Join(dirPath, zipFileName) // ZIP faylning to‘liq yo‘li

	out, err := os.Create(zipFilePath) // ZIP faylni yaratadi
	if err != nil {                    // Agar xatolik bo‘lsa
		return "", "", fmt.Errorf("ZIP fayl yaratishda xatolik: %v", err) // Xatolik xabarini qaytaradi
	}
	defer out.Close() // Funksiya tugagach, faylni yopadi

	_, err = out.Write(fileData) // Dekodlangan ZIP ma’lumotlarini faylga yozadi
	if err != nil {              // Agar xatolik bo‘lsa
		return "", "", fmt.Errorf("ZIP faylga yozishda xatolik: %v", err) // Xatolik xabarini qaytaradi
	}

	// ZIP faylni ochish
	zipReader, err := zip.OpenReader(zipFilePath) // ZIP faylni o‘qish uchun ochadi
	if err != nil {                               // Agar xatolik bo‘lsa
		return "", "", fmt.Errorf("ZIP faylni ochishda xatolik: %v", err) // Xatolik xabarini qaytaradi
	}

	// ZIP ichidagi fayllarni chiqarish
	for _, f := range zipReader.File { // ZIP ichidagi har bir fayl bo‘yicha tsikl
		fPath := filepath.Join(dirPath, f.Name) // Faylning chiqariladigan yo‘li
		if f.FileInfo().IsDir() {               // Agar bu papka bo‘lsa
			os.MkdirAll(fPath, os.ModePerm) // Papkani yaratadi
			continue                        // Keyingi faylga o‘tadi
		}

		rc, err := f.Open() // ZIP ichidagi faylni ochadi
		if err != nil {     // Agar xatolik bo‘lsa
			zipReader.Close()                                                          // ZIP o‘quvchini yopadi
			return "", "", fmt.Errorf("ZIP ichidagi faylni ochishda xatolik: %v", err) // Xatolik xabarini qaytaradi
		}
		defer rc.Close() // Fayl o‘qish tugagach, yopadi

		outFile, err := os.Create(fPath) // Chiqariladigan faylni yaratadi
		if err != nil {                  // Agar xatolik bo‘lsa
			zipReader.Close()                                             // ZIP o‘quvchini yopadi
			return "", "", fmt.Errorf("fayl yaratishda xatolik: %v", err) // Xatolik xabarini qaytaradi
		}
		defer outFile.Close() // Funksiya tugagach, faylni yopadi

		_, err = io.Copy(outFile, rc) // ZIP ichidagi faylni nusxalaydi
		if err != nil {               // Agar xatolik bo‘lsa
			zipReader.Close()                                              // ZIP o‘quvchini yopadi
			return "", "", fmt.Errorf("faylni saqlashda xatolik: %v", err) // Xatolik xabarini qaytaradi
		}

		if f.Name == downloadResp.MainFile { // Agar bu asosiy fayl bo‘lsa
			mainFilePath = fPath // Asosiy fayl yo‘lini saqlaydi
		}
	}

	// ZIP readerni yopish
	err = zipReader.Close() // ZIP o‘quvchini yopadi
	if err != nil {         // Agar xatolik bo‘lsa
		return "", "", fmt.Errorf("ZIP readerni yopishda xatolik: %v", err) // Xatolik xabarini qaytaradi
	}

	// Ikonka faylni dekodlash va saqlash
	iconData, err := base64.StdEncoding.DecodeString(downloadResp.Icon) // Ikonkani Base64 dan dekod qiladi
	if err != nil {                                                     // Agar xatolik bo‘lsa
		return "", "", fmt.Errorf("Base64 ikonka dekodlashda xatolik: %v", err) // Xatolik xabarini qaytaradi
	}

	iconFilePath = filepath.Join(dirPath, downloadResp.Name+".png") // Ikonka faylning yo‘li
	iconOut, err := os.Create(iconFilePath)                         // Ikonka faylni yaratadi
	if err != nil {                                                 // Agar xatolik bo‘lsa
		return "", "", fmt.Errorf("ikonka fayl yaratishda xatolik: %v", err) // Xatolik xabarini qaytaradi
	}
	defer iconOut.Close() // Funksiya tugagach, faylni yopadi

	_, err = iconOut.Write(iconData) // Ikonka ma’lumotlarini faylga yozadi
	if err != nil {                  // Agar xatolik bo‘lsa
		return "", "", fmt.Errorf("ikonka faylga yozishda xatolik: %v", err) // Xatolik xabarini qaytaradi
	}

	if mainFilePath == "" && len(zipReader.File) > 0 { // Agar asosiy fayl topilmasa va ZIP da fayllar bo‘lsa
		mainFilePath = filepath.Join(dirPath, zipReader.File[0].Name) // Birinchi faylni asosiy deb oladi
	}

	return mainFilePath, iconFilePath, nil // Asosiy fayl va ikonka yo‘llarini qaytaradi
}

// API dan dastur ma'lumotlarini olish funksiyasi
func FetchAPIData(url string) ([]models.Software, error) {
	resp, err := http.Get(url) // URL ga HTTP GET so‘rov yuboradi
	if err != nil {            // Agar xatolik bo‘lsa
		return nil, err // Xatolikni qaytaradi
	}
	defer resp.Body.Close() // Funksiya tugagach, javobni yopadi

	var data models.ResponseData                   // Javob uchun tuzilma
	err = json.NewDecoder(resp.Body).Decode(&data) // JSON ni dekod qiladi
	if err != nil {                                // Agar xatolik bo‘lsa
		return nil, err // Xatolikni qaytaradi
	}

	return data.Object, nil // Dasturlar ro‘yxatini qaytaradi
}

// Foydalanuvchi mahalliy papkasini olish
func GetUserLocalPath() string {
	homeDir, err := os.UserHomeDir() // Foydalanuvchi asosiy papkasini oladi
	if err != nil {                  // Agar xatolik bo‘lsa
		fmt.Println("Xatolik: Foydalanuvchi papkasi aniqlanmadi.") // Xatolikni konsolga chiqaradi
		return "C:\\Users\\Default\\AppData\\Local"                // Standart yo‘lni qaytaradi
	}
	return filepath.Join(homeDir, "AppData", "Local") // Foydalanuvchi Local papkasini qaytaradi
}

// ZIP faylni Local papkaga ochish
func ExtractZIPToLocal(src string) error {
	dest := GetUserLocalPath()            // Foydalanuvchi Local papkasini oladi
	err := os.MkdirAll(dest, os.ModePerm) // Agar papka mavjud bo‘lmasa, yaratadi
	if err != nil {                       // Agar xatolik bo‘lsa
		return err // Xatolikni qaytaradi
	}

	r, err := zip.OpenReader(src) // ZIP faylni o‘qish uchun ochadi
	if err != nil {               // Agar xatolik bo‘lsa
		return err // Xatolikni qaytaradi
	}
	defer r.Close() // Funksiya tugagach, ZIP ni yopadi

	var extractedFilePath string // Chiqarilgan fayl yo‘li uchun o‘zgaruvchi

	for _, file := range r.File { // ZIP ichidagi fayllar bo‘yicha tsikl
		if file.FileInfo().IsDir() { // Agar bu papka bo‘lsa
			continue // Keyingi faylga o‘tadi
		}

		// Asosiy faylni chiqarish
		extractedFilePath = filepath.Join(dest, file.Name) // Fayl yo‘lini birlashtiradi
		outFile, err := os.Create(extractedFilePath)       // Chiqariladigan faylni yaratadi
		if err != nil {                                    // Agar xatolik bo‘lsa
			return err // Xatolikni qaytaradi
		}
		defer outFile.Close() // Funksiya tugagach, faylni yopadi

		rc, err := file.Open() // ZIP ichidagi faylni ochadi
		if err != nil {        // Agar xatolik bo‘lsa
			return err // Xatolikni qaytaradi
		}
		defer rc.Close() // Fayl o‘qish tugagach, yopadi

		_, err = io.Copy(outFile, rc) // Faylni nusxalaydi
		if err != nil {               // Agar xatolik bo‘lsa
			return err // Xatolikni qaytaradi
		}
		break // Faqat birinchi faylni chiqarib, to‘xtaydi
	}

	return nil // Muvaffaqiyatli yakunlanadi
}

// Shortcut yaratish
func CreateShortcut(targetPath, appName string) error {
	desktopPath := filepath.Join(os.Getenv("USERPROFILE"), "Desktop", appName+".lnk")                                // Ishchi stolda yorliq yo‘li
	startMenuPath := filepath.Join(os.Getenv("APPDATA"), "Microsoft\\Windows\\Start Menu\\Programs", appName+".lnk") // Start menyuda yorliq yo‘li

	shortcutScript := fmt.Sprintf(` // VBScript kodini tayyorlaydi
		set WshShell = WScript.CreateObject("WScript.Shell")
		set shortcut = WshShell.CreateShortcut("%s")
		shortcut.TargetPath = "%s"
		shortcut.Save
		set shortcut = WshShell.CreateShortcut("%s")
		shortcut.TargetPath = "%s"
		shortcut.Save
	`, desktopPath, targetPath, startMenuPath, targetPath)

	tempScript := filepath.Join(os.Getenv("TEMP"), "create_shortcut.vbs") // Vaqtinchalik VBS fayl yo‘li
	err := os.WriteFile(tempScript, []byte(shortcutScript), 0644)         // VBS skriptni faylga yozadi
	if err != nil {                                                       // Agar xatolik bo‘lsa
		return err // Xatolikni qaytaradi
	}

	cmd := exec.Command("wscript", tempScript) // VBS skriptni ishga tushirish buyrug‘i
	err = cmd.Run()                            // Buyruqni bajaradi
	if err != nil {                            // Agar xatolik bo‘lsa
		return err // Xatolikni qaytaradi
	}

	os.Remove(tempScript) // Vaqtinchalik faylni o‘chiradi
	return nil            // Muvaffaqiyatli yakunlanadi
}

// Shortcut'larni o‘chirish
func RemoveShortcut(appName string) error {
	desktopPath := filepath.Join(os.Getenv("USERPROFILE"), "Desktop", appName+".lnk")                                // Ishchi stolda yorliq yo‘li
	startMenuPath := filepath.Join(os.Getenv("APPDATA"), "Microsoft\\Windows\\Start Menu\\Programs", appName+".lnk") // Start menyuda yorliq yo‘li

	// Ishchi stoldagi shortcut’ni o‘chirish
	if _, err := os.Stat(desktopPath); err == nil { // Agar yorliq mavjud bo‘lsa
		err := os.Remove(desktopPath) // Yorliqni o‘chiradi
		if err != nil {               // Agar xatolik bo‘lsa
			fmt.Println("Ishchi stoldan shortcut o‘chirishda xatolik:", err) // Xatolikni konsolga chiqaradi
		} else { // Agar muvaffaqiyatli bo‘lsa
			fmt.Println("Ishchi stol shortcut o‘chirildi:", desktopPath) // Muvaffaqiyat xabarini chiqaradi
		}
	}

	// Start menyudan shortcut’ni o‘chirish
	if _, err := os.Stat(startMenuPath); err == nil { // Agar yorliq mavjud bo‘lsa
		err := os.Remove(startMenuPath) // Yorliqni o‘chiradi
		if err != nil {                 // Agar xatolik bo‘lsa
			fmt.Println("Start menyudan shortcut o‘chirishda xatolik:", err) // Xatolikni konsolga chiqaradi
		} else { // Agar muvaffaqiyatli bo‘lsa
			fmt.Println("Start menyu shortcut o‘chirildi:", startMenuPath) // Muvaffaqiyat xabarini chiqaradi
		}
	}

	return nil // Muvaffaqiyatli yakunlanadi (xatolik bo‘lsa ham nil qaytaradi)
}
