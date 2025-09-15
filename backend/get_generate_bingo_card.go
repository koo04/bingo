package main

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func generateNewBingoCard(c echo.Context) error {
	user := c.Get("user").(*User)

	if len(bingoItems) < 25 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Not enough bingo items available"})
	}

	// Shuffle and select 25 items
	shuffled := make([]string, len(bingoItems))
	copy(shuffled, bingoItems)
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
		Items:       items,
		MarkedItems: markedItems,
		CreatedAt:   time.Now(),
		IsWinner:    false,
	}

	db.BingoCards = append(db.BingoCards, card)
	saveDatabase()

	return c.JSON(http.StatusOK, card)
}
