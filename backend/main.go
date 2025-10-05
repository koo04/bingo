package main

import (
	"cmp"
	"encoding/json"
	"fmt"
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

func main() {
	if err := loadDatabase(); err != nil {
		log.Fatal("Error loading database:", err)
	}

	fmt.Printf("discord client id: %s\n", discordOAuth.ClientID)

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
	e.POST("/auth/discord/exchange", handleAuthCodeExchange)

	apiRoutes := e.Group("/api")
	apiRoutes.GET("/user", getCurrentUser, authMiddleware)
	apiRoutes.GET("/users", getAllUsers, authMiddleware)
	apiRoutes.GET("/themes", getThemes, authMiddleware)
	apiRoutes.GET("/themes/:id/items", getThemeItemsRequest, authMiddleware)
	apiRoutes.GET("/themes/:id/cards/mine", getMyBingoCard, authMiddleware)

	// Admin routes
	adminRoutes := apiRoutes.Group("/admin", authMiddleware, adminMiddleware)

	adminRoutes.GET("/check", checkAdminAccess)

	// admin theme management
	adminRoutes.POST("/themes/:themeId/items/:itemId/toggle", toggleItem)
	adminRoutes.POST("/themes", createTheme)
	adminRoutes.PUT("/themes/:id", updateTheme)
	adminRoutes.DELETE("/themes/:id", deleteTheme)
	adminRoutes.GET("/themes/:id/cards", getAllBingoCards)
	adminRoutes.POST("/themes/:id/complete", markThemeComplete)
	adminRoutes.POST("/themes/active", setActiveTheme)

	// WebSocket endpoint
	e.GET("/ws", handleWebSocket)

	e.GET("/api/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	port := cmp.Or(os.Getenv("PORT"), "8080")
	log.Printf("Server starting on port %s", port)
	e.Logger.Fatal(e.Start(":" + port))
}

func getAllUsers(c echo.Context) error {
	return c.JSON(http.StatusOK, db.Users)
}

func getMyBingoCard(c echo.Context) error {
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

	card, err := theme.NewBingoCard(user)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	if err := saveDatabase(); err != nil {
		c.Logger().Error("Error saving database:", err)
	}

	return c.JSON(http.StatusOK, card)
}

func loadDatabase() error {
	data, err := os.ReadFile("data/database.json")
	if err != nil {
		log.Println("Database file not found, creating new one")
		db = Database{
			Users:           []*User{},
			BingoCards:      []*BingoCard{},
			AdminDiscordIDs: []string{}, // Add admin Discord IDs here manually or via environment
			Themes:          []*Theme{},
			ActiveThemeID:   "",
		}
		return saveDatabase()
	}

	if err := json.Unmarshal(data, &db); err != nil {
		log.Fatal("Error parsing database:", err)
	}

	// No longer need to link card items to pointers since we store IDs directly

	// Initialize fields if they don't exist
	if db.AdminDiscordIDs == nil {
		db.AdminDiscordIDs = []string{}
	}

	// Add admin IDs from environment variable
	envAdminIDs := os.Getenv("ADMIN_DISCORD_IDS")
	if envAdminIDs == "" {
		return nil
	}

	envIDs := strings.SplitSeq(envAdminIDs, ",")
	for id := range envIDs {
		id = strings.TrimSpace(id)
		if id != "" {
			// Check if already exists
			exists := slices.Contains(db.AdminDiscordIDs, id)
			if !exists {
				db.AdminDiscordIDs = append(db.AdminDiscordIDs, id)
			}
		}
	}

	return saveDatabase()
}

func saveDatabase() error {
	if err := os.MkdirAll("data", 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile("data/database.json", data, 0644)
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

// Toggle item
func toggleItem(c echo.Context) error {
	var req struct {
		ThemeID string `json:"theme_id" param:"themeId"`
		ItemId  string `json:"item_id" param:"itemId"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	theme, found := getThemeByID(req.ThemeID)
	if !found {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Theme not found"})
	}

	item, found := theme.GetItem(req.ItemId)
	if !found {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Item not found"})
	}

	item.Marked = !item.Marked

	winners := checkForWinners(theme)
	if len(winners) > 0 {
		// Broadcast winner
		broadcastUpdate("winners", map[string]any{
			"cards": winners,
		})
	}

	if err := saveDatabase(); err != nil {
		c.Logger().Error("Error saving database:", err)
	}

	// Broadcast to all WebSocket connections
	broadcastUpdate("item_updated", item)

	return c.JSON(http.StatusOK, map[string]string{"status": "marked"})
}

func checkForWinners(theme *Theme) []*BingoCard {
	var winners []*BingoCard
	// Check all cards for winners
	for _, card := range theme.Cards {
		card.checkBingo(theme)
		if card.IsWinner {
			winners = append(winners, card)
		}
	}
	return winners
}

// Get all bingo cards
func getAllBingoCards(c echo.Context) error {
	type request struct {
		ThemeID string `json:"theme_id" param:"id"`
	}

	var req request
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	if req.ThemeID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Theme ID is required"})
	}

	theme, found := getThemeByID(req.ThemeID)
	if !found {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Theme not found"})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"cards": theme.Cards,
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
func broadcastUpdate(eventType string, item any) {
	connMutex.RLock()
	defer connMutex.RUnlock()

	message := map[string]any{
		"type": eventType,
		"data": item,
	}

	for conn := range connections {
		if err := conn.WriteJSON(message); err != nil {
			delete(connections, conn)
			conn.Close()
		}
	}
}
