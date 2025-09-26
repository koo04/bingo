package main

import (
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Theme struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Items       []string  `json:"items"`
	IsComplete  bool      `json:"is_complete"`
	CreatedAt   time.Time `json:"created_at"`
}

// Get all themes
func getThemes(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]any{
		"themes":          db.Themes,
		"active_theme_id": db.ActiveThemeID,
	})
}

// Set active theme (admin only)
func setActiveTheme(c echo.Context) error {
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
		saveDatabase()
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
			selectedTheme = &theme
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
	saveDatabase()

	// Broadcast theme change to all connected clients
	broadcastUpdate("theme_changed", db.ActiveThemeID)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":         "Active theme updated",
		"active_theme_id": db.ActiveThemeID,
	})
}

// Create new theme (admin only)
func createTheme(c echo.Context) error {
	user := c.Get("user").(*User)
	if !slices.Contains(db.AdminDiscordIDs, user.DiscordID) {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Admin access required"})
	}

	var request struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Items       []string `json:"items"`
	}

	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	if len(request.Items) < 25 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Theme must have at least 25 items"})
	}

	theme := Theme{
		ID:          uuid.New().String(),
		Name:        request.Name,
		Description: request.Description,
		Items:       request.Items,
		CreatedAt:   time.Now(),
	}

	db.Themes = append(db.Themes, theme)
	saveDatabase()

	broadcastUpdate("theme_created", theme)

	return c.JSON(http.StatusCreated, theme)
}

// Update theme (admin only)
func updateTheme(c echo.Context) error {
	user := c.Get("user").(*User)
	if !slices.Contains(db.AdminDiscordIDs, user.DiscordID) {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Admin access required"})
	}

	themeID := c.Param("id")

	var request struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Items       []string `json:"items"`
		IsComplete  *bool    `json:"is_complete,omitempty"` // Pointer to allow null/undefined values
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

			saveDatabase()
			return c.JSON(http.StatusOK, db.Themes[i])
		}
	}

	return c.JSON(http.StatusNotFound, map[string]string{"error": "Theme not found"})
}

// Delete theme (admin only)
func deleteTheme(c echo.Context) error {
	user := c.Get("user").(*User)
	if !slices.Contains(db.AdminDiscordIDs, user.DiscordID) {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Admin access required"})
	}

	defer saveDatabase()

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
	var remainingCards []BingoCard
	for _, card := range db.BingoCards {
		if card.ThemeID != themeID {
			remainingCards = append(remainingCards, card)
		}
	}

	db.BingoCards = remainingCards

	return c.JSON(http.StatusNotFound, map[string]string{"error": "Theme not found"})
}

// Mark theme as complete (admin only)
func markThemeComplete(c echo.Context) error {
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
			saveDatabase()

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

// Get active theme items
func getActiveThemeItems() []string {
	fmt.Printf("DEBUG: Getting active theme items...\n")
	fmt.Printf("DEBUG: ActiveThemeID: %s\n", db.ActiveThemeID)
	fmt.Printf("DEBUG: Number of themes: %d\n", len(db.Themes))

	if db.ActiveThemeID == "" {
		fmt.Printf("DEBUG: No active theme ID set\n")
		return []string{} // Return empty slice instead of fallback
	}

	for _, theme := range db.Themes {
		if theme.ID == db.ActiveThemeID {
			fmt.Printf("DEBUG: Found active theme '%s' with %d items\n", theme.Name, len(theme.Items))
			return theme.Items
		}
	}

	fmt.Printf("DEBUG: Active theme not found\n")
	return []string{} // Return empty slice instead of fallback
}

// Get all unique items from all themes
func getAllThemeItems() []string {
	itemSet := make(map[string]bool)
	var allItems []string

	for _, theme := range db.Themes {
		for _, item := range theme.Items {
			if !itemSet[item] {
				itemSet[item] = true
				allItems = append(allItems, item)
			}
		}
	}

	return allItems
}
