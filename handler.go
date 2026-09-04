package main

import (
	"embed"
	"encoding/hex"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/Angus-Warman/httpmin/parserequest"
	"github.com/Angus-Warman/stf"
	"github.com/jmoiron/sqlx"
)

type Handler struct {
	db *sqlx.DB
}

func NewHandler() (*Handler, error) {
	dsn := os.Getenv("MYSQL_CONN")

	if dsn == "" {
		return nil, fmt.Errorf("no MYSQL_CONN in .env")
	}

	dsn = dsn + "?parseTime=true"

	db, err := sqlx.Connect("mysql", dsn)

	if err != nil {
		return nil, err
	}

	h := &Handler{
		db,
	}

	return h, nil
}

func (h *Handler) IndexPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "not found", 404)
		return
	}

	http.Redirect(w, r, "/dv-logs", http.StatusSeeOther)
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

type GridParams struct {
	NumRows int
	Page    int
	Filter  *string
	SortBy  *string
}

func (h *Handler) DvLogsPage(w http.ResponseWriter, r *http.Request) {
	dvLogs, err := GetDvLogs(h.db, &GridParams{Page: 1, NumRows: 50})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sortOptions := []string{}
	sortOptions = append(sortOptions, stf.Convert(DvLogColumns, func(c string) string { return c + " ASC" })...)
	sortOptions = append(sortOptions, stf.Convert(DvLogColumns, func(c string) string { return c + " DESC" })...)

	data := DvLogPageData{
		Rows:        toTableRows(1, dvLogs),
		Columns:     DvLogColumns,
		SortOptions: sortOptions,
		Page:        1,
		NumRows:     50,
		Colspan:     len(DvLogColumns) + 3,
	}

	h.render(w, "dv-logs.tmpl", data)
}

func (h *Handler) DvLogsData(w http.ResponseWriter, r *http.Request) {
	gp, err := parserequest.As[GridParams](r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	dvLogs, err := GetDvLogs(h.db, &gp)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := DvLogPageData{
		Rows:    toTableRows(((gp.Page-1)*gp.NumRows)+1, dvLogs),
		Colspan: len(DvLogColumns) + 3,
	}

	w.Header().Set("Content-Type", "text/html")
	err = tmpls.ExecuteTemplate(w, "dv-log-table-rows", data)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) Play(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if _, err := hex.DecodeString(id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.render(w, "play-modal", id)
}

func (h *Handler) Media(w http.ResponseWriter, r *http.Request) {
	id, err := hex.DecodeString(r.PathValue("id"))

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	path, err := DvLogVideoPath(h.db, id)

	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	log.Println("media:", path)

	http.ServeFile(w, r, path)
}

func (h *Handler) EditDvLogCell(w http.ResponseWriter, r *http.Request) {
	type EditSignal struct {
		RowID string // Hex
		Value string
	}

	s, err := parserequest.As[EditSignal](r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	idBytes, err := hex.DecodeString(s.RowID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = UpdateDvLogComment(h.db, idBytes, s.Value)

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
