package main

import (
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/google/uuid"
)

type Theme struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Items       []*Item          `json:"items"`
	IsComplete  bool             `json:"is_complete"`
	Cards       map[string]*Card `json:"cards"`
	CreatedAt   time.Time        `json:"created_at"`
}

// Get theme by ID
func getThemeByID(themeID string) (*Theme, bool) {
	for _, theme := range db.Themes {
		if theme.ID == themeID {
			return theme, true
		}
	}
	return nil, false
}

func (t *Theme) GetItem(itemID string) (*Item, bool) {
	for _, item := range t.Items {
		if item.ID == itemID {
			return item, true
		}
	}
	return nil, false
}

func (t *Theme) NewBingoCard(user *User) (*Card, error) {
	if t.IsComplete {
		return nil, fmt.Errorf("cannot generate card from a completed theme")
	}

	if len(t.Items) < 25 {
		return nil, fmt.Errorf("theme has insufficient items: need at least 25, have %d", len(t.Items))
	}

	// Create a new bingo card
	card := &Card{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		ThemeID:   t.ID,
		Items:     make([][]string, 5),
		CreatedAt: time.Now(),
		IsWinner:  false,
	}

	// Shuffle and select 25 items
	shuffled := make([]*Item, len(t.Items))
	copy(shuffled, t.Items)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})
	selected := shuffled[:25]

	// Fill the 5x5 grid
	for i := range card.Items {
		card.Items[i] = make([]string, 5)
		for j := range card.Items[i] {
			card.Items[i][j] = selected[i*5+j].ID
		}
	}

	// Free space in the middle
	card.Items[2][2] = "FREE_SPACE"

	if t.Cards == nil {
		t.Cards = make(map[string]*Card)
	}
	t.Cards[user.ID] = card

	return card, saveDatabase()
}

func (t *Theme) checkForWinners() []*Card {
	var winners []*Card
	// Check all cards for winners
	for _, card := range t.Cards {
		card.checkBingo(t)
		if card.IsWinner {
			winners = append(winners, card)
		}
	}
	return winners
}

// Set active theme (admin only)

// Create new theme (admin only)

// Update theme (admin only)

// Delete theme (admin only)

// Mark theme as complete (admin only)
