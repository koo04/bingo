package main

import (
	"log/slog"
	"net/http"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func createThemeHandler(c echo.Context) error {
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
