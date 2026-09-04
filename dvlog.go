package main

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

type DvLogRow struct {
	ID           []byte    `db:"SURF_DV_LOG_ID"`
	Workpack     string    `db:"TITLE"`
	Installation string    `db:"INSTALLATION"`
	Substructure string    `db:"SUBSTRUCTURE"`
	Component    string    `db:"COMPONENT"`
	SpreadCode   string    `db:"CODE"`
	Date         time.Time `db:"VIDEO_DATE"`
	Comment      string    `db:"LOG_COMMENT"`
}

type DvLog struct {
	ID           string // Base64-encoded
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

var baseQuery = `SELECT dv.SURF_DV_LOG_ID,
w.TITLE,
i.NAME AS INSTALLATION,
p.NAME AS SUBSTRUCTURE,
c.NAME AS COMPONENT,
s.CODE,
dv.VIDEO_DATE,
dv.LOG_COMMENT
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

func GetDvLogs(db *sqlx.DB, gp *GridParams) ([]DvLog, error) {
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

	var rows []DvLogRow
	if err := db.Select(&rows, query); err != nil {
		return nil, err
	}

	dvLogs := make([]DvLog, 0, len(rows))

	for _, row := range rows {
		dvLog := DvLog{
			ID:           hex.EncodeToString(row.ID),
			Workpack:     row.Workpack,
			Installation: row.Installation,
			Substructure: row.Substructure,
			Component:    row.Component,
			SpreadCode:   row.SpreadCode,
			Date:         row.Date,
			Comment:      row.Comment,
		}

		dvLogs = append(dvLogs, dvLog)
	}

	return dvLogs, nil
}

func UpdateDvLogComment(db *sqlx.DB, id []byte, comment string) error {
	_, err := db.Exec(`UPDATE surf_dv_logs SET LOG_COMMENT = ? WHERE SURF_DV_LOG_ID = ?`, comment, id)
	return err
}

type VideoDataRow struct {
	LogID        []byte         `db:"SURF_DV_LOG_ID"`
	VideoSubPath string         `db:"VIDEO_SUBPATH"`
	FirstFile    string         `db:"FIRST_FILE"`
	LastFile     sql.NullString `db:"LAST_FILE"`
}

type VideoData struct {
	LogID        string // Hex-encoded
	VideoSubPath string
	Files        []string
}

func GetVideoData(db *sqlx.DB, id []byte) (*VideoData, error) {
	const query = `SELECT SURF_DV_LOG_ID, VIDEO_SUBPATH, FIRST_FILE, LAST_FILE
FROM surf_dv_logs WHERE SURF_DV_LOG_ID = ?`

	var row VideoDataRow
	if err := db.Get(&row, query, id); err != nil {
		return nil, err
	}

	data := &VideoData{
		LogID:        hex.EncodeToString(row.LogID),
		VideoSubPath: row.VideoSubPath,
		Files:        []string{row.FirstFile},
	}

	if row.LastFile.Valid && row.LastFile.String != row.FirstFile {
		data.Files = append(data.Files, row.LastFile.String)

		// TODO: Use suffix to check for additional files
	}

	return data, nil
}

func (v *VideoData) FilePath(idx int) (string, error) {
	root := os.Getenv("VIDEO_ROOT")

	if root == "" {
		return "", fmt.Errorf("no video root")
	}

	if idx < 0 || idx >= len(v.Files) {
		return "", fmt.Errorf("file index %d out of range", idx)
	}

	return filepath.Join(root, v.VideoSubPath, v.Files[idx]), nil
}
