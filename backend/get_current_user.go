package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func getCurrentUser(c echo.Context) error {
	user := c.Get("user").(*User)
	return c.JSON(http.StatusOK, user)
}
