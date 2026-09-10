package main

import (
	"net/http"
)

type WorkpacksPageData struct {
	Workpacks []Workpack
}

type Workpack struct {
	Name           string `db:"CODE"`
	Ongoing        bool
	AssignedPrefix string
	TotalVideos    int
	MovedVideos    int
}

func (h *Handler) WorkpacksPage(w http.ResponseWriter, r *http.Request) {
	data, err := h.getWorkpackData()

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	h.render(w, "workpacks.tmpl", data)
}

func (h *Handler) getWorkpackData() (*WorkpacksPageData, error) {
	const query = `SELECT CODE FROM workpacks`

	var rows []Workpack
	if err := h.db.Select(&rows, query); err != nil {
		return nil, err
	}

	return &WorkpacksPageData{Workpacks: rows}, nil
}
