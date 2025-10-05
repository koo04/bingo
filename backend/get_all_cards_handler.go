package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func getAllCardsHandler(c echo.Context) error {
	type request struct {
		ThemeID string `json:"theme_id" param:"id"`
	}

	var req request
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	if req.ThemeID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Theme ID is required"})
	}

	theme, found := getThemeByID(req.ThemeID)
	if !found {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Theme not found"})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"cards": theme.Cards,
		"users": db.Users,
	})
}
