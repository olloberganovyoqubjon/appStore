package services

// API bilan ishlash va fayl yuklash xizmatlari uchun paket
import (
	"encoding/json" // JSON bilan ishlash uchun
	"io"            // Kirish/chiqish operatsiyalari uchun
	"main/models"   // Loyiha modellarini ishlatish uchun
	"net/http"      // HTTP so'rovlar uchun
	"os"            // Fayl tizimi bilan ishlash uchun
)

// Default yuklab olish papkasi yo'li (global o'zgaruvchi)
var DownloadPath = "C:/Downloads"

// API dan dastur ma'lumotlarini olish funksiyasi
func FetchAPIData(url string) ([]models.Software, error) {
	// URL ga HTTP GET so'rov yuborish
	resp, err := http.Get(url)
	if err != nil { // Agar so'rovda xatolik bo'lsa
		return nil, err // Xatolikni qaytarish
	}
	// Funksiya tugagach javob tanasini yopish
	defer resp.Body.Close()

	// API javobini saqlash uchun ResponseData strukturasi
	var data models.ResponseData

	// JSON ma'lumotni dekodlash
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil { // Agar dekodlashda xatolik bo'lsa
		return nil, err // Xatolikni qaytarish
	}

	// Dasturlar ro'yxatini qaytarish
	return data.Object, nil
}

// Faylni URL dan yuklab olish funksiyasi
func DownloadFile(url, filePath string) error {
	// URL ga HTTP GET so'rov yuborish
	resp, err := http.Get(url)
	if err != nil { // Agar so'rovda xatolik bo'lsa
		return err // Xatolikni qaytarish
	}
	// Javob tanasini yopish
	defer resp.Body.Close()

	// Mahalliy faylni yaratish
	out, err := os.Create(filePath)
	if err != nil { // Agar fayl yaratishda xatolik bo'lsa
		return err // Xatolikni qaytarish
	}
	// Faylni yopish
	defer out.Close()

	// HTTP javobidan faylga ma'lumotni ko'chirish
	_, err = io.Copy(out, resp.Body)
	return err // Xatolik bo'lsa qaytarish, aks holda nil
}
