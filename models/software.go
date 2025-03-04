package models

// Dastur ma'lumotlari uchun struktura (API dan keladi)
type Software struct {
	ID          string `json:"id"`          // Dastur identifikatori
	Name        string `json:"name"`        // Dastur nomi
	Description string `json:"description"` // Dastur tavsifi
	Version     string `json:"version"`     // Dastur versiyasi
	MainFile    string `json:"mainFile"`    // Asosiy fayl nomi
	Icon        string `json:"icon"`        // Base64 kodlangan ikonka
}

// API javobini saqlash uchun struktura
type ResponseData struct {
	Message string     `json:"message"` // API xabari yoki holati
	Object  []Software `json:"object"`  // Dasturlar ro'yxati
}

// Yuklangan dastur ma'lumotlari uchun struktura
type DownloadedSoftware struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Version      string `json:"version"`
	DirPath      string `json:"dir_path"`
	MainFile     string `json:"main_file"`
	IconPath     string `json:"icon_path"` // Yangi maydon
	DownloadDate string `json:"download_date"`
}
