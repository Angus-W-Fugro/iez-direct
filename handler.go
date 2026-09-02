package main

import (
	"database/sql"
	"embed"
	"encoding/base64"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/Angus-Warman/httpmin/parserequest"
	_ "github.com/go-sql-driver/mysql"
)

type Handler struct {
	db *sql.DB
}

func NewHandler() (*Handler, error) {
	dsn := os.Getenv("MYSQL_CONN")
	dsn = dsn + "?parseTime=true"

	db, err := sql.Open("mysql", dsn)

	if err != nil {
		return nil, err
	}

	h := &Handler{
		db,
	}

	return h, nil
}

func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	err := h.db.Ping()

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

type DatagridInit struct {
	Endpoint string
	Columns  bool
}

type GridParams struct {
	NumRows int
	Page    int
}

func (h *Handler) DvLogsPage(w http.ResponseWriter, r *http.Request) {
	data := DatagridInit{
		Endpoint: "/api/dv-logs",
	}

	h.render(w, "dv-logs.tmpl", data)
}

func (h *Handler) DvLogsData(w http.ResponseWriter, r *http.Request) {
	gp, err := parserequest.As[GridParams](r)

	if err != nil {
		htmlResponse(w, "<span id='response'>"+err.Error()+"</span>")
		return
	}

	grid, err := h.getDvLogsData(gp)

	if err != nil {
		htmlResponse(w, "<span id='response'>"+err.Error()+"</span>")
		return
	}

	gridResponse(w, grid)
}

func (h *Handler) getDvLogsData(gp GridParams) (Grid, error) {
	return DvLogDefinition.Grid(h.db, gp)
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

	idBytes, err := base64.StdEncoding.DecodeString(s.RowID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = DvLogDefinition.UpdateCell(h.db, idBytes, s.Column, s.Value)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Saved"))
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

func gridResponse(w http.ResponseWriter, grid Grid) {
	err := tmpls.ExecuteTemplate(w, "datagrid-rows", grid)

	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 500)
	}
}
