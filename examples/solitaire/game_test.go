package main

import "testing"

// fixedRNG returns values from a fixed sequence, wrapping around.
type fixedRNG struct {
	vals []int
	idx  int
}

func (r *fixedRNG) IntN(n int) int {
	v := r.vals[r.idx%len(r.vals)] % n
	r.idx++
	return v
}

func TestNewDeck(t *testing.T) {
	deck := newDeck()
	if len(deck) != 52 {
		t.Fatalf("deck size = %d, want 52", len(deck))
	}
	seen := map[[2]int]bool{}
	for _, c := range deck {
		key := [2]int{int(c.Suit), c.Rank}
		if seen[key] {
			t.Fatalf("duplicate card: suit=%d rank=%d", c.Suit, c.Rank)
		}
		seen[key] = true
	}
}

func TestShuffle(t *testing.T) {
	d1 := newDeck()
	d2 := newDeck()
	rng := &fixedRNG{vals: []int{3, 7, 1, 5, 2, 9, 0, 4, 8, 6}}
	shuffle(d1, rng)

	rng2 := &fixedRNG{vals: []int{3, 7, 1, 5, 2, 9, 0, 4, 8, 6}}
	shuffle(d2, rng2)

	for i := range d1 {
		if d1[i] != d2[i] {
			t.Fatalf("deterministic shuffle mismatch at %d", i)
		}
	}
}

func TestDeal(t *testing.T) {
	g := NewGame(DrawOne, &fixedRNG{vals: []int{0}})
	totalCards := len(g.Stock)
	for col := range 7 {
		n := len(g.Tableau[col])
		if n != col+1 {
			t.Errorf("tableau[%d] has %d cards, want %d", col, n, col+1)
		}
		totalCards += n
		// Top card face-up.
		if !g.Tableau[col][col].FaceUp {
			t.Errorf("tableau[%d] top card not face-up", col)
		}
		// Others face-down.
		for row := range col {
			if g.Tableau[col][row].FaceUp {
				t.Errorf("tableau[%d][%d] should be face-down", col, row)
			}
		}
	}
	if totalCards != 52 {
		t.Errorf("total cards = %d, want 52", totalCards)
	}
	if len(g.Stock) != 24 {
		t.Errorf("stock = %d, want 24", len(g.Stock))
	}
}

func TestDrawOne(t *testing.T) {
	g := NewGame(DrawOne, &fixedRNG{vals: []int{0}})
	stockBefore := len(g.Stock)
	g.Draw()
	if len(g.Waste) != 1 {
		t.Fatalf("waste = %d, want 1", len(g.Waste))
	}
	if len(g.Stock) != stockBefore-1 {
		t.Fatalf("stock = %d, want %d", len(g.Stock), stockBefore-1)
	}
	if !g.Waste[0].FaceUp {
		t.Error("waste card should be face-up")
	}
}

func TestDrawThree(t *testing.T) {
	g := NewGame(DrawThree, &fixedRNG{vals: []int{0}})
	g.Draw()
	if len(g.Waste) != 3 {
		t.Fatalf("waste = %d, want 3", len(g.Waste))
	}
	if len(g.Stock) != 21 {
		t.Fatalf("stock = %d, want 21", len(g.Stock))
	}
}

func TestDrawRecycle(t *testing.T) {
	g := NewGame(DrawOne, &fixedRNG{vals: []int{0}})
	// Drain stock.
	for len(g.Stock) > 0 {
		g.Draw()
	}
	wasteCount := len(g.Waste)
	if wasteCount != 24 {
		t.Fatalf("waste after drain = %d, want 24", wasteCount)
	}
	// Recycle.
	g.Draw()
	if len(g.Stock) != 24 {
		t.Fatalf("stock after recycle = %d, want 24", len(g.Stock))
	}
	if len(g.Waste) != 0 {
		t.Fatalf("waste after recycle = %d, want 0", len(g.Waste))
	}
}

func TestDrawThreeRecycleScore(t *testing.T) {
	g := NewGame(DrawThree, &fixedRNG{vals: []int{0}})
	g.Score = 200
	for len(g.Stock) > 0 {
		g.Draw()
	}
	g.Draw() // recycle
	if g.Score != 100 {
		t.Errorf("score after recycle = %d, want 100", g.Score)
	}
}

func TestTableauMoveValid(t *testing.T) {
	g := &Game{DrawMode: DrawOne}
	// Place a black 5 on tableau column 0.
	g.Tableau[0] = []Card{{Suit: Spades, Rank: 5, FaceUp: true}}
	// Red 4 should fit.
	c := Card{Suit: Hearts, Rank: 4}
	if !g.CanPlaceOnTableau(c, 0) {
		t.Error("red 4 on black 5 should be valid")
	}
}

func TestTableauMoveInvalid(t *testing.T) {
	g := &Game{DrawMode: DrawOne}
	g.Tableau[0] = []Card{{Suit: Spades, Rank: 5, FaceUp: true}}
	// Same color.
	if g.CanPlaceOnTableau(Card{Suit: Clubs, Rank: 4}, 0) {
		t.Error("black 4 on black 5 should be invalid")
	}
	// Wrong rank.
	if g.CanPlaceOnTableau(Card{Suit: Hearts, Rank: 3}, 0) {
		t.Error("red 3 on black 5 should be invalid")
	}
}

func TestKingOnEmpty(t *testing.T) {
	g := &Game{DrawMode: DrawOne}
	if !g.CanPlaceOnTableau(Card{Suit: Spades, Rank: 13}, 0) {
		t.Error("king on empty should be valid")
	}
	if g.CanPlaceOnTableau(Card{Suit: Spades, Rank: 12}, 0) {
		t.Error("queen on empty should be invalid")
	}
}

func TestFoundationBuild(t *testing.T) {
	g := &Game{DrawMode: DrawOne}
	if !g.CanPlaceOnFoundation(Card{Suit: Hearts, Rank: 1}) {
		t.Error("ace on empty foundation should be valid")
	}
	g.Foundation[Hearts] = []Card{{Suit: Hearts, Rank: 1, FaceUp: true}}
	if !g.CanPlaceOnFoundation(Card{Suit: Hearts, Rank: 2}) {
		t.Error("2 on ace should be valid")
	}
}

func TestFoundationReject(t *testing.T) {
	g := &Game{DrawMode: DrawOne}
	// Non-ace on empty.
	if g.CanPlaceOnFoundation(Card{Suit: Hearts, Rank: 2}) {
		t.Error("2 on empty foundation should be invalid")
	}
	// Wrong suit.
	g.Foundation[Hearts] = []Card{{Suit: Hearts, Rank: 1, FaceUp: true}}
	if g.CanPlaceOnFoundation(Card{Suit: Spades, Rank: 2}) {
		t.Error("spade 2 on hearts foundation should be invalid")
	}
}

func TestMoveWasteToTableau(t *testing.T) {
	g := &Game{DrawMode: DrawOne}
	g.Tableau[0] = []Card{{Suit: Spades, Rank: 5, FaceUp: true}}
	g.Waste = []Card{{Suit: Hearts, Rank: 4, FaceUp: true}}
	if !g.MoveWasteToTableau(0) {
		t.Fatal("move should succeed")
	}
	if len(g.Waste) != 0 {
		t.Error("waste should be empty")
	}
	if len(g.Tableau[0]) != 2 {
		t.Error("tableau should have 2 cards")
	}
	if g.Score != 5 {
		t.Errorf("score = %d, want 5", g.Score)
	}
}

func TestMoveWasteToFoundation(t *testing.T) {
	g := &Game{DrawMode: DrawOne}
	g.Waste = []Card{{Suit: Hearts, Rank: 1, FaceUp: true}}
	if !g.MoveWasteToFoundation() {
		t.Fatal("move should succeed")
	}
	if len(g.Foundation[Hearts]) != 1 {
		t.Error("foundation should have 1 card")
	}
	if g.Score != 10 {
		t.Errorf("score = %d, want 10", g.Score)
	}
}

func TestStackMove(t *testing.T) {
	g := &Game{DrawMode: DrawOne}
	g.Tableau[0] = []Card{
		{Suit: Spades, Rank: 7, FaceUp: true},
		{Suit: Hearts, Rank: 6, FaceUp: true},
		{Suit: Clubs, Rank: 5, FaceUp: true},
	}
	g.Tableau[1] = []Card{{Suit: Hearts, Rank: 8, FaceUp: true}}
	// Move 7♠ + stack to col 1 on top of 8♥.
	if !g.MoveTableauToTableau(0, 0, 1) {
		t.Fatal("stack move should succeed")
	}
	if len(g.Tableau[0]) != 0 {
		t.Errorf("src col = %d cards, want 0", len(g.Tableau[0]))
	}
	if len(g.Tableau[1]) != 4 {
		t.Errorf("dst col = %d cards, want 4", len(g.Tableau[1]))
	}
}

func TestFlipAfterMove(t *testing.T) {
	g := &Game{DrawMode: DrawOne}
	g.Tableau[0] = []Card{
		{Suit: Clubs, Rank: 9, FaceUp: false},
		{Suit: Hearts, Rank: 6, FaceUp: true},
	}
	g.Tableau[1] = []Card{{Suit: Spades, Rank: 7, FaceUp: true}}
	g.MoveTableauToTableau(0, 1, 1)
	if !g.Tableau[0][0].FaceUp {
		t.Error("exposed card should be flipped face-up")
	}
}

func TestMoveTableauToFoundation(t *testing.T) {
	g := &Game{DrawMode: DrawOne}
	g.Tableau[0] = []Card{{Suit: Spades, Rank: 1, FaceUp: true}}
	if !g.MoveTableauToFoundation(0) {
		t.Fatal("move should succeed")
	}
	if len(g.Foundation[Spades]) != 1 {
		t.Error("foundation should have 1 card")
	}
	if len(g.Tableau[0]) != 0 {
		t.Error("tableau should be empty")
	}
}

func TestMoveFoundationToTableau(t *testing.T) {
	g := &Game{DrawMode: DrawOne}
	g.Foundation[Hearts] = []Card{{Suit: Hearts, Rank: 3, FaceUp: true}}
	g.Tableau[0] = []Card{{Suit: Spades, Rank: 4, FaceUp: true}}
	if !g.MoveFoundationToTableau(Hearts, 0) {
		t.Fatal("move should succeed")
	}
	if g.Score != -15 {
		t.Errorf("score = %d, want -15", g.Score)
	}
}

func TestWinDetection(t *testing.T) {
	g := &Game{DrawMode: DrawOne}
	for s := Spades; s <= Clubs; s++ {
		for r := 1; r <= 13; r++ {
			g.Foundation[s] = append(g.Foundation[s],
				Card{Suit: s, Rank: r, FaceUp: true})
		}
	}
	g.checkWin()
	if g.State != StateWon {
		t.Error("game should be won")
	}
}

func TestAutoComplete(t *testing.T) {
	g := &Game{DrawMode: DrawOne}
	// Cards stacked K(bottom)→A(top) so auto-complete peels
	// from top: A, 2, 3, …, K.
	for s := Spades; s <= Clubs; s++ {
		col := int(s)
		for r := 13; r >= 1; r-- {
			g.Tableau[col] = append(g.Tableau[col],
				Card{Suit: s, Rank: r, FaceUp: true})
		}
	}
	if !g.AutoComplete() {
		t.Error("auto-complete should move cards")
	}
	if g.State != StateWon {
		t.Error("game should be won after auto-complete")
	}
}

func TestAllFaceUp(t *testing.T) {
	g := &Game{DrawMode: DrawOne}
	g.Tableau[0] = []Card{{Suit: Spades, Rank: 5, FaceUp: true}}
	if !g.AllFaceUp() {
		t.Error("should be all face-up")
	}
	g.Stock = []Card{{Suit: Hearts, Rank: 1}}
	if g.AllFaceUp() {
		t.Error("should not be all face-up with stock")
	}
}

func TestAllFaceUpWithFaceDown(t *testing.T) {
	g := &Game{DrawMode: DrawOne}
	g.Tableau[0] = []Card{
		{Suit: Spades, Rank: 5, FaceUp: false},
		{Suit: Hearts, Rank: 4, FaceUp: true},
	}
	if g.AllFaceUp() {
		t.Error("should not be all face-up with face-down card")
	}
}

func TestVisibleWaste(t *testing.T) {
	g := &Game{DrawMode: DrawThree}
	if len(g.VisibleWaste()) != 0 {
		t.Error("empty waste should return nil")
	}
	g.Waste = []Card{
		{Suit: Spades, Rank: 1, FaceUp: true},
		{Suit: Hearts, Rank: 2, FaceUp: true},
		{Suit: Clubs, Rank: 3, FaceUp: true},
		{Suit: Diamonds, Rank: 4, FaceUp: true},
	}
	vis := g.VisibleWaste()
	if len(vis) != 3 {
		t.Fatalf("visible waste = %d, want 3", len(vis))
	}
	if vis[2].Rank != 4 {
		t.Error("top visible card should be rank 4")
	}
}

func TestRankString(t *testing.T) {
	cases := []struct {
		rank int
		want string
	}{
		{1, "A"}, {2, "2"}, {10, "10"},
		{11, "J"}, {12, "Q"}, {13, "K"},
	}
	for _, tc := range cases {
		c := Card{Rank: tc.rank}
		got := c.RankString()
		if got != tc.want {
			t.Errorf("RankString(%d) = %q, want %q",
				tc.rank, got, tc.want)
		}
	}
}

func TestSuitSymbol(t *testing.T) {
	cases := []struct {
		suit Suit
		want string
	}{
		{Spades, "♠"}, {Hearts, "♥"},
		{Diamonds, "♦"}, {Clubs, "♣"},
	}
	for _, tc := range cases {
		if tc.suit.Symbol() != tc.want {
			t.Errorf("Symbol(%d) = %q, want %q",
				tc.suit, tc.suit.Symbol(), tc.want)
		}
	}
}

func TestSuitIsRed(t *testing.T) {
	if !Hearts.IsRed() {
		t.Error("Hearts should be red")
	}
	if !Diamonds.IsRed() {
		t.Error("Diamonds should be red")
	}
	if Spades.IsRed() {
		t.Error("Spades should not be red")
	}
	if Clubs.IsRed() {
		t.Error("Clubs should not be red")
	}
}

func TestDrawEmptyStockAndWaste(t *testing.T) {
	g := &Game{DrawMode: DrawOne}
	g.Draw() // should be a no-op
	if len(g.Stock) != 0 || len(g.Waste) != 0 {
		t.Error("draw on empty should be no-op")
	}
}
