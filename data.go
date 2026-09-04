package main

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/Angus-Warman/stf"
)

func openDB(dsn string) (*sql.DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("no conn string")
	}

	dsn = dsn + "?parseTime=true"

	return sql.Open("mysql", dsn)
}

type EntityDefinition struct {
	TableName string
	IDColumn  string
	Columns   []Column
}

var DvLogDefinition = EntityDefinition{
	TableName: "surf_dv_logs",
	IDColumn:  "SURF_DV_LOG_ID",
	Columns: []Column{
		{
			Name: "SURF_SPREAD_ID",
		},
		{
			Name: "VIDEO_DATE",
		},
		{
			Name:     "LOG_COMMENT",
			Editable: true,
		},
	},
}

type Grid struct {
	Columns []Column
	Rows    []Row
}

func (g *Grid) NumColumns() int {
	return len(g.Columns) + 1 // Include the # column
}

type Column struct {
	Name     string
	Editable bool
}

type Row struct {
	ID      string
	Number  int
	Cells   []any
	Actions any
}

func (e *EntityDefinition) Grid(db *sql.DB, gp GridParams) (*Grid, error) {
	query := e.SelectQuery(gp)

	rows, err := db.Query(query)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	allRows, err := scanRows(rows, len(e.Columns)+1)

	if err != nil {
		return nil, err
	}

	startNumber := ((gp.Page - 1) * gp.NumRows) + 1

	return e.ToGrid(startNumber, allRows), nil
}

func (e *EntityDefinition) SelectQuery(gp GridParams) string {
	offset := (gp.Page - 1) * gp.NumRows
	limit := gp.NumRows

	cols := make([]string, 0, len(e.Columns)+1)
	cols = append(cols, e.IDColumn)
	for _, c := range e.Columns {
		cols = append(cols, c.Name)
	}

	filter := ""
	if gp.Filter != nil {
		filter = "WHERE " + *gp.Filter
	}

	orderBy := ""
	if gp.SortBy != nil {
		orderBy = "ORDER BY " + *gp.SortBy
	}

	return fmt.Sprintf("SELECT %s FROM %s %v %v LIMIT %d OFFSET %d", strings.Join(cols, ", "), e.TableName, filter, orderBy, limit, offset)
}

func (e *EntityDefinition) UpdateCell(db *sql.DB, rowID []byte, colName string, value string) error {
	column, ok := stf.First(e.Columns, func(c Column) bool { return c.Name == colName })

	if !ok {
		return fmt.Errorf("cannot edit %q: column not in definition", colName)
	}

	query := fmt.Sprintf("UPDATE %s SET %s = ? WHERE %s = ?", e.TableName, column.Name, e.IDColumn)

	_, err := db.Exec(query, value, rowID)

	return err
}

func (e *EntityDefinition) VideoPath(db *sql.DB, rowID []byte) (string, error) {
	var subpath, firstFile string

	query := fmt.Sprintf("SELECT VIDEO_SUBPATH, FIRST_FILE FROM %s WHERE %s = ?", e.TableName, e.IDColumn)

	err := db.QueryRow(query, rowID).Scan(&subpath, &firstFile)

	if err != nil {
		return "", err
	}

	return filepath.Join(os.Getenv("VIDEO_ROOT"), subpath, firstFile), nil
}

func (e *EntityDefinition) ToGrid(startNumber int, rows [][]any) *Grid {
	g := &Grid{}
	g.Columns = e.Columns

	for i, row := range rows {
		id := hex.EncodeToString([]byte(fmt.Sprintf("%v", row[0])))
		g.Rows = append(g.Rows, Row{
			Number:  startNumber + i,
			ID:      id,
			Cells:   row[1:],
			Actions: template.HTML(fmt.Sprintf("<button hx-get='/api/play/%v' hx-target='body' hx-swap='beforeend'>Play</button>", id)),
		})
	}

	return g
}

func scanRows(rows *sql.Rows, colCount int) ([][]any, error) {
	var allRows [][]any

	for rows.Next() {
		values := make([]any, colCount)
		dest := make([]any, colCount)
		for i := range values {
			dest[i] = &values[i]
		}

		if err := rows.Scan(dest...); err != nil {
			return nil, err
		}

		for i, v := range values {
			if b, ok := v.([]byte); ok {
				values[i] = string(b)
			}
		}

		allRows = append(allRows, values)
	}

	return allRows, rows.Err()
}
