package main

import "testing"

type stubRNG struct {
	values []int
	idx    int
}

func (s *stubRNG) Intn(n int) int {
	if n <= 0 {
		return 0
	}
	if len(s.values) == 0 {
		return 0
	}
	v := s.values[s.idx%len(s.values)]
	s.idx++
	if v < 0 {
		v = -v
	}
	return v % n
}

func TestNewGameInitialState(t *testing.T) {
	g := NewGame(20, 20, &stubRNG{})
	if g.Score != 0 {
		t.Errorf("NewGame score = %d, want 0", g.Score)
	}
	if g.Direction != DirRight {
		t.Errorf("NewGame direction = %v, want %v", g.Direction, DirRight)
	}
	if len(g.Snake) != 3 {
		t.Errorf("len(NewGame().Snake) = %d, want 3", len(g.Snake))
	}
	if containsPoint(g.Snake, g.Food) {
		t.Errorf("NewGame food = %+v overlaps snake %+v", g.Food, g.Snake)
	}
}

func TestSetDirectionRejectsOpposite(t *testing.T) {
	g := NewGame(10, 10, &stubRNG{})
	if ok := g.SetDirection(DirLeft); ok {
		t.Errorf("SetDirection(%v) = %v, want %v", DirLeft, ok, false)
	}
	if g.queuedDir != DirRight {
		t.Errorf("queued direction after opposite turn = %v, want %v", g.queuedDir, DirRight)
	}
}

func TestTickMovesForward(t *testing.T) {
	g := NewGame(10, 10, &stubRNG{})
	start := g.Snake[0]
	g.Food = Point{X: 0, Y: 0}

	g.Tick()

	gotHead := g.Snake[0]
	wantHead := Point{X: start.X + 1, Y: start.Y}
	if gotHead != wantHead {
		t.Errorf("Tick() head = %+v, want %+v", gotHead, wantHead)
	}
	if len(g.Snake) != 3 {
		t.Errorf("len(Snake) after move = %d, want 3", len(g.Snake))
	}
}

func TestTickGrowsAndIncrementsScore(t *testing.T) {
	g := NewGame(8, 8, &stubRNG{values: []int{0}})
	head := g.Snake[0]
	g.Food = Point{X: head.X + 1, Y: head.Y}

	g.Tick()

	if g.Score != 1 {
		t.Errorf("score after eating = %d, want 1", g.Score)
	}
	if len(g.Snake) != 4 {
		t.Errorf("len(Snake) after growth = %d, want 4", len(g.Snake))
	}
	if containsPoint(g.Snake, g.Food) {
		t.Errorf("spawned food %+v overlaps snake %+v", g.Food, g.Snake)
	}
}

func TestTickDetectsWallCollision(t *testing.T) {
	g := &Game{
		Width:     4,
		Height:    4,
		Snake:     []Point{{X: 3, Y: 1}, {X: 2, Y: 1}, {X: 1, Y: 1}},
		Food:      Point{X: 0, Y: 0},
		Direction: DirRight,
		queuedDir: DirRight,
		rng:       &stubRNG{},
	}

	g.Tick()

	if !g.GameOver {
		t.Errorf("Tick() at wall GameOver = %v, want %v", g.GameOver, true)
	}
}

func TestTickDetectsSelfCollision(t *testing.T) {
	g := &Game{
		Width:  6,
		Height: 6,
		Snake: []Point{
			{X: 2, Y: 2},
			{X: 2, Y: 3},
			{X: 1, Y: 3},
			{X: 1, Y: 2},
			{X: 1, Y: 1},
			{X: 2, Y: 1},
		},
		Food:      Point{X: 5, Y: 5},
		Direction: DirDown,
		queuedDir: DirLeft,
		rng:       &stubRNG{},
	}

	g.Tick()

	if !g.GameOver {
		t.Errorf("Tick() self collision GameOver = %v, want %v", g.GameOver, true)
	}
}

func TestSpawnFoodDeterministicFromFreeCells(t *testing.T) {
	rng := &stubRNG{values: []int{2}}
	g := &Game{
		Width:  3,
		Height: 3,
		Snake:  []Point{{X: 0, Y: 0}, {X: 1, Y: 0}},
		rng:    rng,
	}

	g.spawnFood()

	free := []Point{
		{X: 2, Y: 0},
		{X: 0, Y: 1},
		{X: 1, Y: 1},
		{X: 2, Y: 1},
		{X: 0, Y: 2},
		{X: 1, Y: 2},
		{X: 2, Y: 2},
	}
	want := free[2]
	if g.Food != want {
		t.Errorf("spawnFood() picked %+v, want %+v", g.Food, want)
	}
}

func TestSpawnFoodWhenBoardFullSetsWon(t *testing.T) {
	g := &Game{
		Width:  2,
		Height: 2,
		Snake:  []Point{{X: 0, Y: 0}, {X: 0, Y: 1}, {X: 1, Y: 0}, {X: 1, Y: 1}},
		rng:    &stubRNG{},
	}

	g.spawnFood()

	if !g.Won {
		t.Errorf("spawnFood() Won = %v, want %v", g.Won, true)
	}
	if !g.GameOver {
		t.Errorf("spawnFood() GameOver = %v, want %v", g.GameOver, true)
	}
}
