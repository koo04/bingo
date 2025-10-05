package main

import "time"

type Card struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	ThemeID   string     `json:"theme_id"`
	Items     [][]string `json:"items"`
	CreatedAt time.Time  `json:"created_at"`
	IsWinner  bool       `json:"is_winner"`
}

func (c *Card) checkBingo(theme *Theme) {
	const gridSize = 5

	// Check rows
	if c.hasWinningRow(gridSize, theme) {
		c.IsWinner = true
		return
	}

	// Check columns
	if c.hasWinningColumn(gridSize, theme) {
		c.IsWinner = true
		return
	}

	// Check diagonals
	if c.hasWinningDiagonal(gridSize, theme) {
		c.IsWinner = true
		return
	}

	c.IsWinner = false
}

// hasWinningRow checks if any row has all items marked
func (c *Card) hasWinningRow(gridSize int, theme *Theme) bool {
	for row := range gridSize {
		if c.isRowComplete(row, gridSize, theme) {
			return true
		}
	}
	return false
}

// hasWinningColumn checks if any column has all items marked
func (c *Card) hasWinningColumn(gridSize int, theme *Theme) bool {
	for col := range gridSize {
		if c.isColumnComplete(col, gridSize, theme) {
			return true
		}
	}
	return false
}

// hasWinningDiagonal checks if either diagonal has all items marked
func (c *Card) hasWinningDiagonal(gridSize int, theme *Theme) bool {
	return c.isMainDiagonalComplete(gridSize, theme) || c.isAntiDiagonalComplete(gridSize, theme)
}

// isItemMarked checks if an item with the given ID is marked in the theme
func (c *Card) isItemMarked(itemID string, theme *Theme) bool {
	// Handle FREE_SPACE specially - it's always considered marked
	if itemID == "FREE_SPACE" {
		return true
	}

	for _, item := range theme.Items {
		if item.ID == itemID {
			return item.Marked
		}
	}
	return false
}

// isRowComplete checks if a specific row has all items marked
func (c *Card) isRowComplete(row, gridSize int, theme *Theme) bool {
	for col := range gridSize {
		if !c.isItemMarked(c.Items[row][col], theme) {
			return false
		}
	}
	return true
}

// isColumnComplete checks if a specific column has all items marked
func (c *Card) isColumnComplete(col, gridSize int, theme *Theme) bool {
	for row := range gridSize {
		if !c.isItemMarked(c.Items[row][col], theme) {
			return false
		}
	}
	return true
}

// isMainDiagonalComplete checks if the main diagonal (top-left to bottom-right) has all items marked
func (c *Card) isMainDiagonalComplete(gridSize int, theme *Theme) bool {
	for i := range gridSize {
		if !c.isItemMarked(c.Items[i][i], theme) {
			return false
		}
	}
	return true
}

// isAntiDiagonalComplete checks if the anti-diagonal (top-right to bottom-left) has all items marked
func (c *Card) isAntiDiagonalComplete(gridSize int, theme *Theme) bool {
	for i := range gridSize {
		if !c.isItemMarked(c.Items[i][gridSize-1-i], theme) {
			return false
		}
	}
	return true
}
