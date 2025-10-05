package main

import (
	"net/http"
	"slices"

	"github.com/labstack/echo/v4"
)

func deleteThemeHandler(c echo.Context) error {
	user := c.Get("user").(*User)
	if !slices.Contains(db.AdminDiscordIDs, user.DiscordID) {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Admin access required"})
	}

	defer func() {
		if err := saveDatabase(); err != nil {
			c.Logger().Error("Error saving database:", err)
		}
	}()

	themeID := c.Param("id")

	// Don't allow deleting the active theme
	if db.ActiveThemeID == themeID {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Cannot delete the active theme. Set a different theme as active first."})
	}

	// Find and remove the theme
	for i, theme := range db.Themes {
		if theme.ID == themeID {
			// Store theme info for broadcast before deletion
			deletedTheme := theme
			// Remove theme from slice
			db.Themes = append(db.Themes[:i], db.Themes[i+1:]...)

			broadcastUpdate("theme_deleted", map[string]any{
				"id":   deletedTheme.ID,
				"name": deletedTheme.Name,
			})

			return c.JSON(http.StatusOK, map[string]string{"message": "Theme deleted successfully"})
		}
	}

	// Remove all cards associated with the deleted theme
	var remainingCards []*Card
	for _, card := range db.BingoCards {
		if card.ThemeID != themeID {
			remainingCards = append(remainingCards, card)
		}
	}

	db.BingoCards = remainingCards

	return c.JSON(http.StatusNotFound, map[string]string{"error": "Theme not found"})
}
