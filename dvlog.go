package main

import (
	"database/sql"
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

var baseQuery = `SELECT dv.SURF_DV_LOG_ID AS 'ID', w.TITLE AS 'Workpack', c.NAME AS 'Component'
FROM surf_dv_logs dv 
JOIN workpacks w ON dv.WORKPACK_ID = dv.WORKPACK_ID 
JOIN components c ON c.COMPONENT_ID = dv.COMPONENT_ID`

func GetDvLogs(db *sql.DB, gp *GridParams) ([]DvLog, error) {
	return nil, fmt.Errorf("not implemented")
}
