package main

import "time"

type BingoCard struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	Items       [][]string `json:"items"`
	MarkedItems [][]bool   `json:"marked_items"`
	CreatedAt   time.Time  `json:"created_at"`
	IsWinner    bool       `json:"is_winner"`
}
