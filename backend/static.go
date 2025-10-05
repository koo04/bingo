package main

import (
	"embed"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v4"
)

//go:embed dist
var staticFiles embed.FS

// getStaticFS returns the embedded static file system
func getStaticFS() fs.FS {
	fsys, err := fs.Sub(staticFiles, "dist")
	if err != nil {
		panic(err)
	}
	return fsys
}

// staticHandler serves static files from the embedded filesystem
func staticHandler(c echo.Context) error {
	fsys := getStaticFS()

	// Get the requested path
	requestPath := c.Request().URL.Path

	// Remove leading slash unconditionally
	requestPath = strings.TrimPrefix(requestPath, "/")

	// If no path specified, serve index.html
	if requestPath == "" {
		requestPath = "index.html"
	}

	// Try to open the file
	file, err := fsys.Open(requestPath)
	if err != nil {
		// If file not found, serve index.html for SPA routing
		file, err = fsys.Open("index.html")
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "File not found")
		}
		requestPath = "index.html"
	}
	defer file.Close()

	// Get file info
	stat, err := file.Stat()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not get file info")
	}

	// If it's a directory, serve index.html
	if stat.IsDir() {
		file.Close()
		file, err = fsys.Open(filepath.Join(requestPath, "index.html"))
		if err != nil {
			file, err = fsys.Open("index.html")
			if err != nil {
				return echo.NewHTTPError(http.StatusNotFound, "File not found")
			}
		}
		defer file.Close()
		stat, _ = file.Stat()
	}

	// Determine content type based on file extension
	contentType := "text/html"
	ext := filepath.Ext(requestPath)
	switch ext {
	case ".js":
		contentType = "application/javascript"
	case ".css":
		contentType = "text/css"
	case ".png":
		contentType = "image/png"
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".gif":
		contentType = "image/gif"
	case ".svg":
		contentType = "image/svg+xml"
	case ".ico":
		contentType = "image/x-icon"
	case ".ttf":
		contentType = "font/ttf"
	case ".woff":
		contentType = "font/woff"
	case ".woff2":
		contentType = "font/woff2"
	case ".eot":
		contentType = "application/vnd.ms-fontobject"
	}

	c.Response().Header().Set("Content-Type", contentType)
	return c.Stream(http.StatusOK, contentType, file)
}
