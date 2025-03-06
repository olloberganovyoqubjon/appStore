package storage

import (
	"encoding/json" // JSON ma'lumotlarni kodlash va dekodlash uchun
	"errors"        // Maxsus xatoliklarni yaratish uchun
	"fmt"           // Xatolik xabarlarini formatlash uchun
	"main/models"   // `models` paketidagi tuzilmalarni ishlatish uchun
	"os"            // Fayl tizimi bilan ishlash uchun (fayl ochish, yozish)
)

// Maxsus xatoliklar
var (
	ErrFileNotFound     = errors.New("fayl topilmadi")            // Fayl tizimda mavjud emasligini bildiruvchi xatolik
	ErrFileOpenFailed   = errors.New("faylni ochishda xatolik")   // Faylni ochishda muammo bo‘lganini bildiruvchi xatolik
	ErrFileCreateFailed = errors.New("faylni yaratishda xatolik") // Fayl yaratishda xatolik yuz berganini bildiruvchi xatolik
	ErrJSONDecodeFailed = errors.New("JSON dekodlashda xatolik")  // JSON ma'lumotlarni dekod qilishda xatolikni ko‘rsatuvchi xatolik
	ErrJSONEncodeFailed = errors.New("JSON kodlashda xatolik")    // JSON ma'lumotlarni kodlashda xatolikni ko‘rsatuvchi xatolik
	ErrSoftwareNotFound = errors.New("dastur topilmadi")          // Berilgan ID ga mos dastur topilmaganini bildiruvchi xatolik
)

// Yuklangan dastur ma'lumotlarini faylga saqlash funksiyasi
func SaveDownloadedSoftware(software models.DownloadedSoftware, filePath string) error {
	var softwares []models.DownloadedSoftware // Dasturlar ro‘yxati uchun bo‘sh massiv

	// Fayl mavjudligini tekshirish
	if _, err := os.Stat(filePath); err == nil { // Fayl mavjudligini tekshiradi
		file, err := os.Open(filePath) // Faylni o‘qish uchun ochadi
		if err != nil {                // Agar faylni ochishda xatolik bo‘lsa
			return fmt.Errorf("%w: %s - %v", ErrFileOpenFailed, filePath, err) // Maxsus xatolik bilan qaytaradi
		}
		defer file.Close() // Funksiya tugagach, faylni yopadi

		if err := json.NewDecoder(file).Decode(&softwares); err != nil { // JSON dan dasturlarni dekod qiladi
			return fmt.Errorf("%w: %s - %v", ErrJSONDecodeFailed, filePath, err) // Dekodlash xatoligi bilan qaytaradi
		}
	} else if !os.IsNotExist(err) { // Agar fayl mavjud bo‘lmasa va boshqa xatolik bo‘lsa
		return fmt.Errorf("%w: %s - %v", ErrFileNotFound, filePath, err) // Fayl topilmadi xatoligi bilan qaytaradi
	}

	// Dasturni yangilash yoki qo‘shish
	updated := false                     // Dastur yangilanganligini tekshirish uchun flag
	for i, existing := range softwares { // Mavjud dasturlar bo‘yicha tsikl
		if existing.ID == software.ID { // Agar ID mos kelsa
			softwares[i] = software // Mavjud dasturni yangilaydi
			updated = true          // Yangilanganligini belgilaydi
			break                   // Tsikldan chiqadi
		}
	}
	if !updated { // Agar yangilanmagan bo‘lsa
		softwares = append(softwares, software) // Yangi dasturni ro‘yxatga qo‘shadi
	}

	// Faylni qayta yaratish va yozish
	file, err := os.Create(filePath) // Faylni qayta yozish uchun ochadi (ustiga yozadi)
	if err != nil {                  // Agar fayl yaratishda xatolik bo‘lsa
		return fmt.Errorf("%w: %s - %v", ErrFileCreateFailed, filePath, err) // Yaratish xatoligi bilan qaytaradi
	}
	defer file.Close() // Funksiya tugagach, faylni yopadi

	if err := json.NewEncoder(file).Encode(softwares); err != nil { // Yangilangan ro‘yxatni JSON ga kodlaydi
		return fmt.Errorf("%w: %s - %v", ErrJSONEncodeFailed, filePath, err) // Kodlash xatoligi bilan qaytaradi
	}

	return nil // Muvaffaqiyatli yakunlanadi
}

// JSON fayldan yuklangan dasturlarni o'qish funksiyasi
func LoadDownloadedSoftware(filePath string) ([]models.DownloadedSoftware, error) {
	var softwares []models.DownloadedSoftware // Dasturlar ro‘yxati uchun bo‘sh massiv

	if _, err := os.Stat(filePath); err == nil { // Fayl mavjudligini tekshiradi
		file, err := os.Open(filePath) // Faylni o‘qish uchun ochadi
		if err != nil {                // Agar faylni ochishda xatolik bo‘lsa
			return nil, fmt.Errorf("%w: %s - %v", ErrFileOpenFailed, filePath, err) // Ochish xatoligi bilan qaytaradi
		}
		defer file.Close() // Funksiya tugagach, faylni yopadi

		if err := json.NewDecoder(file).Decode(&softwares); err != nil { // JSON dan dasturlarni dekod qiladi
			return nil, fmt.Errorf("%w: %s - %v", ErrJSONDecodeFailed, filePath, err) // Dekodlash xatoligi bilan qaytaradi
		}
	} else if !os.IsNotExist(err) { // Agar fayl mavjud bo‘lmasa va boshqa xatolik bo‘lsa
		return nil, fmt.Errorf("%w: %s - %v", ErrFileNotFound, filePath, err) // Fayl topilmadi xatoligi bilan qaytaradi
	}

	return softwares, nil // Dasturlar ro‘yxatini qaytaradi (bo‘sh bo‘lsa ham nil xatoliksiz)
}

// JSON fayldan dasturni o'chirish funksiyasi
func DeleteSoftware(id string, filePath string) error {
	var softwares []models.DownloadedSoftware // Dasturlar ro‘yxati uchun bo‘sh massiv

	// Fayl mavjudligini tekshirish va o'qish
	if _, err := os.Stat(filePath); err == nil { // Fayl mavjudligini tekshiradi
		file, err := os.Open(filePath) // Faylni o‘qish uchun ochadi
		if err != nil {                // Agar faylni ochishda xatolik bo‘lsa
			return fmt.Errorf("%w: %s - %v", ErrFileOpenFailed, filePath, err) // Ochish xatoligi bilan qaytaradi
		}
		defer file.Close() // Funksiya tugagach, faylni yopadi

		if err := json.NewDecoder(file).Decode(&softwares); err != nil { // JSON dan dasturlarni dekod qiladi
			return fmt.Errorf("%w: %s - %v", ErrJSONDecodeFailed, filePath, err) // Dekodlash xatoligi bilan qaytaradi
		}
	} else if !os.IsNotExist(err) { // Agar fayl mavjud bo‘lmasa va boshqa xatolik bo‘lsa
		return fmt.Errorf("%w: %s - %v", ErrFileNotFound, filePath, err) // Fayl topilmadi xatoligi bilan qaytaradi
	}

	// Yangi ro'yxat yaratib, mos ID ni o'chirish
	var updatedSoftwares []models.DownloadedSoftware // Yangilangan ro‘yxat uchun bo‘sh massiv
	found := false                                   // Dastur topilganligini tekshirish uchun flag
	for _, s := range softwares {                    // Mavjud dasturlar bo‘yicha tsikl
		if s.ID != id { // Agar ID mos kelmasa
			updatedSoftwares = append(updatedSoftwares, s) // Dasturni yangi ro‘yxatga qo‘shadi
		} else { // Agar ID mos kelsa
			found = true // Dastur topilganligini belgilaydi
		}
	}

	if !found { // Agar dastur topilmasa
		return fmt.Errorf("%w: ID = %s", ErrSoftwareNotFound, id) // Dastur topilmadi xatoligi bilan qaytaradi
	}

	// Faylni qayta yaratish
	file, err := os.Create(filePath) // Faylni qayta yozish uchun ochadi (ustiga yozadi)
	if err != nil {                  // Agar fayl yaratishda xatolik bo‘lsa
		return fmt.Errorf("%w: %s - %v", ErrFileCreateFailed, filePath, err) // Yaratish xatoligi bilan qaytaradi
	}
	defer file.Close() // Funksiya tugagach, faylni yopadi

	if err := json.NewEncoder(file).Encode(updatedSoftwares); err != nil { // Yangilangan ro‘yxatni JSON ga kodlaydi
		return fmt.Errorf("%w: %s - %v", ErrJSONEncodeFailed, filePath, err) // Kodlash xatoligi bilan qaytaradi
	}

	return nil // Muvaffaqiyatli yakunlanadi
}

// JSON fayldan bitta dasturni ID bo'yicha olish funksiyasi
func GetSoftwareByID(id string, filePath string) (*models.DownloadedSoftware, error) {
	var softwares []models.DownloadedSoftware // Dasturlar ro‘yxati uchun bo‘sh massiv

	// Fayl mavjudligini tekshirish va o'qish
	if _, err := os.Stat(filePath); err == nil { // Fayl mavjudligini tekshiradi
		file, err := os.Open(filePath) // Faylni o‘qish uchun ochadi
		if err != nil {                // Agar faylni ochishda xatolik bo‘lsa
			return nil, fmt.Errorf("%w: %s - %v", ErrFileOpenFailed, filePath, err) // Ochish xatoligi bilan qaytaradi
		}
		defer file.Close() // Funksiya tugagach, faylni yopadi

		if err := json.NewDecoder(file).Decode(&softwares); err != nil { // JSON dan dasturlarni dekod qiladi
			return nil, fmt.Errorf("%w: %s - %v", ErrJSONDecodeFailed, filePath, err) // Dekodlash xatoligi bilan qaytaradi
		}
	} else if !os.IsNotExist(err) { // Agar fayl mavjud bo‘lmasa va boshqa xatolik bo‘lsa
		return nil, fmt.Errorf("%w: %s - %v", ErrFileNotFound, filePath, err) // Fayl topilmadi xatoligi bilan qaytaradi
	}

	// ID bo'yicha dasturni qidirish
	for _, s := range softwares { // Mavjud dasturlar bo‘yicha tsikl
		if s.ID == id { // Agar ID mos kelsa
			return &s, nil // Topilgan dastur ko‘rsatkichini qaytaradi
		}
	}

	return nil, fmt.Errorf("%w: ID = %s", ErrSoftwareNotFound, id) // Dastur topilmadi xatoligi bilan qaytaradi
}
