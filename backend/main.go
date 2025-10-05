package main

import (
	"cmp"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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
	e.POST("/auth/discord/exchange", authCodeExchangeHandler)

	apiRoutes := e.Group("/api")
	apiRoutes.GET("/user", getCurrentUser, authMiddleware)
	apiRoutes.GET("/users", getAllUsersHandler, authMiddleware)
	apiRoutes.GET("/themes", getThemesHandler, authMiddleware)
	apiRoutes.GET("/themes/:id/items", getThemeItemsHandler, authMiddleware)
	apiRoutes.GET("/themes/:id/cards/mine", getCardByUserIdHandler, authMiddleware)

	// Admin routes
	adminRoutes := apiRoutes.Group("/admin", authMiddleware, adminMiddleware)

	adminRoutes.GET("/check", checkAdminAccessHandler)

	// admin theme management
	adminRoutes.POST("/themes/:themeId/items/:itemId/toggle", toggleItemHandler)
	adminRoutes.POST("/themes", createThemeHandler)
	adminRoutes.PUT("/themes/:id", updateThemeHandler)
	adminRoutes.DELETE("/themes/:id", deleteThemeHandler)
	adminRoutes.GET("/themes/:id/cards", getAllCardsHandler)
	adminRoutes.POST("/themes/:id/complete", setThemeCompleteHandler)
	adminRoutes.POST("/themes/active", setActiveThemeHandler)

	// WebSocket endpoint
	e.GET("/ws", webSocketHandler)

	e.GET("/api/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	port := cmp.Or(os.Getenv("PORT"), "8080")
	log.Printf("Server starting on port %s", port)
	e.Logger.Fatal(e.Start(":" + port))
}
