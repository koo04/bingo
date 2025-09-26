package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/google/uuid"
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

	themeItems := getActiveThemeItems()
	if len(themeItems) < 25 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": fmt.Sprintf("Active theme only has %d items (need at least 25). Please contact an administrator to add more items to the theme.", len(themeItems)),
		})
	}

	card, err := newBingoCard(user)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, card)
}

func newBingoCard(user *User) (BingoCard, error) {
	// Get items from active theme
	themeItems := getActiveThemeItems()

	if len(themeItems) < 25 {
		return BingoCard{}, fmt.Errorf("insufficient items: need 25, have %d", len(themeItems))
	}

	// Shuffle and select 25 items
	shuffled := make([]string, len(themeItems))
	copy(shuffled, themeItems)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	// Create 5x5 grid
	items := make([][]string, 5)
	markedItems := make([][]bool, 5)
	for i := range items {
		items[i] = make([]string, 5)
		markedItems[i] = make([]bool, 5)
		for j := range items[i] {
			items[i][j] = shuffled[i*5+j]
		}
	}

	// Free space in the middle
	items[2][2] = "FREE SPACE"
	markedItems[2][2] = true

	card := BingoCard{
		ID:          uuid.New().String(),
		UserID:      user.ID,
		ThemeID:     db.ActiveThemeID,
		Items:       items,
		MarkedItems: markedItems,
		CreatedAt:   time.Now(),
		IsWinner:    false,
	}

	db.BingoCards = append(db.BingoCards, card)
	saveDatabase()

	return card, nil
}
