package main

import (
	"net/http"
	"slices"

	"github.com/labstack/echo/v4"
)

func checkAdminAccessHandler(c echo.Context) error {
	user := c.Get("user").(*User)

	isAdmin := slices.Contains(db.AdminDiscordIDs, user.DiscordID)

	return c.JSON(http.StatusOK, map[string]bool{"is_admin": isAdmin})
}
