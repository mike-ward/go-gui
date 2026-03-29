package main

import (
	"math/rand/v2"
)

type Grid [4][4]int

type GameState int

const (
	StatePlaying GameState = iota
	StateWin
	StateGameOver
)

type Point struct {
	X, Y int
}

type Game struct {
	Grid       Grid
	Score      int
	State      GameState
	TargetWin  int  // Usually 2048
	IsModified bool // True if the last move changed the grid
}

func NewGame() *Game {
	g := &Game{
		TargetWin: 2048,
		State:     StatePlaying,
	}
	g.Spawn()
	g.Spawn()
	return g
}

func (g *Game) Spawn() {
	var empty []Point
	for y := range 4 {
		for x := range 4 {
			if g.Grid[y][x] == 0 {
				empty = append(empty, Point{x, y})
			}
		}
	}
	if len(empty) == 0 {
		return
	}
	p := empty[rand.IntN(len(empty))]
	val := 2
	if rand.Float32() < 0.1 {
		val = 4
	}
	g.Grid[p.Y][p.X] = val
}

type Direction int

const (
	DirUp Direction = iota
	DirDown
	DirLeft
	DirRight
)

func (g *Game) Move(dir Direction) bool {
	if g.State == StateGameOver {
		return false
	}

	modified := false
	scoreAdded := 0

	// We'll process each row/column based on direction
	for i := range 4 {
		var line [4]int
		// Extract line
		for j := range 4 {
			switch dir {
			case DirUp:
				line[j] = g.Grid[j][i]
			case DirDown:
				line[j] = g.Grid[3-j][i]
			case DirLeft:
				line[j] = g.Grid[i][j]
			case DirRight:
				line[j] = g.Grid[i][3-j]
			}
		}

		newLine, s, m := g.mergeLine(line)
		scoreAdded += s
		if m {
			modified = true
		}

		// Put line back
		for j := range 4 {
			switch dir {
			case DirUp:
				g.Grid[j][i] = newLine[j]
			case DirDown:
				g.Grid[3-j][i] = newLine[j]
			case DirLeft:
				g.Grid[i][j] = newLine[j]
			case DirRight:
				g.Grid[i][3-j] = newLine[j]
			}
		}
	}

	if modified {
		g.Score += scoreAdded
		g.Spawn()
		if g.checkWin() {
			g.State = StateWin
		} else if g.checkGameOver() {
			g.State = StateGameOver
		}
	}
	g.IsModified = modified
	return modified
}

func (g *Game) mergeLine(line [4]int) ([4]int, int, bool) {
	var next [4]int
	pos := 0
	score := 0
	modified := false

	// Slide all non-zero elements
	for i := range 4 {
		if line[i] != 0 {
			next[pos] = line[i]
			if pos != i {
				modified = true
			}
			pos++
		}
	}

	// Merge adjacent identical elements
	for i := range 3 {
		if next[i] != 0 && next[i] == next[i+1] {
			next[i] *= 2
			score += next[i]
			modified = true
			// Shift remaining
			for j := i + 1; j < 3; j++ {
				next[j] = next[j+1]
			}
			next[3] = 0
		}
	}

	return next, score, modified
}

func (g *Game) checkWin() bool {
	for y := range 4 {
		for x := range 4 {
			if g.Grid[y][x] >= g.TargetWin {
				return true
			}
		}
	}
	return false
}

func (g *Game) checkGameOver() bool {
	// If there's an empty cell, not over
	for y := range 4 {
		for x := range 4 {
			if g.Grid[y][x] == 0 {
				return false
			}
		}
	}
	// Check for adjacent identical tiles
	for y := range 4 {
		for x := range 4 {
			val := g.Grid[y][x]
			if (x < 3 && g.Grid[y][x+1] == val) || (y < 3 && g.Grid[y+1][x] == val) {
				return false
			}
		}
	}
	return true
}

func (g *Game) Reset() {
	g.Grid = Grid{}
	g.Score = 0
	g.State = StatePlaying
	g.Spawn()
	g.Spawn()
}
