package main

import (
	"cmp"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func handleDiscordCallback(c echo.Context) error {
	code := c.QueryParam("code")
	if code == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Authorization code not provided"})
	}

	token, err := discordOAuth.Exchange(c.Request().Context(), code)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Failed to exchange code for token"})
	}

	// Get user info from Discord
	client := discordOAuth.Client(c.Request().Context(), token)
	resp, err := client.Get("https://discord.com/api/users/@me")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get user info"})
	}
	defer resp.Body.Close()

	var discordUser DiscordUser
	if err := json.NewDecoder(resp.Body).Decode(&discordUser); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to parse user info"})
	}

	// Find or create user
	var user *User
	for i := range db.Users {
		if db.Users[i].DiscordID == discordUser.ID {
			user = &db.Users[i]
			break
		}
	}

	if user == nil {
		newUser := User{
			ID:        uuid.New().String(),
			DiscordID: discordUser.ID,
			Username:  discordUser.Username,
			Avatar:    discordUser.Avatar,
			CreatedAt: time.Now(),
		}
		db.Users = append(db.Users, newUser)
		user = &db.Users[len(db.Users)-1]
		saveDatabase()
	}

	// Generate JWT token
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(), // 7 days
	})

	tokenString, err := jwtToken.SignedString(jwtSecret)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate token"})
	}

	// Redirect to frontend with token
	frontendURL := cmp.Or(os.Getenv("FRONTEND_URL"), "http://localhost:3000/login")
	return c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s?token=%s", frontendURL, tokenString))
}
