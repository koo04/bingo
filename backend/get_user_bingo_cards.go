package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func getUserBingoCards(c echo.Context) error {
	user := c.Get("user").(*User)

	var userCards []BingoCard
	for _, card := range db.BingoCards {
		if card.UserID == user.ID {
			userCards = append(userCards, card)
		}
	}

	if len(userCards) == 0 {
		c := newBingoCard(user)
		userCards = append(userCards, c)
	}

	return c.JSON(http.StatusOK, userCards)
}
