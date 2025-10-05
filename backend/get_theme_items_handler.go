package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func getThemeItemsHandler(c echo.Context) error {
	themeID := c.Param("id")

	theme, found := getThemeByID(themeID)

	if found {
		return c.JSON(http.StatusOK, map[string]any{
			"items": theme.Items,
		})
	}

	return c.JSON(http.StatusNotFound, map[string]string{"error": "Theme not found"})
}
