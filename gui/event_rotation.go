package gui

// rotateMouseInverse transforms screen-space mouse coordinates
// into the unrotated internal space of a rotated container.
// Returns the original coordinates for restoration.
func rotateMouseInverse(s *Shape, e *Event) (origX, origY float32) {
	origX, origY = e.MouseX, e.MouseY
	if s == nil || s.QuarterTurns == 0 {
		return
	}
	cx := s.X + s.Width/2
	cy := s.Y + s.Height/2
	dx, dy := e.MouseX-cx, e.MouseY-cy
	switch s.QuarterTurns {
	case 1: // inverse of 90° CW = 90° CCW
		e.MouseX = cx + dy
		e.MouseY = cy - dx
	case 2:
		e.MouseX = cx - dx
		e.MouseY = cy - dy
	case 3: // inverse of 270° CW = 270° CCW
		e.MouseX = cx - dy
		e.MouseY = cy + dx
	}
	return
}
