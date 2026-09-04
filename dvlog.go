package main

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
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
	Date         time.Time
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
	Date         time.Time
	Comment      string
}

var DvLogColumns = []string{
	"Workpack",
	"Installation",
	"Substructure",
	"Component",
	"SpreadCode",
	"Date",
}

type DvLogTableRow struct {
	Number  int
	Cells   []string
	Comment string
	ID      string
}

type DvLogPageData struct {
	Rows        []DvLogTableRow
	Columns     []string
	SortOptions []string
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
		d.Date.Format("2006/01/02 03:04:05"),
	}
}

func toTableRows(startNumber int, dvLogs []DvLog) []DvLogTableRow {
	rows := make([]DvLogTableRow, 0, len(dvLogs))
	for i, dvLog := range dvLogs {
		rows = append(rows, DvLogTableRow{
			Number:  startNumber + i,
			Cells:   dvLog.Cells(),
			Comment: dvLog.Comment,
			ID:      dvLog.ID,
		})
	}
	return rows
}

var baseQuery = `SELECT dv.SURF_DV_LOG_ID, dv.FIRST_FILE, dv.LAST_FILE, w.TITLE, i.NAME, p.NAME, c.NAME, s.CODE, dv.VIDEO_DATE, dv.LOG_COMMENT
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
	"Date":         "dv.VIDEO_DATE",
	"Comment":      "dv.LOG_COMMENT",
}

func GetDvLogs(db *sql.DB, gp *GridParams) ([]DvLog, error) {
	offset := (gp.Page - 1) * gp.NumRows
	query := baseQuery

	if gp.SortBy != nil && *gp.SortBy != "" {
		parts := strings.Split(*gp.SortBy, " ")
		field := parts[0]

		if sqlCol, ok := sortFieldMap[field]; ok {
			query += " ORDER BY " + sqlCol

			if len(parts) > 1 && (parts[1] == "DESC" || parts[1] == "desc") {
				query += " DESC"
			}
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
		if err := rows.Scan(&row.ID, &row.FirstFile, &row.LastFile, &row.Workpack, &row.Installation, &row.Substructure, &row.Component, &row.SpreadCode, &row.Date, &row.Comment); err != nil {
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
			Date:         row.Date,
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

func UpdateDvLogComment(db *sql.DB, id []byte, comment string) error {
	_, err := db.Exec(`UPDATE surf_dv_logs SET LOG_COMMENT = ? WHERE SURF_DV_LOG_ID = ?`, comment, id)
	return err
}
