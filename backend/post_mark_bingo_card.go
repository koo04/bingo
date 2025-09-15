package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func markBingoItem(c echo.Context) error {
	user := c.Get("user").(*User)

	var request struct {
		CardID string `json:"card_id"`
		Row    int    `json:"row"`
		Col    int    `json:"col"`
	}

	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	// Find the card
	var cardIndex = -1
	for i, card := range db.BingoCards {
		if card.ID == request.CardID && card.UserID == user.ID {
			cardIndex = i
			break
		}
	}

	if cardIndex == -1 {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Card not found"})
	}

	// Validate coordinates
	if request.Row < 0 || request.Row >= 5 || request.Col < 0 || request.Col >= 5 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid coordinates"})
	}

	// Mark the item
	db.BingoCards[cardIndex].MarkedItems[request.Row][request.Col] = !db.BingoCards[cardIndex].MarkedItems[request.Row][request.Col]

	// Check for bingo
	db.BingoCards[cardIndex].IsWinner = checkBingo(db.BingoCards[cardIndex].MarkedItems)

	saveDatabase()

	return c.JSON(http.StatusOK, db.BingoCards[cardIndex])
}
