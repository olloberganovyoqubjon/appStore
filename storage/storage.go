package storage

import (
	"encoding/json"
	"fmt"
	"main/models"
	"os"
)

// Yuklangan dastur ma'lumotlarini faylga saqlash funksiyasi
func SaveDownloadedSoftware(software models.DownloadedSoftware, filePath string) error {
	var softwares []models.DownloadedSoftware

	if _, err := os.Stat(filePath); err == nil {
		file, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer file.Close()
		json.NewDecoder(file).Decode(&softwares)
	}

	updated := false
	for i, existing := range softwares {
		if existing.ID == software.ID {
			softwares[i] = software
			updated = true
			break
		}
	}
	if !updated {
		softwares = append(softwares, software)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(softwares)
}

// JSON fayldan yuklangan dasturlarni o'qish funksiyasi
func LoadDownloadedSoftware(filePath string) ([]models.DownloadedSoftware, error) {
	var softwares []models.DownloadedSoftware

	if _, err := os.Stat(filePath); err == nil {
		file, err := os.Open(filePath)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		err = json.NewDecoder(file).Decode(&softwares)
		if err != nil {
			return nil, err
		}
	}

	return softwares, nil
}

// JSON fayldan dasturni o'chirish funksiyasi
func DeleteSoftware(id string, filePath string) error {
	var softwares []models.DownloadedSoftware

	// Fayl mavjudligini tekshirish va o'qish
	if _, err := os.Stat(filePath); err == nil {
		file, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer file.Close()
		json.NewDecoder(file).Decode(&softwares)
	}

	// Yangi ro'yxat yaratib, mos ID ni o'chirish
	var updatedSoftwares []models.DownloadedSoftware
	for _, s := range softwares {
		if s.ID != id { // ID mos kelmasa, ro'yxatga qo'shish
			updatedSoftwares = append(updatedSoftwares, s)
		}
	}

	// Faylni qayta yaratish
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Yangilangan ro'yxatni JSON sifatida yozish
	return json.NewEncoder(file).Encode(updatedSoftwares)
}

// JSON fayldan bitta dasturni ID bo'yicha olish funksiyasi
func GetSoftwareByID(id string, filePath string) (*models.DownloadedSoftware, error) {
	var softwares []models.DownloadedSoftware

	// Fayl mavjudligini tekshirish va o'qish
	if _, err := os.Stat(filePath); err == nil {
		file, err := os.Open(filePath)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		err = json.NewDecoder(file).Decode(&softwares)
		if err != nil {
			return nil, err
		}
	}

	// ID bo'yicha dasturni qidirish
	for _, s := range softwares {
		if s.ID == id {
			return &s, nil // Topilgan dasturni qaytarish
		}
	}

	// Agar topilmasa, xatolik qaytarish
	return nil, fmt.Errorf("dastur topilmadi: ID = %s", id)
}
