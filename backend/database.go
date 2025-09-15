package main

type Database struct {
	Users             []User      `json:"users"`
	BingoCards        []BingoCard `json:"bingo_cards"`
	AdminDiscordIDs   []string    `json:"admin_discord_ids"`
	GlobalMarkedItems []string    `json:"global_marked_items"`
}
