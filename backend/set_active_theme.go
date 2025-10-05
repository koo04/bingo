package main

import (
	"net/http"
	"slices"

	"github.com/labstack/echo/v4"
)

func setActiveThemeHandler(c echo.Context) error {
	user := c.Get("user").(*User)
	if !slices.Contains(db.AdminDiscordIDs, user.DiscordID) {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Admin access required"})
	}

	var request struct {
		ThemeID string `json:"theme_id"`
	}

	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	if request.ThemeID == "" {
		db.ActiveThemeID = ""
		_ = saveDatabase()
		broadcastUpdate("theme_changed", db.ActiveThemeID)
		return c.JSON(http.StatusOK, map[string]any{
			"message":         "Active theme cleared",
			"active_theme_id": db.ActiveThemeID,
		})
	}

	// Verify theme exists and is not complete
	var selectedTheme *Theme
	for _, theme := range db.Themes {
		if theme.ID == request.ThemeID {
			selectedTheme = theme
			break
		}
	}

	if selectedTheme == nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Theme not found"})
	}

	if selectedTheme.IsComplete {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Cannot set a completed theme as active"})
	}

	db.ActiveThemeID = request.ThemeID
	if err := saveDatabase(); err != nil {
		c.Logger().Error("Error saving database:", err)
	}

	// Broadcast theme change to all connected clients
	broadcastUpdate("theme_changed", db.ActiveThemeID)

	return c.JSON(http.StatusOK, map[string]any{
		"message":         "Active theme updated",
		"active_theme_id": db.ActiveThemeID,
	})
}
