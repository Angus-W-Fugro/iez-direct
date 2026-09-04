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
	for _, dvLog := range dvLogs {
		require.NotEmpty(t, dvLog.ID)
		require.NotEmpty(t, dvLog.Files)
		require.NotEmpty(t, dvLog.Workpack)
		require.NotEmpty(t, dvLog.Installation)
		require.NotEmpty(t, dvLog.Substructure)
		require.NotEmpty(t, dvLog.Component)
		require.NotEmpty(t, dvLog.SpreadCode)
		require.NotEmpty(t, dvLog.Comment)

		if len(dvLog.Files) > 1 {
			require.NotEqual(t, dvLog.Files[0], dvLog.Files[1], "duplicate files detected, LastFile is often an exact match of FirstFile")
		}
	}
}
