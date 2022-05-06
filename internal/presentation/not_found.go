package presentation

import (
	"net/http"
)

// NotFound такой страницы не существует
func (a *App) NotFound(w http.ResponseWriter, _ *http.Request) {
	data := ViewData{
		Title: "Такой страницы не существует",
	}
	tmpl, err := loadTemplate("not_found.html")
	if err != nil {
		handleError("not_found", "load template", err, w)
		return
	}
	err = tmpl.ExecuteTemplate(w, "base", data)
	if err != nil {
		handleError("not_found", "execute template", err, w)
		return
	}
}
