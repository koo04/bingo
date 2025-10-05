package main

import (
	"net/http"
	"slices"

	"github.com/labstack/echo/v4"
)

func updateThemeHandler(c echo.Context) error {
	user := c.Get("user").(*User)
	if !slices.Contains(db.AdminDiscordIDs, user.DiscordID) {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Admin access required"})
	}

	themeID := c.Param("id")

	var request struct {
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Items       []*Item `json:"items"`
		IsComplete  *bool   `json:"is_complete,omitempty"` // Pointer to allow null/undefined values
	}

	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	if len(request.Items) < 25 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Theme must have at least 25 items"})
	}

	// Find and update the theme
	for i, theme := range db.Themes {
		if theme.ID == themeID {
			db.Themes[i].Name = request.Name
			db.Themes[i].Description = request.Description
			db.Themes[i].Items = request.Items

			// Update completion status if provided
			if request.IsComplete != nil {
				db.Themes[i].IsComplete = *request.IsComplete

				// If marking as incomplete and this was the active theme, keep it active
				// If marking as complete and this is the active theme, clear the active theme
				if *request.IsComplete && db.ActiveThemeID == themeID {
					db.ActiveThemeID = ""
				}
			}

			broadcastUpdate("theme_updated", db.Themes[i])

			if err := saveDatabase(); err != nil {
				c.Logger().Error("Error saving database:", err)
			}

			return c.JSON(http.StatusOK, db.Themes[i])
		}
	}

	return c.JSON(http.StatusNotFound, map[string]string{"error": "Theme not found"})
}
