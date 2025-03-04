package main

// Asosiy paket - dastur ishga tushadigan joy
import (
	UI "main/ui" // Loyihaning UI komponentlari paketi

	"fyne.io/fyne/v2"     // Fyne GUI frameworkning asosiy paketi
	"fyne.io/fyne/v2/app" // Fyne ilovasini boshqarish uchun
)

func main() {
	// Yangi Fyne ilovasini "men-go-fyne" ID bilan yaratish
	myApp := app.NewWithID("men-go-fyne")

	// "Men Go Fyne" nomli yangi oyna yaratish
	myWindow := myApp.NewWindow("Men Go Fyne")

	// Oyna o'lchamini 800x500 pikselga sozlash
	myWindow.Resize(fyne.NewSize(800, 500))

	// UI ni sozlash funksiyasini chaqirish (ui paketidan)
	UI.SetupUI(myWindow)

	// Oynani ko'rsatish va dasturni ishga tushirish
	myWindow.ShowAndRun()
}
