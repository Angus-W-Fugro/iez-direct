package main

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/Angus-Warman/stf"
)

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

type Column struct {
	Name     string
	Editable bool
}

type Row struct {
	ID     string
	Number int
	Cells  []any
}

func (e *EntityDefinition) Grid(db *sql.DB, gp GridParams) (Grid, error) {
	query := DvLogDefinition.SelectQuery(gp)

	rows, err := db.Query(query)

	if err != nil {
		return Grid{}, err
	}

	defer rows.Close()

	allRows, err := scanRows(rows, len(DvLogDefinition.Columns)+1)

	if err != nil {
		return Grid{}, err
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

	return fmt.Sprintf("SELECT %s FROM %s %v LIMIT %d OFFSET %d", strings.Join(cols, ", "), e.TableName, filter, limit, offset)
}

func (e *EntityDefinition) UpdateCell(db *sql.DB, rowID []byte, colName string, value string) error {
	column, _ := stf.First(e.Columns, func(c Column) bool { return c.Name == colName })

	query := fmt.Sprintf("UPDATE %s SET %s = ? WHERE %s = ?", e.TableName, column.Name, e.IDColumn)

	_, err := db.Exec(query, value, rowID)

	return err
}

func (e *EntityDefinition) ToGrid(startNumber int, rows [][]any) Grid {
	g := Grid{}
	g.Columns = e.Columns

	for i, row := range rows {
		id := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%v", row[0])))
		g.Rows = append(g.Rows, Row{
			Number: startNumber + i,
			ID:     id,
			Cells:  row[1:],
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
