package main

import (
	"net/http"
	"slices"
	"strings"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
)

func authMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Authorization header required"})
		}

		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
		}

		claims := token.Claims.(jwt.MapClaims)
		userID := claims["user_id"].(string)

		var user *User
		for i := range db.Users {
			if db.Users[i].ID == userID {
				user = db.Users[i]
				break
			}
		}

		if user == nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "User not found"})
		}

		c.Set("user", user)
		return next(c)
	}
}

// Admin middleware
func adminMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		user := c.Get("user").(*User)

		if slices.Contains(db.AdminDiscordIDs, user.DiscordID) {
			return next(c)
		}

		return c.JSON(http.StatusForbidden, map[string]string{"error": "Admin access required"})
	}
}

// Check admin access
func checkAdminAccess(c echo.Context) error {
	user := c.Get("user").(*User)

	isAdmin := slices.Contains(db.AdminDiscordIDs, user.DiscordID)

	return c.JSON(http.StatusOK, map[string]bool{"is_admin": isAdmin})
}
