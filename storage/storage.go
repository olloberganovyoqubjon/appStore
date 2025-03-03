package storage

// Ma'lumotlarni saqlash logikasi uchun paket
import (
	"encoding/json" // JSON bilan ishlash uchun
	"main/models"   // Loyiha modellarini ishlatish uchun
	"os"            // Fayl tizimi bilan ishlash uchun
)

// Yuklangan dastur ma'lumotlarini faylga saqlash funksiyasi
func SaveDownloadedSoftware(software models.DownloadedSoftware, filePath string) error {
	// Oldin yuklangan dasturlarni saqlash uchun ro'yxat
	var softwares []models.DownloadedSoftware

	// Fayl mavjudligini tekshirish
	if _, err := os.Stat(filePath); err == nil {
		// Faylni ochish
		file, err := os.Open(filePath)
		if err != nil { // Agar ochishda xatolik bo'lsa
			return err // Xatolikni qaytarish
		}
		// Faylni yopish
		defer file.Close()
		// JSON dan ma'lumotlarni o'qish
		json.NewDecoder(file).Decode(&softwares)
	}

	// Yangi dasturni ro'yxatga qo'shish
	softwares = append(softwares, software)

	// Faylni qayta yaratish (eski ma'lumot ustiga yozish uchun)
	file, err := os.Create(filePath)
	if err != nil { // Agar yaratishda xatolik bo'lsa
		return err // Xatolikni qaytarish
	}
	// Faylni yopish
	defer file.Close()

	// Yangilangan ro'yxatni JSON sifatida yozish
	return json.NewEncoder(file).Encode(softwares)
}
