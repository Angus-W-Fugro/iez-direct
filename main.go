package main

import (
	"embed"
	"fmt"
	"os"

	"github.com/Angus-Warman/httpmin"
	"github.com/Angus-Warman/httpmin/theme"
)

func main() {
	err := startServer()

	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		fmt.Println("Press enter to close...")
		fmt.Scanln()
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
		RouteHandler("/theme.css", theme.Modern()).
		ServeStatic(staticFiles).
		Route("/dv-logs", h.DvLogsPage).
		Route("/api/dv-logs", h.DvLogsData).
		Route("PATCH /api/dv-log/edit", h.EditDvLogCell).
		Route("/api/play/{id}", h.Play).
		Route("/api/media/{logID}/{fileIdx}", h.Media).
		Route("/workpacks", h.WorkpacksPage).
		Route("/", h.IndexPage)

	return c.Serve()
}
