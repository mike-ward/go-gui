package svg

// xml_use.go — resolves SVG <use href="#id"> references by inlining
// a clone of the target subtree. <symbol> targets surface as their
// children wrapped in a synthesized <g>; cycles are guarded by a
// per-path visited set and a hard depth cap.

import (
	"maps"
	"strconv"
	"strings"

	"github.com/mike-ward/go-gui/gui"
)

// maxUseDepth bounds nested <use> resolution. Real-world SVGs nest
// 1–3 deep (icon sets that use symbol-of-symbol composition).
const maxUseDepth = 8

// maxUseExpandClones caps the total node count produced by <use>
// expansion. The decode-side maxElements limit bounds the source
// tree; this companion budget bounds the post-expansion blowup so a
// single 1KB <symbol> referenced 10 000 times cannot multiply into a
// hundred-million-node tree before any later pass can reject it.
const maxUseExpandClones = maxElements

// expandUseElements rewrites root in place: every <use href="#id">
// element is replaced with a synthesized <g> that wraps a clone of
// the referenced subtree. The clone has its id stripped so the
// post-expansion tree carries no duplicate ids.
func expandUseElements(root *xmlNode) {
	if root == nil || !hasUseElement(root) {
		return
	}
	idIndex := make(map[string]*xmlNode)
	indexByID(root, idIndex)
	visited := make(map[string]struct{}, 4)
	budget := maxUseExpandClones
	expandUseRec(root, idIndex, visited, 0, &budget)
}

// hasUseElement returns true if the subtree contains any <use> node.
// Most icon SVGs have none; skipping the id-index walk keeps parse
// hot path lean.
func hasUseElement(n *xmlNode) bool {
	for i := range n.Children {
		c := &n.Children[i]
		if c.Name == "use" {
			return true
		}
		if hasUseElement(c) {
			return true
		}
	}
	return false
}

func indexByID(n *xmlNode, out map[string]*xmlNode) {
	if id := n.AttrMap["id"]; id != "" {
		if _, exists := out[id]; !exists {
			out[id] = n
		}
	}
	for i := range n.Children {
		indexByID(&n.Children[i], out)
	}
}

func expandUseRec(n *xmlNode, idIndex map[string]*xmlNode,
	visited map[string]struct{}, depth int, budget *int) {
	for i := range n.Children {
		if *budget <= 0 {
			return
		}
		c := &n.Children[i]
		if c.Name == "use" {
			if depth >= maxUseDepth {
				continue
			}
			href := c.AttrMap["href"]
			if href == "" {
				href = c.AttrMap["xlink:href"]
			}
			if !strings.HasPrefix(href, "#") || len(href) <= 1 {
				continue
			}
			id := href[1:]
			if _, cyc := visited[id]; cyc {
				continue
			}
			target, ok := idIndex[id]
			if !ok {
				continue
			}
			visited[id] = struct{}{}
			*c = synthesizeUseGroup(c, target, budget)
			expandUseRec(c, idIndex, visited, depth+1, budget)
			delete(visited, id)
			continue
		}
		expandUseRec(c, idIndex, visited, depth, budget)
	}
}

// synthesizeUseGroup builds a <g> wrapper that replaces a <use>
// element. The wrapper carries the <use>'s presentation attrs (fill,
// class, style, ...) so they cascade into the inlined clone, plus a
// translate(x,y) transform composed with any author-supplied
// transform on the <use>. <symbol> targets contribute their children;
// other targets are inlined as a single cloned subtree.
func synthesizeUseGroup(useNode, target *xmlNode, budget *int) xmlNode {
	x := useNode.AttrMap["x"]
	y := useNode.AttrMap["y"]

	var children []xmlNode
	if target.Name == "symbol" {
		children = cloneChildrenForUse(target.Children, budget)
	} else {
		clone := cloneNode(target, budget)
		stripID(&clone)
		children = []xmlNode{clone}
	}

	gAttrs := make([]xmlAttr, 0, len(useNode.Attrs))
	gAttrMap := make(map[string]string, len(useNode.AttrMap))
	for _, a := range useNode.Attrs {
		switch a.Name {
		case "x", "y", "href", "xlink:href", "id", "width", "height":
			continue
		}
		gAttrs = append(gAttrs, a)
		gAttrMap[a.Name] = a.Value
	}

	posTransform := positioningTransform(useNode, target, x, y)
	if posTransform != "" {
		composed := posTransform
		if existing, ok := gAttrMap["transform"]; ok {
			composed = existing + " " + posTransform
		}
		gAttrMap["transform"] = composed
		replaced := false
		for i := range gAttrs {
			if gAttrs[i].Name == "transform" {
				gAttrs[i].Value = composed
				replaced = true
				break
			}
		}
		if !replaced {
			gAttrs = append(gAttrs,
				xmlAttr{Name: "transform", Value: composed})
		}
	}

	g := xmlNode{
		Name:     "g",
		Attrs:    gAttrs,
		AttrMap:  gAttrMap,
		Children: children,
	}
	g.OpenTag = buildOpenTag(g.Name, g.Attrs, false)
	return g
}

// positioningTransform returns the SVG transform string that places
// the synthesized <g> at the <use>'s (x,y) and, when the target is a
// <symbol> with a viewBox plus author-specified width/height on the
// <use>, scales the symbol's viewport to fit the requested box per
// preserveAspectRatio.
//
// SVG transform list reads left-to-right but applies right-to-left to
// child points: child p gets translate(-vbX,-vbY) first, then scale,
// then translate(x+ax, y+ay). Non-symbol targets and symbols without a
// viewBox (or without numeric use width/height) get only the
// translate. ax/ay are align offsets; with align=none they're zero
// and the scale is non-uniform — the SVG-1.1 default for <symbol> is
// xMidYMid meet, so authors who relied on stretch must opt in via
// preserveAspectRatio="none" on the symbol.
//
// Slice scaling uses uniform max(sx,sy). TODO: emit clip-to-use-box
// (spec requires it; without the clip, slice content can overflow).
//
// Parses x/y numerically rather than splicing raw author strings into
// the transform attribute: `<use x="0)scale(99)">` would otherwise
// inject extra transforms. parseLength clamps to ±maxCoordinate and
// rejects NaN/Inf.
func positioningTransform(useNode, target *xmlNode, x, y string) string {
	if useNode == nil || target == nil {
		return ""
	}
	// Percentages on x/y resolve against the parent viewport, which is
	// out of scope here. Drop rather than treat "50%" as raw 50.
	if hasPercent(x) || hasPercent(y) {
		return ""
	}
	wantTranslate := x != "" || y != ""
	tx := parseLength(x)
	ty := parseLength(y)

	sx, sy, vbX, vbY, ax, ay, ok := symbolViewportScale(useNode, target)
	if !ok {
		if !wantTranslate {
			return ""
		}
		return writeTranslate(tx, ty)
	}
	combX, combY := tx+ax, ty+ay
	if !boundedScale(combX) || !boundedScale(combY) {
		return ""
	}
	var b strings.Builder
	b.Grow(64)
	writeTranslateTo(&b, combX, combY)
	switch {
	case sx == sy && sx != 1:
		b.WriteString(" scale(")
		writeF32(&b, sx)
		b.WriteByte(')')
	case sx != 1 || sy != 1:
		b.WriteString(" scale(")
		writeF32(&b, sx)
		b.WriteByte(',')
		writeF32(&b, sy)
		b.WriteByte(')')
	}
	if vbX != 0 || vbY != 0 {
		b.WriteString(" ")
		writeTranslateTo(&b, -vbX, -vbY)
	}
	return b.String()
}

func writeTranslate(x, y float32) string {
	var b strings.Builder
	b.Grow(32)
	writeTranslateTo(&b, x, y)
	return b.String()
}

func writeTranslateTo(b *strings.Builder, x, y float32) {
	b.WriteString("translate(")
	writeF32(b, x)
	b.WriteByte(',')
	writeF32(b, y)
	b.WriteByte(')')
}

func writeF32(b *strings.Builder, v float32) {
	var buf [32]byte
	b.Write(strconv.AppendFloat(buf[:0], float64(v), 'f', -1, 32))
}

// symbolViewportScale returns the scale factors, viewBox origin, and
// alignment offsets needed to map a <symbol viewBox=...>'s viewport
// into the <use>'s width/height box per preserveAspectRatio.
//
// Default is xMidYMid meet (SVG 1.1): uniform min-scale + center.
// preserveAspectRatio="none" on the symbol opts into the legacy
// independent-axis stretch. Slice scaling uses uniform max-scale; the
// accompanying clip-to-box is unimplemented (overflow not enforced).
//
// ok is false when the target is not a <symbol>, when the symbol has
// no usable viewBox, when use width/height are missing or
// percentage-based, or when the resulting scale would exceed
// ±maxCoordinate (pathological viewBox like 1e-30 against use width=1).
func symbolViewportScale(useNode, target *xmlNode) (
	sx, sy, vbX, vbY, ax, ay float32, ok bool,
) {
	if target.Name != "symbol" {
		return
	}
	vb := target.AttrMap["viewBox"]
	if vb == "" {
		vb = target.AttrMap["viewbox"]
	}
	useW := useNode.AttrMap["width"]
	useH := useNode.AttrMap["height"]
	if vb == "" || (useW == "" && useH == "") {
		return
	}
	if hasPercent(useW) || hasPercent(useH) {
		return
	}
	nums := parseNumberList(vb)
	if len(nums) < 4 || nums[2] <= 0 || nums[3] <= 0 {
		return
	}
	uw := parseLength(useW)
	uh := parseLength(useH)
	if useW == "" {
		uw = nums[2]
	}
	if useH == "" {
		uh = nums[3]
	}
	if uw <= 0 || uh <= 0 {
		return
	}
	rawSx := uw / nums[2]
	rawSy := uh / nums[3]

	align, slice := effectivePreserveAspectRatio(target)
	if align == gui.SvgAlignNone {
		sx, sy = rawSx, rawSy
	} else {
		var s float32
		if slice {
			s = max(rawSx, rawSy)
		} else {
			s = min(rawSx, rawSy)
		}
		sx, sy = s, s
		xFrac, yFrac := gui.PreserveAlignFractions(align)
		ax = xFrac * (uw - nums[2]*s)
		ay = yFrac * (uh - nums[3]*s)
	}
	if !boundedScale(sx) || !boundedScale(sy) {
		return 0, 0, 0, 0, 0, 0, false
	}
	return sx, sy, nums[0], nums[1], ax, ay, true
}

// effectivePreserveAspectRatio returns the preserveAspectRatio value
// for a <symbol> referenced by <use>. SVG 1.1 only honors the
// symbol's own attribute (default xMidYMid meet); SVG 2 added
// override on <use> but is not yet supported here.
func effectivePreserveAspectRatio(target *xmlNode) (gui.SvgAlign, bool) {
	if target == nil || target.Name != "symbol" {
		return gui.SvgAlignXMidYMid, false
	}
	if v, ok := target.AttrMap["preserveAspectRatio"]; ok && v != "" {
		return parsePreserveAspectRatio(v)
	}
	return gui.SvgAlignXMidYMid, false
}

func boundedScale(v float32) bool {
	if !finiteF32(v) {
		return false
	}
	if v < 0 {
		v = -v
	}
	return v <= maxCoordinate
}

func hasPercent(s string) bool {
	return strings.IndexByte(s, '%') >= 0
}

func cloneChildrenForUse(src []xmlNode, budget *int) []xmlNode {
	if len(src) == 0 {
		return nil
	}
	out := make([]xmlNode, 0, len(src))
	for i := range src {
		if *budget <= 0 {
			break
		}
		c := cloneNode(&src[i], budget)
		stripID(&c)
		out = append(out, c)
	}
	return out
}

// cloneNode copies n into a fresh subtree, decrementing budget per
// node. When budget is exhausted the returned subtree is truncated
// (children dropped) so a malicious <use> fanout cannot inflate the
// tree past maxUseExpandClones.
func cloneNode(n *xmlNode, budget *int) xmlNode {
	if *budget <= 0 {
		return xmlNode{}
	}
	*budget--
	out := xmlNode{
		Name:      n.Name,
		OpenTag:   n.OpenTag,
		Leading:   n.Leading,
		Text:      n.Text,
		Tail:      n.Tail,
		SelfClose: n.SelfClose,
	}
	if len(n.Attrs) > 0 {
		out.Attrs = make([]xmlAttr, len(n.Attrs))
		copy(out.Attrs, n.Attrs)
	}
	if len(n.AttrMap) > 0 {
		out.AttrMap = make(map[string]string, len(n.AttrMap))
		maps.Copy(out.AttrMap, n.AttrMap)
	}
	if len(n.Children) > 0 {
		out.Children = make([]xmlNode, 0, len(n.Children))
		for i := range n.Children {
			if *budget <= 0 {
				break
			}
			out.Children = append(out.Children, cloneNode(&n.Children[i], budget))
		}
	}
	return out
}

// stripID removes id attributes from n and every descendant. Cloned
// <use> subtrees must not duplicate ids: leftover descendant ids would
// collide with the original's ids and corrupt url(#id) resolution,
// CSS #id selector matching, and animation targeting.
func stripID(n *xmlNode) {
	if _, ok := n.AttrMap["id"]; ok {
		delete(n.AttrMap, "id")
		filtered := make([]xmlAttr, 0, len(n.Attrs))
		for _, a := range n.Attrs {
			if a.Name != "id" {
				filtered = append(filtered, a)
			}
		}
		n.Attrs = filtered
		n.OpenTag = buildOpenTag(n.Name, n.Attrs, n.SelfClose)
	}
	for i := range n.Children {
		stripID(&n.Children[i])
	}
}
