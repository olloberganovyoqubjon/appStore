package main

import (
	"encoding/json"
	"os"
)

// DownloadedSoftware - yuklangan dastur ma’lumotlari
type DownloadedSoftware struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Version      string `json:"version"`
	FilePath     string `json:"file_path"`
	DownloadDate string `json:"download_date"`
	IsDesktop    bool   `json:"isDesktop"`
	IsStartup    bool   `json:"isStartup"`
	IsAutoStart  bool   `json:"isAutoStart"`
}

// Ma’lumotlarni faylga yozish
func SaveDownloadedSoftware(software DownloadedSoftware, filePath string) error {
	var softwares []DownloadedSoftware

	// Agar fayl mavjud bo'lsa, eski ma’lumotlarni o‘qib olish
	if _, err := os.Stat(filePath); err == nil {
		file, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer file.Close()
		json.NewDecoder(file).Decode(&softwares)
	}

	// Yangi dasturni ro‘yxatga qo‘shish
	softwares = append(softwares, software)

	// Yangilangan ma’lumotlarni faylga yozish
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(softwares)
}
