package main

import (
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
	FirstFile    string    `db:"FIRST_FILE"`
	LastFile     string    `db:"LAST_FILE"`
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

var baseQuery = `SELECT dv.SURF_DV_LOG_ID,
dv.FIRST_FILE,
dv.LAST_FILE,
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

	return dvLogs, nil
}

func UpdateDvLogComment(db *sqlx.DB, id []byte, comment string) error {
	_, err := db.Exec(`UPDATE surf_dv_logs SET LOG_COMMENT = ? WHERE SURF_DV_LOG_ID = ?`, comment, id)
	return err
}

type VideoData struct {
	VideoSubPath string `db:"VIDEO_SUBPATH"`
	FirstFile    string `db:"FIRST_FILE"`
}

func DvLogVideoPath(db *sqlx.DB, id []byte) (string, error) {
	root := os.Getenv("VIDEO_ROOT")

	if root == "" {
		return "", fmt.Errorf("no video root")
	}

	const query = `SELECT VIDEO_SUBPATH, FIRST_FILE FROM surf_dv_logs WHERE SURF_DV_LOG_ID = ?`

	data := VideoData{}
	err := db.QueryRowx(query, id).StructScan(&data)
	if err != nil {
		return "", err
	}

	return filepath.Join(root, data.VideoSubPath, data.FirstFile), nil
}
