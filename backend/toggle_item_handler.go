package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func toggleItemHandler(c echo.Context) error {
	var req struct {
		ThemeID string `json:"theme_id" param:"themeId"`
		ItemId  string `json:"item_id" param:"itemId"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	theme, found := getThemeByID(req.ThemeID)
	if !found {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Theme not found"})
	}

	item, found := theme.GetItem(req.ItemId)
	if !found {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Item not found"})
	}

	item.Marked = !item.Marked

	winners := theme.checkForWinners()
	if len(winners) > 0 {
		// Broadcast winner
		broadcastUpdate("winners", map[string]any{
			"cards": winners,
		})
	}

	if err := saveDatabase(); err != nil {
		c.Logger().Error("Error saving database:", err)
	}

	// Broadcast to all WebSocket connections
	broadcastUpdate("item_updated", item)

	return c.JSON(http.StatusOK, map[string]string{"status": "marked"})
}
