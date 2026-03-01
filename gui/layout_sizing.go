package gui

import "math"

// distributeMode controls whether space distribution grows or shrinks.
type distributeMode uint8

const (
	distributeGrow   distributeMode = iota
	distributeShrink
)

// distributeAxis selects the dimension.
type distributeAxis uint8

const (
	distributeHorizontal distributeAxis = iota
	distributeVertical
)

// getSize returns width or height depending on axis.
func getSize(shape *Shape, axis distributeAxis) float32 {
	if axis == distributeHorizontal {
		return shape.Width
	}
	return shape.Height
}

// setSize sets width or height depending on axis.
func setSize(shape *Shape, axis distributeAxis, value float32) {
	if axis == distributeHorizontal {
		shape.Width = value
	} else {
		shape.Height = value
	}
}

func getMinSize(shape *Shape, axis distributeAxis) float32 {
	if axis == distributeHorizontal {
		return shape.MinWidth
	}
	return shape.MinHeight
}

func getMaxSize(shape *Shape, axis distributeAxis) float32 {
	if axis == distributeHorizontal {
		return shape.MaxWidth
	}
	return shape.MaxHeight
}

func getSizing(shape *Shape, axis distributeAxis) SizingType {
	if axis == distributeHorizontal {
		return shape.Sizing.Width
	}
	return shape.Sizing.Height
}

type distributionExtrema struct {
	extremum    float32
	nextExtrema float32
}

func collectDistributionCandidates(layout *Layout, axis distributeAxis, mode distributeMode, fillIndices, fixedIndices *[]int) {
	*fillIndices = (*fillIndices)[:0]
	if mode == distributeShrink {
		*fixedIndices = (*fixedIndices)[:0]
	}
	for i := range layout.Children {
		if getSizing(layout.Children[i].Shape, axis) == SizingFill {
			*fillIndices = append(*fillIndices, i)
		} else if mode == distributeShrink {
			*fixedIndices = append(*fixedIndices, i)
		}
	}
}

func shouldContinueDistribution(remaining, prevRemaining float32, mode distributeMode, fillCount int) bool {
	if !f32IsFinite(remaining) || fillCount == 0 {
		return false
	}
	if mode == distributeGrow {
		return remaining > f32Tolerance
	}
	return remaining < -f32Tolerance
}

func findDistributionExtrema(layout *Layout, axis distributeAxis, mode distributeMode, fillIndices, fixedIndices []int) (distributionExtrema, bool) {
	if len(fillIndices) == 0 {
		return distributionExtrema{}, false
	}
	extrema := getSize(layout.Children[fillIndices[0]].Shape, axis)
	var nextExtrema float32
	if mode == distributeGrow {
		nextExtrema = float32(math.MaxUint32)
	}

	for _, idx := range fillIndices {
		childSize := getSize(layout.Children[idx].Shape, axis)
		if mode == distributeGrow {
			if childSize < extrema {
				nextExtrema = extrema
				extrema = childSize
			} else if childSize > extrema {
				nextExtrema = f32Min(nextExtrema, childSize)
			}
		} else {
			if childSize > extrema {
				nextExtrema = extrema
				extrema = childSize
			} else if childSize < extrema {
				nextExtrema = f32Max(nextExtrema, childSize)
			}
		}
	}
	if mode == distributeShrink {
		for _, idx := range fixedIndices {
			childSize := getSize(layout.Children[idx].Shape, axis)
			if childSize > extrema {
				nextExtrema = extrema
				extrema = childSize
			} else if childSize < extrema {
				nextExtrema = f32Max(nextExtrema, childSize)
			}
		}
	}
	if !f32IsFinite(extrema) || !f32IsFinite(nextExtrema) {
		return distributionExtrema{}, false
	}
	return distributionExtrema{extremum: extrema, nextExtrema: nextExtrema}, true
}

func computeDistributionDelta(layout *Layout, remaining float32, mode distributeMode, axis distributeAxis, extrema distributionExtrema, fillCount, fixedCount int) (float32, bool) {
	var sizeDelta float32
	if mode == distributeGrow {
		if extrema.nextExtrema == float32(math.MaxUint32) {
			sizeDelta = remaining
		} else {
			sizeDelta = extrema.nextExtrema - extrema.extremum
		}
	} else {
		if extrema.extremum > 0 {
			if extrema.nextExtrema == 0 {
				sizeDelta = remaining
			} else {
				sizeDelta = extrema.nextExtrema - extrema.extremum
			}
		} else {
			sizeDelta = remaining
		}
	}
	if !f32IsFinite(sizeDelta) {
		return 0, false
	}
	if mode == distributeGrow {
		sizeDelta = f32Min(sizeDelta, remaining/float32(fillCount))
	} else {
		totalCount := fillCount + fixedCount
		if totalCount > 0 {
			sizeDelta = f32Max(sizeDelta, remaining/float32(totalCount))
		}
	}
	if !f32IsFinite(sizeDelta) {
		return 0, false
	}
	saneDeltaLimit := f32Max(f32Abs(getSize(layout.Shape, axis)), f32Abs(remaining))
	saneDeltaLimit = f32Max(saneDeltaLimit*4, 1_000_000)
	if !f32IsFinite(saneDeltaLimit) || saneDeltaLimit <= 0 {
		return 0, false
	}
	return f32Clamp(sizeDelta, -saneDeltaLimit, saneDeltaLimit), true
}

func applyDistributionDelta(layout *Layout, axis distributeAxis, extremum, sizeDelta, remainingIn float32, fillIndices *[]int) (float32, bool) {
	remaining := remainingIn
	keepIdx := 0
	fi := *fillIndices
	for i := range fi {
		idx := fi[i]
		child := &layout.Children[idx]
		keepChild := true
		childSize := getSize(child.Shape, axis)
		if childSize == extremum {
			prevSize := childSize
			newSize := childSize + sizeDelta
			if !f32IsFinite(newSize) {
				return 0, false
			}
			setSize(child.Shape, axis, newSize)

			constrained := false
			minSize := getMinSize(child.Shape, axis)
			maxSize := getMaxSize(child.Shape, axis)
			currentSize := getSize(child.Shape, axis)
			if currentSize <= minSize {
				setSize(child.Shape, axis, minSize)
				constrained = true
			} else if maxSize > 0 && currentSize >= maxSize {
				setSize(child.Shape, axis, maxSize)
				constrained = true
			}
			remaining -= getSize(child.Shape, axis) - prevSize
			if !f32IsFinite(remaining) {
				return 0, false
			}
			if constrained {
				keepChild = false
			}
		}
		if keepChild {
			if keepIdx != i {
				fi[keepIdx] = idx
			}
			keepIdx++
		}
	}
	*fillIndices = fi[:keepIdx]
	return remaining, true
}

func distributeSpace(layout *Layout, remainingIn float32, mode distributeMode, axis distributeAxis, candidates, fixedIndices *[]int) float32 {
	if !f32IsFinite(remainingIn) {
		return 0
	}
	remaining := remainingIn
	prevRemaining := float32(0)

	collectDistributionCandidates(layout, axis, mode, candidates, fixedIndices)

	for {
		if !shouldContinueDistribution(remaining, prevRemaining, mode, len(*candidates)) {
			break
		}
		if f32AreClose(remaining, prevRemaining) {
			break
		}
		prevRemaining = remaining
		extrema, ok := findDistributionExtrema(layout, axis, mode, *candidates, *fixedIndices)
		if !ok {
			break
		}
		sizeDelta, ok := computeDistributionDelta(layout, remaining, mode, axis, extrema, len(*candidates), len(*fixedIndices))
		if !ok {
			break
		}
		remaining, ok = applyDistributionDelta(layout, axis, extrema.extremum, sizeDelta, remaining, candidates)
		if !ok {
			break
		}
	}
	return remaining
}

// layoutWidths arranges children horizontally (bottom-up).
func layoutWidths(layout *Layout) {
	padding := layout.Shape.PaddingWidth()
	if layout.Shape.Axis == AxisLeftToRight {
		sp := layout.spacing()
		if layout.Shape.Sizing.Width == SizingFixed {
			for i := range layout.Children {
				layoutWidths(&layout.Children[i])
			}
		} else {
			minWidths := padding + sp
			for i := range layout.Children {
				layoutWidths(&layout.Children[i])
				layout.Shape.Width += layout.Children[i].Shape.Width
				if layout.Shape.Wrap || layout.Shape.Overflow {
					minWidths = f32Max(minWidths, layout.Children[i].Shape.Width+padding)
				} else if !layout.Shape.Clip {
					minWidths += layout.Children[i].Shape.MinWidth
				}
			}
			if !layout.Shape.Wrap && !layout.Shape.Overflow {
				layout.Shape.MinWidth = f32Max(minWidths, layout.Shape.MinWidth+padding+sp)
			} else {
				layout.Shape.MinWidth = f32Max(minWidths, layout.Shape.MinWidth)
			}
			layout.Shape.Width += padding + sp
			if layout.Shape.MaxWidth > 0 {
				layout.Shape.Width = f32Min(layout.Shape.MaxWidth, layout.Shape.Width)
				layout.Shape.MinWidth = f32Min(layout.Shape.MaxWidth, layout.Shape.MinWidth)
			}
			if layout.Shape.MinWidth > 0 {
				layout.Shape.Width = f32Max(layout.Shape.MinWidth, layout.Shape.Width)
			}
		}
	} else if layout.Shape.Axis == AxisTopToBottom {
		for i := range layout.Children {
			layoutWidths(&layout.Children[i])
			if layout.Shape.Sizing.Width != SizingFixed {
				layout.Shape.Width = f32Max(layout.Shape.Width, layout.Children[i].Shape.Width+padding)
				if !layout.Shape.Clip {
					layout.Shape.MinWidth = f32Max(layout.Shape.MinWidth, layout.Children[i].Shape.MinWidth+padding)
				}
			}
		}
		if layout.Shape.MinWidth > 0 {
			layout.Shape.Width = f32Max(layout.Shape.Width, layout.Shape.MinWidth)
		}
		if layout.Shape.MaxWidth > 0 {
			layout.Shape.Width = f32Min(layout.Shape.Width, layout.Shape.MaxWidth)
		}
	}
}

// layoutHeights arranges children vertically (bottom-up).
func layoutHeights(layout *Layout) {
	padding := layout.Shape.PaddingHeight()
	if layout.Shape.Axis == AxisTopToBottom {
		sp := layout.spacing()
		if layout.Shape.Sizing.Height == SizingFixed {
			for i := range layout.Children {
				layoutHeights(&layout.Children[i])
			}
		} else {
			minHeights := padding + sp
			for i := range layout.Children {
				layoutHeights(&layout.Children[i])
				layout.Shape.Height += layout.Children[i].Shape.Height
				minHeights += layout.Children[i].Shape.MinHeight
			}
			layout.Shape.MinHeight = f32Max(minHeights, layout.Shape.MinHeight+padding+sp)
			layout.Shape.Height += padding + sp
			if layout.Shape.MaxHeight > 0 {
				layout.Shape.Height = f32Min(layout.Shape.MaxHeight, layout.Shape.Height)
				layout.Shape.MinHeight = f32Min(layout.Shape.MaxHeight, layout.Shape.MinHeight)
			}
			if layout.Shape.MinHeight > 0 {
				layout.Shape.Height = f32Max(layout.Shape.MinHeight, layout.Shape.Height)
			}
			if layout.Shape.Sizing.Height == SizingFill && layout.Shape.IDScroll > 0 {
				layout.Shape.MinHeight = spacingSmall
			}
		}
	} else if layout.Shape.Axis == AxisLeftToRight {
		for i := range layout.Children {
			layoutHeights(&layout.Children[i])
			if layout.Shape.Sizing.Height != SizingFixed {
				layout.Shape.Height = f32Max(layout.Shape.Height, layout.Children[i].Shape.Height+padding)
				layout.Shape.MinHeight = f32Max(layout.Shape.MinHeight, layout.Children[i].Shape.MinHeight+padding)
			}
		}
		if layout.Shape.MinHeight > 0 {
			layout.Shape.Height = f32Max(layout.Shape.Height, layout.Shape.MinHeight)
		}
		if layout.Shape.MaxHeight > 0 {
			layout.Shape.Height = f32Min(layout.Shape.Height, layout.Shape.MaxHeight)
		}
	}
}

// spacingSmall matches the V framework's spacing_small constant.
const spacingSmall = 5

// layoutFillWidths manages horizontal growth/shrinkage.
func layoutFillWidths(layout *Layout) {
	var candidates, fixedIndices []int
	layoutFillWidthsImpl(layout, &candidates, &fixedIndices, true)
}

func layoutFillWidthsImpl(layout *Layout, candidates, fixedIndices *[]int, isRoot bool) {
	remainingWidth := layout.Shape.Width - layout.Shape.PaddingWidth()

	if layout.Shape.Axis == AxisLeftToRight {
		for i := range layout.Children {
			remainingWidth -= layout.Children[i].Shape.Width
		}
		remainingWidth -= layout.spacing()

		if remainingWidth > f32Tolerance {
			remainingWidth = distributeSpace(layout, remainingWidth, distributeGrow, distributeHorizontal, candidates, fixedIndices)
		}
		if remainingWidth < -f32Tolerance && !layout.Shape.Wrap && !layout.Shape.Overflow {
			remainingWidth = distributeSpace(layout, remainingWidth, distributeShrink, distributeHorizontal, candidates, fixedIndices)
		}
	} else if layout.Shape.Axis == AxisTopToBottom {
		if layout.Shape.IDScroll > 0 && layout.Shape.Sizing.Width == SizingFill &&
			layout.Shape.ScrollMode != ScrollVerticalOnly &&
			layout.Parent != nil && layout.Parent.Shape.Axis == AxisLeftToRight {
			var totalChildWidth float32
			for j := range layout.Parent.Children {
				totalChildWidth += layout.Parent.Children[j].Shape.Width
			}
			siblingWidthSum := totalChildWidth - layout.Shape.Width
			targetWidth := layout.Parent.Shape.Width - siblingWidthSum - layout.Parent.spacing() - layout.Parent.Shape.PaddingWidth()
			layout.Shape.Width = f32Max(0, targetWidth)
		}
		if layout.Shape.MinWidth > 0 && layout.Shape.Width < layout.Shape.MinWidth {
			layout.Shape.Width = layout.Shape.MinWidth
		}
		if layout.Shape.MaxWidth > 0 && layout.Shape.Width > layout.Shape.MaxWidth {
			layout.Shape.Width = layout.Shape.MaxWidth
		}
		for i := range layout.Children {
			if layout.Children[i].Shape.Sizing.Width == SizingFill {
				layout.Children[i].Shape.Width = remainingWidth
				if layout.Children[i].Shape.MinWidth > 0 {
					layout.Children[i].Shape.Width = f32Max(layout.Children[i].Shape.Width, layout.Children[i].Shape.MinWidth)
				}
				if layout.Children[i].Shape.MaxWidth > 0 {
					layout.Children[i].Shape.Width = f32Min(layout.Children[i].Shape.Width, layout.Children[i].Shape.MaxWidth)
				}
			}
		}
	}

	for i := range layout.Children {
		layoutFillWidthsImpl(&layout.Children[i], candidates, fixedIndices, false)
	}
}

// layoutFillHeights manages vertical growth/shrinkage.
func layoutFillHeights(layout *Layout) {
	var candidates, fixedIndices []int
	layoutFillHeightsImpl(layout, &candidates, &fixedIndices, true)
}

func layoutFillHeightsImpl(layout *Layout, candidates, fixedIndices *[]int, isRoot bool) {
	remainingHeight := layout.Shape.Height - layout.Shape.PaddingHeight()

	if layout.Shape.Axis == AxisTopToBottom {
		for i := range layout.Children {
			remainingHeight -= layout.Children[i].Shape.Height
		}
		remainingHeight -= layout.spacing()

		if remainingHeight > f32Tolerance {
			remainingHeight = distributeSpace(layout, remainingHeight, distributeGrow, distributeVertical, candidates, fixedIndices)
		}
		if remainingHeight < -f32Tolerance {
			remainingHeight = distributeSpace(layout, remainingHeight, distributeShrink, distributeVertical, candidates, fixedIndices)
		}
	} else if layout.Shape.Axis == AxisLeftToRight {
		if layout.Shape.IDScroll > 0 && layout.Shape.Sizing.Height == SizingFill &&
			layout.Shape.ScrollMode != ScrollHorizontalOnly &&
			layout.Parent != nil && layout.Parent.Shape.Axis == AxisTopToBottom {
			var totalChildHeight float32
			for j := range layout.Parent.Children {
				totalChildHeight += layout.Parent.Children[j].Shape.Height
			}
			siblingHeightSum := totalChildHeight - layout.Shape.Height
			targetHeight := layout.Parent.Shape.Height - siblingHeightSum - layout.Parent.spacing() - layout.Parent.Shape.PaddingHeight()
			layout.Shape.Height = f32Max(0, targetHeight)
		}
		if layout.Shape.MinHeight > 0 && layout.Shape.Height < layout.Shape.MinHeight {
			layout.Shape.Height = layout.Shape.MinHeight
		}
		if layout.Shape.MaxHeight > 0 && layout.Shape.Height > layout.Shape.MaxHeight {
			layout.Shape.Height = layout.Shape.MaxHeight
		}
		for i := range layout.Children {
			if layout.Children[i].Shape.Sizing.Height == SizingFill {
				layout.Children[i].Shape.Height = remainingHeight
				if layout.Children[i].Shape.MinHeight > 0 {
					layout.Children[i].Shape.Height = f32Max(layout.Children[i].Shape.Height, layout.Children[i].Shape.MinHeight)
				}
				if layout.Children[i].Shape.MaxHeight > 0 {
					layout.Children[i].Shape.Height = f32Min(layout.Children[i].Shape.Height, layout.Children[i].Shape.MaxHeight)
				}
			}
		}
	}

	for i := range layout.Children {
		layoutFillHeightsImpl(&layout.Children[i], candidates, fixedIndices, false)
	}
}
