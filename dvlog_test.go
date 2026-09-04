package main

import (
	"encoding/hex"
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
		require.NotEmpty(t, dvLog.Workpack)
		require.NotEmpty(t, dvLog.Installation)
		require.NotEmpty(t, dvLog.Substructure)
		require.NotEmpty(t, dvLog.Component)
		require.NotEmpty(t, dvLog.SpreadCode)
		require.NotEmpty(t, dvLog.Comment)
	}
}

func TestGetVideoData(t *testing.T) {
	t.Setenv("VIDEO_ROOT", "/videos")

	db, err := openDB("root:root@tcp(localhost:3306)/ie01prod")
	require.NoError(t, err)

	dvLogs, err := GetDvLogs(db, &GridParams{Page: 1, NumRows: 1})
	require.NoError(t, err)
	require.Len(t, dvLogs, 1)

	id, err := hex.DecodeString(dvLogs[0].ID)
	require.NoError(t, err)

	data, err := GetVideoData(db, id)
	require.NoError(t, err)
	require.NotEmpty(t, data.Files)
	require.Equal(t, dvLogs[0].ID, data.LogID)

	if len(data.Files) > 1 {
		require.NotEqual(t, data.Files[0], data.Files[1], "duplicate files detected, LastFile is often an exact match of FirstFile")
	}

	path, err := data.FilePath(0)
	require.NoError(t, err)
	require.NotEmpty(t, path)
}
