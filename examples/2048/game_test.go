package main

import (
	"testing"
)

func TestMergeLine(t *testing.T) {
	g := &Game{}

	tests := []struct {
		input  [4]int
		output [4]int
		score  int
		mod    bool
	}{
		{[4]int{0, 0, 0, 0}, [4]int{0, 0, 0, 0}, 0, false},
		{[4]int{2, 0, 0, 0}, [4]int{2, 0, 0, 0}, 0, false},
		{[4]int{0, 2, 0, 0}, [4]int{2, 0, 0, 0}, 0, true},
		{[4]int{2, 2, 0, 0}, [4]int{4, 0, 0, 0}, 4, true},
		{[4]int{2, 2, 2, 2}, [4]int{4, 4, 0, 0}, 8, true},
		{[4]int{4, 2, 2, 0}, [4]int{4, 4, 0, 0}, 4, true},
		{[4]int{2, 4, 8, 16}, [4]int{2, 4, 8, 16}, 0, false},
	}

	for _, tc := range tests {
		out, score, mod := g.mergeLine(tc.input)
		if out != tc.output || score != tc.score || mod != tc.mod {
			t.Errorf("mergeLine(%v) = %v, %d, %v; want %v, %d, %v",
				tc.input, out, score, mod, tc.output, tc.score, tc.mod)
		}
	}
}

func TestMove(t *testing.T) {
	g := &Game{}
	g.Grid = Grid{
		{2, 2, 0, 0},
		{0, 0, 0, 0},
		{0, 0, 0, 0},
		{0, 0, 0, 0},
	}
	g.State = StatePlaying

	moved := g.Move(DirRight)
	if !moved {
		t.Error("Expected move to be successful")
	}
	if g.Grid[0][3] != 4 {
		t.Errorf("Expected 4 at [0][3], got %v", g.Grid[0])
	}
}
