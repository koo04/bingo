package main

import (
	"fmt"
	"net/http"
	"slices"

	"github.com/labstack/echo/v4"
)

func setThemeCompleteHandler(c echo.Context) error {
	user := c.Get("user").(*User)
	if !slices.Contains(db.AdminDiscordIDs, user.DiscordID) {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Admin access required"})
	}

	themeID := c.Param("id")

	var request struct {
		IsComplete bool `json:"is_complete"`
	}

	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	// Find and update the theme
	for i, theme := range db.Themes {
		if theme.ID == themeID {
			// Don't allow marking the active theme as complete
			if request.IsComplete && db.ActiveThemeID == themeID {
				db.ActiveThemeID = ""
				broadcastUpdate("theme_changed", db.ActiveThemeID)
			}

			db.Themes[i].IsComplete = request.IsComplete

			if err := saveDatabase(); err != nil {
				c.Logger().Error("Error saving database:", err)
			}

			statusText := "incomplete"
			if request.IsComplete {
				statusText = "complete"
			}

			broadcastUpdate("theme_updated", db.Themes[i])

			return c.JSON(http.StatusOK, map[string]any{
				"message": fmt.Sprintf("Theme marked as %s", statusText),
				"theme":   db.Themes[i],
			})
		}
	}

	return c.JSON(http.StatusNotFound, map[string]string{"error": "Theme not found"})
}
