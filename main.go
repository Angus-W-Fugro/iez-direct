package main

import (
	"fmt"
	"os"

	"github.com/Angus-Warman/httpmin"
	"github.com/Angus-Warman/httpmin/theme"
)

func main() {
	err := startServer()

	if err != nil {
		fmt.Fprint(os.Stderr, err)
	}
}

func startServer() error {
	c := httpmin.New()

	h, err := NewHandler()

	if err != nil {
		return err
	}

	c.
		OnPort("7588").
		RouteHandler("/styles.css", theme.Minimal()).
		Route("/dv-logs", h.DvLogsPage).
		Route("/api/dv-logs", h.DvLogsData).
		Route("/api/ping", h.Ping)

	return c.Serve()
}
