package main

import (
	_ "bytes"
	"encoding/base64"
	"time"
)

type DvLog struct {
	ID        []byte    `gorm:"column:SURF_DV_LOG_ID"`
	SpreadID  int       `gorm:"column:SURF_SPREAD_ID"`
	VideoDate time.Time `gorm:"column:VIDEO_DATE"`
	Comment   string    `gorm:"column:LOG_COMMENT"`
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

func DvLogsToGrid(logs []DvLog) Grid {
	g := Grid{}
	g.Columns = []Column{
		{
			Name: "SpreadID",
		},
		{
			Name: "VideoDate",
		},
		{
			Name:     "Comment",
			Editable: true,
		},
	}

	for _, l := range logs {
		g.Rows = append(g.Rows, Row{
			ID: base64.StdEncoding.EncodeToString(l.ID),
			Cells: []any{
				l.SpreadID, l.VideoDate, l.Comment,
			},
		})
	}

	return g
}
