package main

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

func openDB(dsn string) (*sqlx.DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("no conn string")
	}

	dsn = dsn + "?parseTime=true"

	return sqlx.Open("mysql", dsn)
}
