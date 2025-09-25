package main

import (
	"cmp"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/oauth2"
)

var (
	db           Database
	jwtSecret    = []byte(cmp.Or(os.Getenv("JWT_SECRET"), uuid.New().String()))
	discordOAuth = &oauth2.Config{
		ClientID:     cmp.Or(os.Getenv("DISCORD_CLIENT_ID"), ""),
		ClientSecret: cmp.Or(os.Getenv("DISCORD_CLIENT_SECRET"), ""),
		RedirectURL:  cmp.Or(os.Getenv("DISCORD_REDIRECT_URL"), "http://localhost:8080/auth/discord/callback"),
		Scopes:       []string{"identify"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://discord.com/api/oauth2/authorize",
			TokenURL: "https://discord.com/api/oauth2/token",
		},
	}
	bingoItems []string
	upgrader   = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow connections from any origin
		},
	}
	connections = make(map[*websocket.Conn]bool)
	connMutex   sync.RWMutex
)

func main() {
	loadDatabase()
	loadBingoItems()

	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))

	// Routes
	e.GET("/auth/discord", handleDiscordAuth)
	e.GET("/auth/discord/callback", handleDiscordCallback)
	e.GET("/api/user", getCurrentUser, authMiddleware)
	e.GET("/api/bingo/new", generateNewBingoCard, authMiddleware)
	e.GET("/api/bingo/cards", getUserBingoCards, authMiddleware)
	e.POST("/api/bingo/mark", markBingoItem, authMiddleware)

	// Admin routes
	e.GET("/api/admin/check", checkAdminAccess, authMiddleware)
	e.GET("/api/admin/items", getGlobalMarkedItems, authMiddleware, adminMiddleware)
	e.POST("/api/admin/items/mark", markGlobalItem, authMiddleware, adminMiddleware)
	e.POST("/api/admin/items/unmark", unmarkGlobalItem, authMiddleware, adminMiddleware)
	e.GET("/api/admin/cards", getAllBingoCards, authMiddleware, adminMiddleware)

	// WebSocket endpoint
	e.GET("/ws", handleWebSocket)

	e.GET("/api/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	port := cmp.Or(os.Getenv("PORT"), "8080")
	log.Printf("Server starting on port %s", port)
	e.Logger.Fatal(e.Start(":" + port))
}

func loadDatabase() {
	data, err := os.ReadFile("data/database.json")
	if err != nil {
		log.Println("Database file not found, creating new one")
		db = Database{
			Users:             []User{},
			BingoCards:        []BingoCard{},
			AdminDiscordIDs:   []string{}, // Add admin Discord IDs here manually or via environment
			GlobalMarkedItems: []string{},
		}
		saveDatabase()
		return
	}

	if err := json.Unmarshal(data, &db); err != nil {
		log.Fatal("Error parsing database:", err)
	}

	// Initialize fields if they don't exist
	if db.AdminDiscordIDs == nil {
		db.AdminDiscordIDs = []string{}
	}
	if db.GlobalMarkedItems == nil {
		db.GlobalMarkedItems = []string{}
	}

	// Add admin IDs from environment variable
	if envAdminIDs := os.Getenv("ADMIN_DISCORD_IDS"); envAdminIDs != "" {
		envIDs := strings.Split(envAdminIDs, ",")
		for _, id := range envIDs {
			id = strings.TrimSpace(id)
			if id != "" {
				// Check if already exists
				exists := false
				for _, existingID := range db.AdminDiscordIDs {
					if existingID == id {
						exists = true
						break
					}
				}
				if !exists {
					db.AdminDiscordIDs = append(db.AdminDiscordIDs, id)
				}
			}
		}
		saveDatabase()
	}
}

func saveDatabase() {
	os.MkdirAll("data", 0755)
	data, _ := json.MarshalIndent(db, "", "  ")
	os.WriteFile("data/database.json", data, 0644)
}

func loadBingoItems() {
	data, err := os.ReadFile("data/bingo-items.txt")
	if err != nil {
		log.Println("Bingo items file not found, creating sample file")
		sampleItems := []string{
			"Someone mentions the weather",
			"Technical difficulties occur",
			"Someone is eating on camera",
			"Pet appears on screen",
			"Someone forgets to mute",
			"\"Can you hear me?\" is said",
			"Someone has virtual background",
			"Audio feedback happens",
			"Someone joins late",
			"Internet connection issues",
			"Someone drinks coffee",
			"Phone rings in background",
			"\"You're on mute\" is said",
			"Someone waves at camera",
			"Screen sharing problems",
			"Someone types loudly",
			"Background noise interruption",
			"Someone leaves early",
			"Technology joke is made",
			"Someone asks to repeat something",
			"Video freezes",
			"Someone multitasks visibly",
			"Awkward silence occurs",
			"Someone sneezes or coughs",
			"Meeting runs over time",
		}

		os.MkdirAll("data", 0755)
		content := strings.Join(sampleItems, "\n")
		os.WriteFile("data/bingo-items.txt", []byte(content), 0644)
		bingoItems = sampleItems
		return
	}

	content := string(data)
	bingoItems = strings.Split(strings.TrimSpace(content), "\n")

	// Filter out empty lines
	var filtered []string
	for _, item := range bingoItems {
		if strings.TrimSpace(item) != "" {
			filtered = append(filtered, strings.TrimSpace(item))
		}
	}
	bingoItems = filtered
}

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
				user = &db.Users[i]
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

func checkBingo(marked [][]bool) bool {
	// Check rows
	for i := 0; i < 5; i++ {
		if marked[i][0] && marked[i][1] && marked[i][2] && marked[i][3] && marked[i][4] {
			return true
		}
	}

	// Check columns
	for j := 0; j < 5; j++ {
		if marked[0][j] && marked[1][j] && marked[2][j] && marked[3][j] && marked[4][j] {
			return true
		}
	}

	// Check diagonals
	if marked[0][0] && marked[1][1] && marked[2][2] && marked[3][3] && marked[4][4] {
		return true
	}
	if marked[0][4] && marked[1][3] && marked[2][2] && marked[3][1] && marked[4][0] {
		return true
	}

	return false
}

// Admin middleware
func adminMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		user := c.Get("user").(*User)

		for _, adminID := range db.AdminDiscordIDs {
			if user.DiscordID == adminID {
				return next(c)
			}
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

// Get globally marked items
func getGlobalMarkedItems(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"marked_items": db.GlobalMarkedItems,
		"all_items":    bingoItems,
	})
}

// Mark global item
func markGlobalItem(c echo.Context) error {
	var req struct {
		Item string `json:"item"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	// Check if item exists in bingo items
	itemExists := false
	for _, item := range bingoItems {
		if item == req.Item {
			itemExists = true
			break
		}
	}

	if !itemExists {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Item not found"})
	}

	// Check if already marked
	for _, markedItem := range db.GlobalMarkedItems {
		if markedItem == req.Item {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Item already marked"})
		}
	}

	db.GlobalMarkedItems = append(db.GlobalMarkedItems, req.Item)

	// Mark this item on all player cards that contain it
	for i := range db.BingoCards {
		card := &db.BingoCards[i]
		for row := 0; row < 5; row++ {
			for col := 0; col < 5; col++ {
				if card.Items[row][col] == req.Item {
					card.MarkedItems[row][col] = true
					// Check if this creates a bingo
					if checkBingo(card.MarkedItems) {
						card.IsWinner = true
					}
				}
			}
		}
	}

	saveDatabase()

	// Broadcast to all WebSocket connections
	broadcastUpdate("item_marked", req.Item)

	return c.JSON(http.StatusOK, map[string]string{"status": "marked"})
}

// Unmark global item
func unmarkGlobalItem(c echo.Context) error {
	var req struct {
		Item string `json:"item"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	// Find and remove item
	for i, markedItem := range db.GlobalMarkedItems {
		if markedItem == req.Item {
			db.GlobalMarkedItems = append(db.GlobalMarkedItems[:i], db.GlobalMarkedItems[i+1:]...)

			// Unmark this item on all player cards that contain it
			for j := range db.BingoCards {
				card := &db.BingoCards[j]
				for row := 0; row < 5; row++ {
					for col := 0; col < 5; col++ {
						if card.Items[row][col] == req.Item {
							card.MarkedItems[row][col] = false
							// Re-check if this card is still a winner
							card.IsWinner = checkBingo(card.MarkedItems)
						}
					}
				}
			}

			saveDatabase()

			// Broadcast to all WebSocket connections
			broadcastUpdate("item_unmarked", req.Item)

			return c.JSON(http.StatusOK, map[string]string{"status": "unmarked"})
		}
	}

	return c.JSON(http.StatusBadRequest, map[string]string{"error": "Item not found in marked items"})
}

// Get all bingo cards
func getAllBingoCards(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"cards": db.BingoCards,
		"users": db.Users,
	})
}

// WebSocket handler
func handleWebSocket(c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer ws.Close()

	connMutex.Lock()
	connections[ws] = true
	connMutex.Unlock()

	defer func() {
		connMutex.Lock()
		delete(connections, ws)
		connMutex.Unlock()
	}()

	// Send current state
	ws.WriteJSON(map[string]interface{}{
		"type":         "initial_state",
		"marked_items": db.GlobalMarkedItems,
	})

	// Keep connection alive
	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			break
		}
	}

	return nil
}

// Broadcast update to all WebSocket connections
func broadcastUpdate(eventType, item string) {
	connMutex.RLock()
	defer connMutex.RUnlock()

	message := map[string]interface{}{
		"type":  eventType,
		"item":  item,
		"cards": db.BingoCards, // Send updated cards
	}

	for conn := range connections {
		if err := conn.WriteJSON(message); err != nil {
			delete(connections, conn)
			conn.Close()
		}
	}
}
