package main

import (
	"embed"
	"html/template"
	"net/http"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Handler struct {
	db *gorm.DB
}

func NewHandler() (*Handler, error) {
	dsn := os.Getenv("MYSQL_CONN")

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		return nil, err
	}

	h := &Handler{
		db,
	}

	return h, nil
}

func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	err := h.db.Exec("SELECT 1;").Error

	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	htmlResponse(w, "<span id='response'>ok</span>")
}

func htmlResponse(w http.ResponseWriter, text string) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(text))
}

func (h *Handler) DvLogsPage(w http.ResponseWriter, r *http.Request) {
	h.render(w, "dv-logs.tmpl", nil)
}

//go:embed all:templates
var templateFiles embed.FS

var tmpls = template.Must(template.ParseFS(templateFiles, "templates/*"))

func (h *Handler) render(w http.ResponseWriter, templateName string, data any) {
	err := tmpls.ExecuteTemplate(w, templateName, data)

	if err != nil {
		http.Error(w, err.Error(), 500)
	}
}
