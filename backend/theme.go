package main

import (
	"fmt"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Theme struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Items       []*Item          `json:"items"`
	IsComplete  bool             `json:"is_complete"`
	Cards       map[string]*Card `json:"cards"`
	CreatedAt   time.Time        `json:"created_at"`
}

// Get theme by ID
func getThemeByID(themeID string) (*Theme, bool) {
	for _, theme := range db.Themes {
		if theme.ID == themeID {
			return theme, true
		}
	}
	return nil, false
}

func (t *Theme) GetItem(itemID string) (*Item, bool) {
	for _, item := range t.Items {
		if item.ID == itemID {
			return item, true
		}
	}
	return nil, false
}

func (t *Theme) NewBingoCard(user *User) (*Card, error) {
	if t.IsComplete {
		return nil, fmt.Errorf("cannot generate card from a completed theme")
	}

	if len(t.Items) < 25 {
		return nil, fmt.Errorf("theme has insufficient items: need at least 25, have %d", len(t.Items))
	}

	// Create a new bingo card
	card := &Card{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		ThemeID:   t.ID,
		Items:     make([][]string, 5),
		CreatedAt: time.Now(),
		IsWinner:  false,
	}

	// Shuffle and select 25 items
	shuffled := make([]*Item, len(t.Items))
	copy(shuffled, t.Items)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})
	selected := shuffled[:25]

	// Fill the 5x5 grid
	for i := range card.Items {
		card.Items[i] = make([]string, 5)
		for j := range card.Items[i] {
			card.Items[i][j] = selected[i*5+j].ID
		}
	}

	// Free space in the middle
	card.Items[2][2] = "FREE_SPACE"

	if t.Cards == nil {
		t.Cards = make(map[string]*Card)
	}
	t.Cards[user.ID] = card

	return card, saveDatabase()
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
		slog.Error("Failed to bind createTheme request", "error", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	if len(request.Items) < 25 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Theme must have at least 25 items"})
	}

	items := make([]*Item, len(request.Items))
	for i, itemName := range request.Items {
		items[i] = &Item{
			ID:   uuid.New().String(),
			Name: itemName,
		}
	}

	theme := &Theme{
		ID:          uuid.New().String(),
		Name:        request.Name,
		Description: request.Description,
		Items:       items,
		Cards:       make(map[string]*Card),
		CreatedAt:   time.Now(),
	}

	db.Themes = append(db.Themes, theme)
	if err := saveDatabase(); err != nil {
		c.Logger().Error("Error saving database:", err)
	}

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

// Delete theme (admin only)
func deleteTheme(c echo.Context) error {
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

func getThemeItemsRequest(c echo.Context) error {
	themeID := c.Param("id")

	theme, found := getThemeByID(themeID)

	if found {
		return c.JSON(http.StatusOK, map[string]any{
			"items": theme.Items,
		})
	}

	return c.JSON(http.StatusNotFound, map[string]string{"error": "Theme not found"})
}
