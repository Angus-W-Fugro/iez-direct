package main

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
)

type DvLogRow struct {
	ID           []byte
	FirstFile    string
	LastFile     string
	Workpack     string
	Installation string
	Substructure string
	Component    string
	SpreadCode   string
	Comment      string
}

type DvLog struct {
	ID           string // Base64-encoded
	Files        []string
	Workpack     string
	Installation string
	Substructure string
	Component    string
	SpreadCode   string
	Comment      string
}

type DvLogColumn struct {
	Name string
}

var DvLogColumns = []DvLogColumn{
	{Name: "Workpack"},
	{Name: "Installation"},
	{Name: "Substructure"},
	{Name: "Component"},
	{Name: "Spread Code"},
	{Name: "Comment"},
}

type SortOption struct {
	Field string
	Name  string
}

var DvLogSortOptions = []SortOption{
	{Field: "Workpack", Name: "Workpack"},
	{Field: "Installation", Name: "Installation"},
	{Field: "Substructure", Name: "Substructure"},
	{Field: "Component", Name: "Component"},
	{Field: "SpreadCode", Name: "Spread Code"},
	{Field: "Comment", Name: "Comment"},
}

type DvLogTableRow struct {
	Number int
	Cells  []string
	Files  string
	ID     string
}

type DvLogPageData struct {
	Rows        []DvLogTableRow
	Columns     []DvLogColumn
	SortOptions []SortOption
	SortBy      string
	Page        int
	NumRows     int
	Colspan     int
}

func (d DvLog) Cells() []string {
	return []string{
		d.Workpack,
		d.Installation,
		d.Substructure,
		d.Component,
		d.SpreadCode,
		d.Comment,
	}
}

func (d DvLog) FilesString() string {
	return strings.Join(d.Files, ", ")
}

func toTableRows(startNumber int, dvLogs []DvLog) []DvLogTableRow {
	rows := make([]DvLogTableRow, 0, len(dvLogs))
	for i, dvLog := range dvLogs {
		rows = append(rows, DvLogTableRow{
			Number: startNumber + i,
			Cells:  dvLog.Cells(),
			Files:  dvLog.FilesString(),
			ID:     dvLog.ID,
		})
	}
	return rows
}

var baseQuery = `SELECT dv.SURF_DV_LOG_ID, dv.FIRST_FILE, dv.LAST_FILE, w.TITLE, i.NAME, p.NAME, c.NAME, s.CODE, dv.LOG_COMMENT
FROM surf_dv_logs dv 
JOIN workpacks w ON dv.WORKPACK_ID = w.WORKPACK_ID 
JOIN components c ON c.COMPONENT_ID = dv.COMPONENT_ID
JOIN components i ON i.COMPONENT_ID = c.INSTALLATION_ID
LEFT JOIN components p ON p.COMPONENT_ID = c.PARENT_ID
JOIN surf_spreads s ON s.SURF_SPREAD_ID = dv.SURF_SPREAD_ID`

var sortFieldMap = map[string]string{
	"Workpack":     "w.TITLE",
	"Installation": "i.NAME",
	"Substructure": "p.NAME",
	"Component":    "c.NAME",
	"SpreadCode":   "s.CODE",
	"Comment":      "dv.LOG_COMMENT",
}

func GetDvLogs(db *sql.DB, gp *GridParams) ([]DvLog, error) {
	offset := (gp.Page - 1) * gp.NumRows
	query := baseQuery

	if gp.SortBy != nil && *gp.SortBy != "" {
		if sqlCol, ok := sortFieldMap[*gp.SortBy]; ok {
			query += " ORDER BY " + sqlCol
		}
	}

	query += fmt.Sprintf(" LIMIT %d OFFSET %d", gp.NumRows, offset)

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dvLogs []DvLog

	for rows.Next() {
		var row DvLogRow
		if err := rows.Scan(&row.ID, &row.FirstFile, &row.LastFile, &row.Workpack, &row.Installation, &row.Substructure, &row.Component, &row.SpreadCode, &row.Comment); err != nil {
			return nil, err
		}

		dvLog := DvLog{
			ID:           hex.EncodeToString(row.ID),
			Files:        []string{row.FirstFile},
			Workpack:     row.Workpack,
			Installation: row.Installation,
			Substructure: row.Substructure,
			Component:    row.Component,
			SpreadCode:   row.SpreadCode,
			Comment:      row.Comment,
		}

		if row.LastFile != row.FirstFile {
			dvLog.Files = append(dvLog.Files, row.LastFile)

			// TODO: use the file suffix to detect if there are *multiple* files
		}

		dvLogs = append(dvLogs, dvLog)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return dvLogs, nil
}
