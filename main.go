package main

import (
	"embed"
	"fmt"
	"os"

	"github.com/Angus-Warman/httpmin"
)

func main() {
	err := startServer()

	if err != nil {
		fmt.Fprint(os.Stderr, err)
	}
}

//go:embed all:static
var staticFiles embed.FS

func startServer() error {
	c := httpmin.New()

	h, err := NewHandler()

	if err != nil {
		return err
	}

	c.
		OnPort("7588").
		ServeStatic(staticFiles).
		Route("/dv-logs", h.DvLogsPage).
		Route("/api/dv-logs", h.DvLogsData).
		Route("/api/ping", h.Ping)

	return c.Serve()
}
