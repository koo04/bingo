package main

import "time"

type BingoCard struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	ThemeID   string     `json:"theme_id"`
	Items     [][]string `json:"items"`
	CreatedAt time.Time  `json:"created_at"`
	IsWinner  bool       `json:"is_winner"`
}

func (b *BingoCard) checkBingo(theme *Theme) {
	const gridSize = 5

	// Check rows
	if b.hasWinningRow(gridSize, theme) {
		b.IsWinner = true
		return
	}

	// Check columns
	if b.hasWinningColumn(gridSize, theme) {
		b.IsWinner = true
		return
	}

	// Check diagonals
	if b.hasWinningDiagonal(gridSize, theme) {
		b.IsWinner = true
		return
	}

	b.IsWinner = false
}

// hasWinningRow checks if any row has all items marked
func (b *BingoCard) hasWinningRow(gridSize int, theme *Theme) bool {
	for row := range gridSize {
		if b.isRowComplete(row, gridSize, theme) {
			return true
		}
	}
	return false
}

// hasWinningColumn checks if any column has all items marked
func (b *BingoCard) hasWinningColumn(gridSize int, theme *Theme) bool {
	for col := range gridSize {
		if b.isColumnComplete(col, gridSize, theme) {
			return true
		}
	}
	return false
}

// hasWinningDiagonal checks if either diagonal has all items marked
func (b *BingoCard) hasWinningDiagonal(gridSize int, theme *Theme) bool {
	return b.isMainDiagonalComplete(gridSize, theme) || b.isAntiDiagonalComplete(gridSize, theme)
}

// isItemMarked checks if an item with the given ID is marked in the theme
func (b *BingoCard) isItemMarked(itemID string, theme *Theme) bool {
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
func (b *BingoCard) isRowComplete(row, gridSize int, theme *Theme) bool {
	for col := range gridSize {
		if !b.isItemMarked(b.Items[row][col], theme) {
			return false
		}
	}
	return true
}

// isColumnComplete checks if a specific column has all items marked
func (b *BingoCard) isColumnComplete(col, gridSize int, theme *Theme) bool {
	for row := range gridSize {
		if !b.isItemMarked(b.Items[row][col], theme) {
			return false
		}
	}
	return true
}

// isMainDiagonalComplete checks if the main diagonal (top-left to bottom-right) has all items marked
func (b *BingoCard) isMainDiagonalComplete(gridSize int, theme *Theme) bool {
	for i := range gridSize {
		if !b.isItemMarked(b.Items[i][i], theme) {
			return false
		}
	}
	return true
}

// isAntiDiagonalComplete checks if the anti-diagonal (top-right to bottom-left) has all items marked
func (b *BingoCard) isAntiDiagonalComplete(gridSize int, theme *Theme) bool {
	for i := range gridSize {
		if !b.isItemMarked(b.Items[i][gridSize-1-i], theme) {
			return false
		}
	}
	return true
}
