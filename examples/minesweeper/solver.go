// CSP solver for no-guess board generation and hints.
package main

import (
	"math/rand/v2"
	"slices"
)

type SolverResult struct {
	Safe  []Point
	Mines []Point
}

type constraint struct {
	vars  []int // variable indices (into frontier list)
	count int   // remaining mines among these vars
}

// Solve analyzes the current board and returns cells that can be
// deterministically identified as safe or mines.
func Solve(g *Game) SolverResult {
	// Try single-constraint deductions first.
	result := easyMoves(g)
	if len(result.Safe) > 0 || len(result.Mines) > 0 {
		return result
	}

	// Build frontier: hidden cells adjacent to revealed numbers.
	frontierSet := make(map[int]bool)
	var constraints []constraint
	for r := range g.Rows {
		for c := range g.Cols {
			cell := g.cell(r, c)
			if cell.State != CellRevealed || cell.Adjacent == 0 {
				continue
			}
			var vars []int
			flagCount := 0
			for _, n := range g.neighbors(r, c) {
				nc := g.cell(n.Row, n.Col)
				switch nc.State {
				case CellFlagged:
					flagCount++
				case CellHidden:
					idx := g.idx(n.Row, n.Col)
					frontierSet[idx] = true
					vars = append(vars, idx)
				}
			}
			remaining := cell.Adjacent - flagCount
			if len(vars) > 0 && remaining >= 0 {
				constraints = append(constraints,
					constraint{vars: vars, count: remaining})
			}
		}
	}
	if len(frontierSet) == 0 {
		return SolverResult{}
	}

	// Map board indices → sequential variable indices.
	frontier := make([]int, 0, len(frontierSet))
	for idx := range frontierSet {
		frontier = append(frontier, idx)
	}
	slices.Sort(frontier)
	varIdx := make(map[int]int, len(frontier))
	for i, bi := range frontier {
		varIdx[bi] = i
	}
	for i := range constraints {
		for j := range constraints[i].vars {
			constraints[i].vars[j] = varIdx[constraints[i].vars[j]]
		}
	}

	// Divide into independent areas and solve each.
	areas := divideIntoAreas(len(frontier), constraints)
	var merged SolverResult
	for _, a := range areas {
		partial := solveArea(a)
		for _, vi := range partial.safeVars {
			bi := frontier[vi]
			merged.Safe = append(merged.Safe,
				Point{bi / g.Cols, bi % g.Cols})
		}
		for _, vi := range partial.mineVars {
			bi := frontier[vi]
			merged.Mines = append(merged.Mines,
				Point{bi / g.Cols, bi % g.Cols})
		}
	}
	return merged
}

// easyMoves finds trivial deductions from individual constraints.
func easyMoves(g *Game) SolverResult {
	var result SolverResult
	for r := range g.Rows {
		for c := range g.Cols {
			cell := g.cell(r, c)
			if cell.State != CellRevealed || cell.Adjacent == 0 {
				continue
			}
			var hidden []Point
			flagCount := 0
			for _, n := range g.neighbors(r, c) {
				nc := g.cell(n.Row, n.Col)
				switch nc.State {
				case CellFlagged:
					flagCount++
				case CellHidden:
					hidden = append(hidden, n)
				}
			}
			if len(hidden) == 0 {
				continue
			}
			remaining := cell.Adjacent - flagCount
			if remaining == 0 {
				result.Safe = append(result.Safe, hidden...)
			} else if remaining == len(hidden) {
				result.Mines = append(result.Mines, hidden...)
			}
		}
	}
	result.Safe = dedupPoints(result.Safe)
	result.Mines = dedupPoints(result.Mines)
	return result
}

func dedupPoints(pts []Point) []Point {
	if len(pts) == 0 {
		return nil
	}
	seen := make(map[Point]bool, len(pts))
	out := make([]Point, 0, len(pts))
	for _, p := range pts {
		if !seen[p] {
			seen[p] = true
			out = append(out, p)
		}
	}
	return out
}

// --- Independent area decomposition via union-find ---

type area struct {
	vars        []int        // global variable indices
	constraints []constraint // constraints remapped to local indices
}

type areaResult struct {
	safeVars []int // global variable indices that must be 0
	mineVars []int // global variable indices that must be 1
}

func divideIntoAreas(numVars int, constraints []constraint) []area {
	parent := make([]int, numVars)
	for i := range parent {
		parent[i] = i
	}
	var find func(int) int
	find = func(x int) int {
		if parent[x] != x {
			parent[x] = find(parent[x])
		}
		return parent[x]
	}
	union := func(a, b int) {
		ra, rb := find(a), find(b)
		if ra != rb {
			parent[ra] = rb
		}
	}
	for _, c := range constraints {
		for i := 1; i < len(c.vars); i++ {
			union(c.vars[0], c.vars[i])
		}
	}

	groups := make(map[int][]int)
	for v := range numVars {
		groups[find(v)] = append(groups[find(v)], v)
	}

	var areas []area
	for _, vars := range groups {
		varSet := make(map[int]bool, len(vars))
		localIdx := make(map[int]int, len(vars))
		for i, v := range vars {
			varSet[v] = true
			localIdx[v] = i
		}
		var ac []constraint
		for _, c := range constraints {
			if !varSet[c.vars[0]] {
				continue
			}
			lv := make([]int, len(c.vars))
			for j, v := range c.vars {
				lv[j] = localIdx[v]
			}
			ac = append(ac, constraint{vars: lv, count: c.count})
		}
		if len(ac) > 0 {
			areas = append(areas, area{vars: vars, constraints: ac})
		}
	}
	return areas
}

// solveArea uses backtracking CSP to find which variables are
// determined (must be mine or must be safe in all solutions).
func solveArea(a area) areaResult {
	n := len(a.vars)
	if n > 25 {
		return areaResult{} // too large for exhaustive search
	}

	assignment := make([]int8, n) // -1=unset, 0=safe, 1=mine
	for i := range assignment {
		assignment[i] = -1
	}
	mineCounts := make([]int, n)
	totalSolutions := 0

	consistent := func() bool {
		for _, c := range a.constraints {
			mines, unset := 0, 0
			for _, v := range c.vars {
				switch assignment[v] {
				case 1:
					mines++
				case -1:
					unset++
				}
			}
			if mines > c.count || mines+unset < c.count {
				return false
			}
		}
		return true
	}

	var solve func(int)
	solve = func(idx int) {
		if idx == n {
			for _, c := range a.constraints {
				mines := 0
				for _, v := range c.vars {
					if assignment[v] == 1 {
						mines++
					}
				}
				if mines != c.count {
					return
				}
			}
			totalSolutions++
			for i := range n {
				if assignment[i] == 1 {
					mineCounts[i]++
				}
			}
			return
		}
		for _, val := range []int8{0, 1} {
			assignment[idx] = val
			if consistent() {
				solve(idx + 1)
			}
		}
		assignment[idx] = -1
	}
	solve(0)

	var result areaResult
	if totalSolutions == 0 {
		return result
	}
	for i := range n {
		switch mineCounts[i] {
		case 0:
			result.safeVars = append(result.safeVars, a.vars[i])
		case totalSolutions:
			result.mineVars = append(result.mineVars, a.vars[i])
		}
	}
	return result
}

// --- No-guess board generation ---

// IsSolvable checks if a board can be fully solved from the given
// start cell using only deterministic deductions.
func IsSolvable(board []Cell, rows, cols, startRow, startCol int) bool {
	mineCount := countMines(board)
	sim := &Game{
		Rows:      rows,
		Cols:      cols,
		MineCount: mineCount,
		Board:     slices.Clone(board),
		State:     GamePlaying,
	}
	sim.floodFill(startRow, startCol)

	target := rows*cols - mineCount
	for sim.Revealed < target {
		result := Solve(sim)
		if len(result.Safe) == 0 && len(result.Mines) == 0 {
			return false
		}
		for _, p := range result.Safe {
			if sim.cell(p.Row, p.Col).State == CellHidden {
				sim.floodFill(p.Row, p.Col)
			}
		}
		for _, p := range result.Mines {
			if sim.cell(p.Row, p.Col).State == CellHidden {
				sim.cell(p.Row, p.Col).State = CellFlagged
				sim.FlagsUsed++
			}
		}
	}
	return true
}

func countMines(board []Cell) int {
	n := 0
	for i := range board {
		if board[i].Mine {
			n++
		}
	}
	return n
}

// GenerateNoGuess generates a board that is fully solvable without
// guessing. Returns nil if generation fails after many attempts.
func GenerateNoGuess(rows, cols, mines, startRow, startCol int, rng RandSource) []Cell {
	if rng == nil {
		rng = rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	}
	maxAttempts := 1000
	if rows*cols > 200 {
		maxAttempts = 200
	}
	for range maxAttempts {
		board := randomBoard(rows, cols, mines, startRow, startCol, rng)
		if IsSolvable(board, rows, cols, startRow, startCol) {
			return board
		}
	}
	return nil
}

func randomBoard(rows, cols, mines, safeRow, safeCol int, rng RandSource) []Cell {
	g := &Game{Rows: rows, Cols: cols, MineCount: mines, rng: rng}
	g.Board = make([]Cell, rows*cols)
	g.placeMinesRandom(safeRow, safeCol)
	return g.Board
}

// FindHint returns a safe cell to reveal. It first tries logical
// deduction via the solver; if that yields nothing it falls back to
// ground truth, picking a safe hidden cell on the frontier (adjacent
// to a revealed cell) with the most revealed neighbors.
func FindHint(g *Game) *Point {
	result := Solve(g)
	// Filter to genuinely safe cells (guards against wrong flags).
	safe := make([]Point, 0, len(result.Safe))
	for _, p := range result.Safe {
		if !g.cell(p.Row, p.Col).Mine {
			safe = append(safe, p)
		}
	}
	// Fallback: any safe frontier cell using ground truth.
	if len(safe) == 0 {
		safe = frontierCells(g, false)
	}
	if len(safe) == 0 {
		return nil
	}
	return pickEasiest(g, safe)
}

// frontierCells returns hidden cells adjacent to at least one
// revealed cell, filtered by mine status.
func frontierCells(g *Game, wantMine bool) []Point {
	var out []Point
	for r := range g.Rows {
		for c := range g.Cols {
			cell := g.cell(r, c)
			if cell.State != CellHidden || cell.Mine != wantMine {
				continue
			}
			for _, n := range g.neighbors(r, c) {
				if g.cell(n.Row, n.Col).State == CellRevealed {
					out = append(out, Point{r, c})
					break
				}
			}
		}
	}
	return out
}

// pickEasiest returns the cell with the most revealed neighbors.
func pickEasiest(g *Game, pts []Point) *Point {
	best := pts[0]
	bestScore := -1
	for _, p := range pts {
		score := 0
		for dr := -1; dr <= 1; dr++ {
			for dc := -1; dc <= 1; dc++ {
				if dr == 0 && dc == 0 {
					continue
				}
				nr, nc := p.Row+dr, p.Col+dc
				if g.inBounds(nr, nc) &&
					g.cell(nr, nc).State == CellRevealed {
					score++
				}
			}
		}
		if score > bestScore {
			bestScore = score
			best = p
		}
	}
	return &best
}
