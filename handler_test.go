package main

import (
	"testing"

	"github.com/Angus-Warman/httpmin"
	"github.com/stretchr/testify/require"
)

func TestGetData(t *testing.T) {
	httpmin.LoadEnvFile()
	h, err := NewHandler()
	require.NoError(t, err)
	g, err := h.getDvLogsData(GridParams{Page: 1, NumRows: 15})
	require.NoError(t, err)
	require.Len(t, g.Rows, 15)
}
