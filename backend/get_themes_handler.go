package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func getThemesHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]any{
		"themes":          db.Themes,
		"active_theme_id": db.ActiveThemeID,
	})
}
