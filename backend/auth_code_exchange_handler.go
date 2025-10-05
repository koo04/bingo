package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func authCodeExchangeHandler(c echo.Context) error {
	var req AuthCodeRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	if req.Code == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Authorization code is required",
		})
	}

	// Exchange code for token
	token, err := discordOAuth.Exchange(c.Request().Context(), req.Code)
	if err != nil {
		log.Println("Failed to exchange authorization code:", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Failed to exchange authorization code",
		})
	}

	// Get user info from Discord
	client := discordOAuth.Client(c.Request().Context(), token)
	resp, err := client.Get("https://discord.com/api/users/@me")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get user information from Discord",
		})
	}
	defer resp.Body.Close()

	var discordUser DiscordUser
	if err := json.NewDecoder(resp.Body).Decode(&discordUser); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to parse user information",
		})
	}

	// Find or create user
	var user *User
	for i := range db.Users {
		if db.Users[i].DiscordID == discordUser.ID {
			user = db.Users[i]
			break
		}
	}

	if user == nil {
		newUser := &User{
			ID:        uuid.New().String(),
			DiscordID: discordUser.ID,
			Username:  discordUser.Username,
			Avatar:    discordUser.Avatar,
			CreatedAt: time.Now(),
		}
		db.Users = append(db.Users, newUser)
		user = db.Users[len(db.Users)-1]
		if err := saveDatabase(); err != nil {
			c.Logger().Error("Failed to save database:", err)
		}
	}

	// Generate JWT token
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(), // 7 days
	})

	tokenString, err := jwtToken.SignedString(jwtSecret)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to generate authentication token",
		})
	}

	// Return the token and user info
	return c.JSON(http.StatusOK, AuthResponse{
		Token: tokenString,
		User:  *user,
	})
}
