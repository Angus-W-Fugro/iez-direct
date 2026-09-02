package main

import "time"

type DvLog struct {
	ID        string    `gorm:"column:SURF_DV_LOG_ID"`
	SpreadID  int       `gorm:"column:SURF_SPREAD_ID"`
	VideoDate time.Time `gorm:"column:VIDEO_DATE"`
	Comment   string    `gorm:"column:LOG_COMMENT"`
}

type Grid struct {
	Columns []Column
	Rows    []Row
}

type Column struct {
	Name string
}

type Row struct {
	Cells []any
}

func DvLogsToGrid(logs []DvLog) Grid {
	g := Grid{}
	g.Columns = []Column{
		{
			Name: "ID",
		},
		{
			Name: "SpreadID",
		},
		{
			Name: "VideoDate",
		},
		{
			Name: "Comment",
		},
	}

	for _, l := range logs {
		g.Rows = append(g.Rows, Row{
			Cells: []any{
				l.ID, l.SpreadID, l.VideoDate, l.Comment,
			},
		})
	}

	return g
}
