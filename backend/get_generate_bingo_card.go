package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func generateNewBingoCard(c echo.Context) error {
	user := c.Get("user").(*User)

	// Check if there are any themes available
	if len(db.Themes) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "No themes available. Please contact an administrator to create themes.",
		})
	}

	// Check if there's an active theme
	if db.ActiveThemeID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "No active theme selected. Please contact an administrator to set an active theme.",
		})
	}

	theme, found := getThemeByID(db.ActiveThemeID)
	if !found {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Active theme not found. Please contact an administrator to set a valid active theme.",
		})
	}

	card, err := theme.NewBingoCard(user)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, card)
}
