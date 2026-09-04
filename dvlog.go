package main

import (
	"database/sql"
	"encoding/hex"
	"fmt"
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
	ID           string // Hex-encoded
	Files        []string
	Workpack     string
	Installation string
	Substructure string
	Component    string
	SpreadCode   string
	Comment      string
}

var baseQuery = `SELECT dv.SURF_DV_LOG_ID, dv.FIRST_FILE, dv.LAST_FILE, w.TITLE, i.NAME, p.NAME, c.NAME, s.CODE, dv.LOG_COMMENT
FROM surf_dv_logs dv 
JOIN workpacks w ON dv.WORKPACK_ID = w.WORKPACK_ID 
JOIN components c ON c.COMPONENT_ID = dv.COMPONENT_ID
JOIN components i ON i.COMPONENT_ID = c.INSTALLATION_ID
LEFT JOIN components p ON p.COMPONENT_ID = c.PARENT_ID
JOIN surf_spreads s ON s.SURF_SPREAD_ID = dv.SURF_SPREAD_ID`

func GetDvLogs(db *sql.DB, gp *GridParams) ([]DvLog, error) {
	offset := (gp.Page - 1) * gp.NumRows
	query := baseQuery + fmt.Sprintf(" LIMIT %d OFFSET %d", gp.NumRows, offset)

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
