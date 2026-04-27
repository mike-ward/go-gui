package svg

// xml_use.go — resolves SVG <use href="#id"> references by inlining
// a clone of the target subtree. <symbol> targets surface as their
// children wrapped in a synthesized <g>; cycles are guarded by a
// per-path visited set and a hard depth cap.

import (
	"maps"
	"strconv"
	"strings"
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
// <use>, scales the symbol's viewport to fill the requested box.
//
// SVG transform list reads left-to-right but applies right-to-left to
// child points: child p gets translate(-vbX,-vbY) first, then scale,
// then translate(x,y). Non-symbol targets and symbols without a
// viewBox (or without numeric use width/height) get only the
// translate.
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
	translate := "translate(" + f32ToString(tx) + "," + f32ToString(ty) + ")"

	sx, sy, vbX, vbY, ok := symbolViewportScale(useNode, target)
	if !ok {
		if !wantTranslate {
			return ""
		}
		return translate
	}
	out := translate
	if sx != 1 || sy != 1 {
		out += " scale(" + f32ToString(sx) + "," + f32ToString(sy) + ")"
	}
	if vbX != 0 || vbY != 0 {
		out += " translate(" + f32ToString(-vbX) + "," + f32ToString(-vbY) + ")"
	}
	return out
}

// symbolViewportScale returns the scale factors and viewBox origin
// needed to map a <symbol viewBox=...>'s viewport to the <use>'s
// width/height box. ok is false when the target is not a <symbol>,
// when the symbol has no usable viewBox, when use width/height are
// missing or percentage-based, or when the resulting scale would
// exceed ±maxCoordinate (pathological viewBox like 1e-30 against
// use width=1).
func symbolViewportScale(useNode, target *xmlNode) (sx, sy, vbX, vbY float32, ok bool) {
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
	sx = uw / nums[2]
	sy = uh / nums[3]
	if !boundedScale(sx) || !boundedScale(sy) {
		return 0, 0, 0, 0, false
	}
	return sx, sy, nums[0], nums[1], true
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

func f32ToString(v float32) string {
	return strconv.FormatFloat(float64(v), 'f', -1, 32)
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
