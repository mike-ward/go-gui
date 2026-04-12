package gui

// rotateMouseInverse transforms screen-space mouse coordinates
// into the unrotated internal space of a rotated container.
// Returns the original coordinates for restoration.
func rotateMouseInverse(s *Shape, e *Event) (origX, origY float32) {
	origX, origY = e.MouseX, e.MouseY
	if s == nil || s.QuarterTurns == 0 {
		return
	}
	e.MouseX, e.MouseY = rotateCoordsInverse(s, origX, origY)
	return
}

// rotateCoordsInverse transforms screen-space coordinates into the
// unrotated internal space of a rotated container without requiring
// an Event allocation.
func rotateCoordsInverse(s *Shape, mx, my float32) (float32, float32) {
	if s == nil || s.QuarterTurns == 0 {
		return mx, my
	}
	cx := s.X + s.Width/2
	cy := s.Y + s.Height/2
	dx, dy := mx-cx, my-cy
	switch s.QuarterTurns {
	case 1: // inverse of 90° CW = 90° CCW
		return cx + dy, cy - dx
	case 2:
		return cx - dx, cy - dy
	case 3: // inverse of 270° CW = 270° CCW
		return cx - dy, cy + dx
	}
	return mx, my
}
