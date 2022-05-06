package presentation

import (
	"log"
	"net/http"
)

type ViewData struct {
	Title string
	Data  interface{}
}

type Message struct {
	Message string
}

type HomePageData struct {
	Message string
}

// Home домашняя страница
func (a *App) Home(w http.ResponseWriter, _ *http.Request) {
	data := ViewData{
		Title: "Главная страница",
		Data: HomePageData{
			Message: "Социальная сеть №1: поиск единомышленников",
		},
	}
	tmpl, err := loadTemplate("index.html")
	if err != nil {
		log.Println("bad template ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.ExecuteTemplate(w, "base", data)
	if err != nil {
		log.Println("template not executed ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
