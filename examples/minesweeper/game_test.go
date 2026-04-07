package main

import (
	"math/rand/v2"
	"testing"
)

// manualBoard creates a game with pre-placed mines (skips first-click logic).
func manualBoard(rows, cols int, mines ...Point) *Game {
	g := &Game{
		Rows:       rows,
		Cols:       cols,
		MineCount:  len(mines),
		Board:      make([]Cell, rows*cols),
		State:      GamePlaying,
		FirstClick: false,
	}
	for _, m := range mines {
		g.Board[g.idx(m.Row, m.Col)].Mine = true
	}
	g.computeAdjacent()
	return g
}

func TestSafeFirstClick(t *testing.T) {
	rng := rand.New(rand.NewPCG(42, 0))
	g := NewGame(9, 9, 10, false, rng)
	g.Reveal(4, 4)
	for dr := -1; dr <= 1; dr++ {
		for dc := -1; dc <= 1; dc++ {
			if g.cell(4+dr, 4+dc).Mine {
				t.Errorf("mine at safe zone (%d,%d)", 4+dr, 4+dc)
			}
		}
	}
	if g.cell(4, 4).State != CellRevealed {
		t.Error("center should be revealed")
	}
}

func TestFloodFill(t *testing.T) {
	// 3x3 grid with mine at (2,2). Flood from (0,0) should
	// reveal all 8 non-mine cells.
	g := manualBoard(3, 3, Point{2, 2})
	g.floodFill(0, 0)
	if g.Revealed != 8 {
		t.Errorf("expected 8 revealed, got %d", g.Revealed)
	}
	if g.cell(2, 2).State != CellHidden {
		t.Error("mine cell should remain hidden")
	}
}

func TestFlagging(t *testing.T) {
	g := manualBoard(3, 3, Point{0, 0})
	g.Flag(0, 0)
	if g.cell(0, 0).State != CellFlagged || g.FlagsUsed != 1 {
		t.Error("cell should be flagged")
	}
	g.Flag(0, 0) // unflag
	if g.cell(0, 0).State != CellHidden || g.FlagsUsed != 0 {
		t.Error("cell should be unflagged")
	}
	// Can't flag a revealed cell.
	g.cell(1, 1).State = CellRevealed
	g.Flag(1, 1)
	if g.cell(1, 1).State != CellRevealed {
		t.Error("should not flag a revealed cell")
	}
}

func TestChording(t *testing.T) {
	g := manualBoard(3, 3, Point{0, 0})
	// Reveal center (1,1) which has Adjacent=1.
	g.cell(1, 1).State = CellRevealed
	g.Revealed = 1
	// Flag the mine.
	g.Flag(0, 0)

	ok := g.Chord(1, 1)
	if !ok {
		t.Error("chord should succeed")
	}
	for _, n := range g.neighbors(1, 1) {
		if n == (Point{0, 0}) {
			continue
		}
		if g.cell(n.Row, n.Col).State != CellRevealed {
			t.Errorf("(%d,%d) should be revealed after chord",
				n.Row, n.Col)
		}
	}
}

func TestChordMineHit(t *testing.T) {
	g := manualBoard(3, 3, Point{0, 0})
	g.cell(1, 1).State = CellRevealed
	g.Revealed = 1
	// Flag the wrong cell.
	g.Flag(0, 1)

	ok := g.Chord(1, 1)
	if ok {
		t.Error("chord should hit mine and return false")
	}
	if g.State != GameLost {
		t.Error("game should be lost")
	}
}

func TestWinDetection(t *testing.T) {
	g := manualBoard(3, 3, Point{2, 2})
	g.floodFill(0, 0)
	g.checkWin()
	if g.State != GameWon {
		t.Errorf("expected GameWon, got %d", g.State)
	}
}

func TestCheckBoard(t *testing.T) {
	g := manualBoard(3, 3, Point{0, 0})
	g.cell(1, 1).State = CellRevealed
	g.Revealed = 1
	// Two flags around a cell with Adjacent=1 → over-flagged.
	g.Flag(0, 0)
	g.Flag(0, 1)
	bad := g.CheckBoard()
	if len(bad) != 1 || bad[0] != (Point{1, 1}) {
		t.Errorf("expected bad check at (1,1), got %v", bad)
	}
}

func TestRevealMine(t *testing.T) {
	g := manualBoard(3, 3, Point{1, 1})
	ok := g.Reveal(1, 1)
	if ok {
		t.Error("revealing mine should return false")
	}
	if g.State != GameLost {
		t.Error("game should be lost")
	}
}

func TestAutoFlagMines(t *testing.T) {
	g := manualBoard(3, 3, Point{0, 0})
	g.AutoFlagMines()
	if g.cell(0, 0).State != CellFlagged || g.FlagsUsed != 1 {
		t.Error("mine should be auto-flagged")
	}
}

// --- Solver tests ---

func TestEasyMovesSafe(t *testing.T) {
	// 3x3 with mine at (0,0). Reveal everything except (0,0).
	g := manualBoard(3, 3, Point{0, 0})
	for i := 1; i < 9; i++ {
		g.Board[i].State = CellRevealed
		g.Revealed++
	}
	// Flag the mine.
	g.cell(0, 0).State = CellFlagged
	g.FlagsUsed = 1

	// Now if we place another mine situation: let's verify safe moves
	// work. We'll test with a fresh board instead.
	g2 := manualBoard(3, 3, Point{0, 0})
	for i := 1; i < 9; i++ {
		g2.Board[i].State = CellRevealed
		g2.Revealed++
	}
	result := easyMoves(g2)
	if len(result.Mines) != 1 || result.Mines[0] != (Point{0, 0}) {
		t.Errorf("expected mine at (0,0), got mines=%v safe=%v",
			result.Mines, result.Safe)
	}
}

func TestEasyMovesAllSafe(t *testing.T) {
	// 3x3 with mine at (2,2). Flag the mine. Reveal (1,1).
	g := manualBoard(3, 3, Point{2, 2})
	g.cell(2, 2).State = CellFlagged
	g.FlagsUsed = 1
	g.cell(1, 1).State = CellRevealed
	g.Revealed = 1

	result := easyMoves(g)
	// (1,1) has Adjacent=1, 1 flag → remaining=0 → all hidden safe.
	if len(result.Safe) == 0 {
		t.Error("expected safe cells from easy moves")
	}
}

func TestIsSolvable(t *testing.T) {
	// 3x3 with mine at (0,0), start at (2,2).
	board := make([]Cell, 9)
	board[0].Mine = true
	board[1].Adjacent = 1
	board[3].Adjacent = 1
	board[4].Adjacent = 1
	if !IsSolvable(board, 3, 3, 2, 2) {
		t.Error("3x3 with 1 mine should be solvable from (2,2)")
	}
}

func TestSolveCSP(t *testing.T) {
	// 3x3 board, mine at (0,1). Reveal rows 1-2 so that easy
	// moves fail but CSP can determine all three hidden cells.
	//   . M .    (hidden row)
	//   1 1 1    (revealed)
	//   0 0 0    (revealed)
	g := manualBoard(3, 3, Point{0, 1})
	for r := 1; r < 3; r++ {
		for c := range 3 {
			g.cell(r, c).State = CellRevealed
			g.Revealed++
		}
	}
	result := Solve(g)
	mineSet := make(map[Point]bool)
	for _, p := range result.Mines {
		mineSet[p] = true
	}
	safeSet := make(map[Point]bool)
	for _, p := range result.Safe {
		safeSet[p] = true
	}
	if !mineSet[Point{0, 1}] {
		t.Errorf("expected mine at (0,1), mines=%v safe=%v",
			result.Mines, result.Safe)
	}
	if !safeSet[Point{0, 0}] || !safeSet[Point{0, 2}] {
		t.Errorf("expected safe at (0,0) and (0,2), mines=%v safe=%v",
			result.Mines, result.Safe)
	}
}

func TestFindHint(t *testing.T) {
	// 3x3 board, mine at (0,1). Reveal rows 1-2.
	g := manualBoard(3, 3, Point{0, 1})
	for r := 1; r < 3; r++ {
		for c := range 3 {
			g.cell(r, c).State = CellRevealed
			g.Revealed++
		}
	}
	hint := FindHint(g)
	if hint == nil {
		t.Fatal("expected a hint")
	}
	// Should suggest (0,0) or (0,2) as safe.
	if (hint.Row != 0 || hint.Col != 0) &&
		(hint.Row != 0 || hint.Col != 2) {
		t.Errorf("expected hint at (0,0) or (0,2), got (%d,%d)",
			hint.Row, hint.Col)
	}
}

func TestDifficultyConfig(t *testing.T) {
	tests := []struct {
		d                 Difficulty
		rows, cols, mines int
	}{
		{DiffBeginner, 9, 9, 10},
		{DiffIntermediate, 16, 16, 40},
		{DiffExpert, 16, 30, 99},
	}
	for _, tt := range tests {
		r, c, m := tt.d.Config()
		if r != tt.rows || c != tt.cols || m != tt.mines {
			t.Errorf("Difficulty %d: got (%d,%d,%d) want (%d,%d,%d)",
				tt.d, r, c, m, tt.rows, tt.cols, tt.mines)
		}
	}
}
