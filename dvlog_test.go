package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetDvLogs(t *testing.T) {
	db, err := openDB("root:root@tcp(localhost:3306)/ie01prod")
	require.NoError(t, err)
	dvLogs, err := GetDvLogs(db, &GridParams{
		Page:    1,
		NumRows: 10,
	})
	require.NoError(t, err)
	require.Len(t, dvLogs, 10)
	// TODO: Assert that all values are populated
}
