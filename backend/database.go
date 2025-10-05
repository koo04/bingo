package main

import (
	"encoding/json"
	"log"
	"os"
	"slices"
	"strings"
)

type Database struct {
	Users           []*User  `json:"users"`
	BingoCards      []*Card  `json:"bingo_cards"`
	AdminDiscordIDs []string `json:"admin_discord_ids"`
	Themes          []*Theme `json:"themes"`
	ActiveThemeID   string   `json:"active_theme_id"`
}

func loadDatabase() error {
	data, err := os.ReadFile("data/database.json")
	if err != nil {
		log.Println("Database file not found, creating new one")
		db = Database{
			Users:           []*User{},
			BingoCards:      []*Card{},
			AdminDiscordIDs: []string{},
			Themes:          []*Theme{},
			ActiveThemeID:   "",
		}
		return saveDatabase()
	}

	if err := json.Unmarshal(data, &db); err != nil {
		log.Fatal("Error parsing database:", err)
	}

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
