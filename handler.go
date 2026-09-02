package main

import (
	"embed"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/Angus-Warman/httpmin/parserequest"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Handler struct {
	db *gorm.DB
}

func NewHandler() (*Handler, error) {
	dsn := os.Getenv("MYSQL_CONN")
	dsn = dsn + "?parseTime=true"

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

func (h *Handler) DvLogsData(w http.ResponseWriter, r *http.Request) {
	data, err := h.getDvLogsData()

	if err != nil {
		htmlResponse(w, "<span id='response'>"+err.Error()+"</span>")
		return
	}

	grid := DvLogsToGrid(data)

	dataGridResponse(w, "#dv-log-table", grid)
}

func (h *Handler) getDvLogsData() ([]DvLog, error) {
	rows := []DvLog{}

	err := h.db.Table("surf_dv_logs").Limit(20).Find(&rows).Error

	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (h *Handler) EditDvLogCell(w http.ResponseWriter, r *http.Request) {
	type EditSignal struct {
		RowID  string // Base64
		Column string
		Value  string
	}

	s, err := parserequest.As[EditSignal](r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Println(s)

	w.WriteHeader(204)
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

func dataGridResponse(w http.ResponseWriter, target string, grid Grid) {
	w.Header().Set("content-type", "text/html")
	w.Header().Set("datastar-selector", target)

	err := tmpls.ExecuteTemplate(w, "datagrid", grid)

	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 500)
	}
}
