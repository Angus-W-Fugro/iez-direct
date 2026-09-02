package main

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"strings"
)

type EntityDefinition struct {
	TableName string
	IDColumn  string
	Columns   []ColumnDefinition
}

type ColumnDefinition struct {
	Label    string
	Name     string
	Editable bool
}

var DvLogDefinition = EntityDefinition{
	TableName: "surf_dv_logs",
	IDColumn:  "SURF_DV_LOG_ID",
	Columns: []ColumnDefinition{
		{
			Name:  "SURF_SPREAD_ID",
			Label: "SpreadID",
		},
		{
			Name:  "VIDEO_DATE",
			Label: "VideoDate",
		},
		{
			Name:     "LOG_COMMENT",
			Label:    "Comment",
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
	ID    string
	Cells []any
}

func (e *EntityDefinition) Grid(db *sql.DB) (Grid, error) {
	query := DvLogDefinition.SelectQuery(10)

	rows, err := db.Query(query)

	if err != nil {
		return Grid{}, err
	}

	defer rows.Close()

	allRows, err := scanRows(rows, len(DvLogDefinition.Columns)+1)

	if err != nil {
		return Grid{}, err
	}

	return e.ToGrid(allRows), nil
}

func (e *EntityDefinition) SelectQuery(limit int) string {
	cols := make([]string, 0, len(e.Columns)+1)
	cols = append(cols, e.IDColumn)
	for _, c := range e.Columns {
		cols = append(cols, c.Name)
	}
	return fmt.Sprintf("SELECT %s FROM %s LIMIT %d", strings.Join(cols, ", "), e.TableName, limit)
}

func (e *EntityDefinition) UpdateCell(db *sql.DB, rowID []byte, column string, value string) error {
	query := fmt.Sprintf("UPDATE %s SET %s = ? WHERE %s = ?", e.TableName, column, e.IDColumn)

	_, err := db.Exec(query, value, rowID)

	return err
}

func (e *EntityDefinition) ToGrid(rows [][]any) Grid {
	g := Grid{}
	g.Columns = make([]Column, len(e.Columns))
	for i, c := range e.Columns {
		g.Columns[i] = Column{Name: c.Label, Editable: c.Editable}
	}

	for _, row := range rows {
		id := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%v", row[0])))
		g.Rows = append(g.Rows, Row{
			ID:    id,
			Cells: row[1:],
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

		allRows = append(allRows, values)
	}

	return allRows, rows.Err()
}
