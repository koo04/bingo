package main

import (
	"cmp"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"golang.org/x/oauth2"
)

var (
	db           Database
	jwtSecret    = []byte(cmp.Or(os.Getenv("JWT_SECRET"), uuid.New().String()))
	discordOAuth = &oauth2.Config{
		ClientID:     cmp.Or(os.Getenv("DISCORD_CLIENT_ID"), ""),
		ClientSecret: cmp.Or(os.Getenv("DISCORD_CLIENT_SECRET"), ""),
		RedirectURL:  fmt.Sprintf("%s/auth/callback", cmp.Or(os.Getenv("FRONTEND_URL"), "http://localhost:3000")),
		Scopes:       []string{"identify"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://discord.com/api/oauth2/authorize",
			TokenURL: "https://discord.com/api/oauth2/token",
		},
	}
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow connections from any origin
		},
	}
	connections = make(map[*websocket.Conn]bool)
	connMutex   sync.RWMutex
)
