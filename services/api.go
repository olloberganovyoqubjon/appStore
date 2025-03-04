package services

import (
	"archive/zip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"main/models"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

type DownloadResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MainFile    string `json:"mainFile"`
	Version     string `json:"version"`
	Icon        string `json:"icon"`
	File        string `json:"file"` // Base64 kodlangan ZIP fayl
}

var DownloadPath = "C:/Downloads"

// Faylni va ikonani URL dan yuklab olish va ZIP ni ochish funksiyasi
func DownloadFile(url, dirPath string) (mainFilePath, iconFilePath string, err error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", "", fmt.Errorf("HTTP so'rovda xatolik: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("serverdan noto'g'ri javob: %s", resp.Status)
	}

	var downloadResp DownloadResponse
	err = json.NewDecoder(resp.Body).Decode(&downloadResp)
	if err != nil {
		return "", "", fmt.Errorf("JSON dekodlashda xatolik: %v", err)
	}

	fileData, err := base64.StdEncoding.DecodeString(downloadResp.File)
	if err != nil {
		return "", "", fmt.Errorf("Base64 fayl dekodlashda xatolik: %v", err)
	}

	zipFileName := downloadResp.Name + ".zip"
	zipFilePath := filepath.Join(dirPath, zipFileName)

	out, err := os.Create(zipFilePath)
	if err != nil {
		return "", "", fmt.Errorf("ZIP fayl yaratishda xatolik: %v", err)
	}
	defer out.Close()

	_, err = out.Write(fileData)
	if err != nil {
		return "", "", fmt.Errorf("ZIP faylga yozishda xatolik: %v", err)
	}

	// ZIP faylni ochish
	zipReader, err := zip.OpenReader(zipFilePath)
	if err != nil {
		return "", "", fmt.Errorf("ZIP faylni ochishda xatolik: %v", err)
	}

	// ZIP ichidagi fayllarni chiqarish
	for _, f := range zipReader.File {
		fPath := filepath.Join(dirPath, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fPath, os.ModePerm)
			continue
		}

		rc, err := f.Open()
		if err != nil {
			zipReader.Close() // Xatolik bo'lsa, darhol yopamiz
			return "", "", fmt.Errorf("ZIP ichidagi faylni ochishda xatolik: %v", err)
		}
		defer rc.Close()

		outFile, err := os.Create(fPath)
		if err != nil {
			zipReader.Close() // Xatolik bo'lsa, darhol yopamiz
			return "", "", fmt.Errorf("fayl yaratishda xatolik: %v", err)
		}
		defer outFile.Close()

		_, err = io.Copy(outFile, rc)
		if err != nil {
			zipReader.Close() // Xatolik bo'lsa, darhol yopamiz
			return "", "", fmt.Errorf("faylni saqlashda xatolik: %v", err)
		}

		if f.Name == downloadResp.MainFile {
			mainFilePath = fPath
		}
	}

	// ZIP readerni yopish
	err = zipReader.Close()
	if err != nil {
		return "", "", fmt.Errorf("ZIP readerni yopishda xatolik: %v", err)
	}

	// Ikonka faylni dekodlash va saqlash
	iconData, err := base64.StdEncoding.DecodeString(downloadResp.Icon)
	if err != nil {
		return "", "", fmt.Errorf("Base64 ikonka dekodlashda xatolik: %v", err)
	}

	iconFilePath = filepath.Join(dirPath, downloadResp.Name+".png")
	iconOut, err := os.Create(iconFilePath)
	if err != nil {
		return "", "", fmt.Errorf("ikonka fayl yaratishda xatolik: %v", err)
	}
	defer iconOut.Close()

	_, err = iconOut.Write(iconData)
	if err != nil {
		return "", "", fmt.Errorf("ikonka faylga yozishda xatolik: %v", err)
	}

	if mainFilePath == "" && len(zipReader.File) > 0 {
		mainFilePath = filepath.Join(dirPath, zipReader.File[0].Name)
	}

	return mainFilePath, iconFilePath, nil
}

// API dan dastur ma'lumotlarini olish funksiyasi (o'zgarmagan)
func FetchAPIData(url string) ([]models.Software, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data models.ResponseData
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, err
	}

	return data.Object, nil
}

func GetUserLocalPath() string {
	homeDir, err := os.UserHomeDir() // Foydalanuvchi asosiy papkasini olish
	if err != nil {
		fmt.Println("Xatolik: Foydalanuvchi papkasi aniqlanmadi.")
		return "C:\\Users\\Default\\AppData\\Local" // Agar topilmasa, default yo‘l
	}
	return filepath.Join(homeDir, "AppData", "Local")
}

func ExtractZIPToLocal(src string) error {
	dest := GetUserLocalPath()            // Foydalanuvchi Local papkasi
	err := os.MkdirAll(dest, os.ModePerm) // Agar papka mavjud bo‘lmasa, yaratish
	if err != nil {
		return err
	}

	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	var extractedFilePath string

	for _, file := range r.File {
		if file.FileInfo().IsDir() {
			continue
		}

		// Asosiy faylni chiqarish
		extractedFilePath = filepath.Join(dest, file.Name)
		outFile, err := os.Create(extractedFilePath)
		if err != nil {
			return err
		}
		defer outFile.Close()

		rc, err := file.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		_, err = io.Copy(outFile, rc)
		if err != nil {
			return err
		}
		break
	}

	return nil
}

// Shortcut yaratish
func CreateShortcut(targetPath, appName string) error {
	desktopPath := filepath.Join(os.Getenv("USERPROFILE"), "Desktop", appName+".lnk")
	startMenuPath := filepath.Join(os.Getenv("APPDATA"), "Microsoft\\Windows\\Start Menu\\Programs", appName+".lnk")

	shortcutScript := fmt.Sprintf(`
		set WshShell = WScript.CreateObject("WScript.Shell")
		set shortcut = WshShell.CreateShortcut("%s")
		shortcut.TargetPath = "%s"
		shortcut.Save
		set shortcut = WshShell.CreateShortcut("%s")
		shortcut.TargetPath = "%s"
		shortcut.Save
	`, desktopPath, targetPath, startMenuPath, targetPath)

	tempScript := filepath.Join(os.Getenv("TEMP"), "create_shortcut.vbs")
	err := os.WriteFile(tempScript, []byte(shortcutScript), 0644)
	if err != nil {
		return err
	}

	cmd := exec.Command("wscript", tempScript)
	err = cmd.Run()
	if err != nil {
		return err
	}

	os.Remove(tempScript)
	return nil
}

// Shortcut'larni o‘chirish
func RemoveShortcut(appName string) error {
	desktopPath := filepath.Join(os.Getenv("USERPROFILE"), "Desktop", appName+".lnk")
	startMenuPath := filepath.Join(os.Getenv("APPDATA"), "Microsoft\\Windows\\Start Menu\\Programs", appName+".lnk")

	// Ishchi stoldagi shortcut’ni o‘chirish
	if _, err := os.Stat(desktopPath); err == nil {
		err := os.Remove(desktopPath)
		if err != nil {
			fmt.Println("Ishchi stoldan shortcut o‘chirishda xatolik:", err)
		} else {
			fmt.Println("Ishchi stol shortcut o‘chirildi:", desktopPath)
		}
	}

	// Start menyudan shortcut’ni o‘chirish
	if _, err := os.Stat(startMenuPath); err == nil {
		err := os.Remove(startMenuPath)
		if err != nil {
			fmt.Println("Start menyudan shortcut o‘chirishda xatolik:", err)
		} else {
			fmt.Println("Start menyu shortcut o‘chirildi:", startMenuPath)
		}
	}

	return nil
}
