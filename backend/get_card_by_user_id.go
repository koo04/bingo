package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func getCardByUserIdHandler(c echo.Context) error {
	user := c.Get("user").(*User)
	themeID := c.Param("id")

	// not found, generate a new one
	theme, found := getThemeByID(themeID)
	if !found {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Theme not found",
		})
	}

	if card, ok := theme.Cards[user.ID]; ok {
		return c.JSON(http.StatusOK, card)
	}

	card, err := theme.NewCard(user)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	if err := saveDatabase(); err != nil {
		c.Logger().Error("Error saving database:", err)
	}

	return c.JSON(http.StatusOK, card)
}
