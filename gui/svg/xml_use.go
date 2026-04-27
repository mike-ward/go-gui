package svg

// xml_use.go — resolves SVG <use href="#id"> references by inlining
// a clone of the target subtree. <symbol> targets surface as their
// children wrapped in a synthesized <g>; cycles are guarded by a
// per-path visited set and a hard depth cap.

import (
	"maps"
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

	if x != "" || y != "" {
		if x == "" {
			x = "0"
		}
		if y == "" {
			y = "0"
		}
		translate := "translate(" + x + "," + y + ")"
		if existing, ok := gAttrMap["transform"]; ok {
			translate = existing + " " + translate
		}
		gAttrMap["transform"] = translate
		replaced := false
		for i := range gAttrs {
			if gAttrs[i].Name == "transform" {
				gAttrs[i].Value = translate
				replaced = true
				break
			}
		}
		if !replaced {
			gAttrs = append(gAttrs,
				xmlAttr{Name: "transform", Value: translate})
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

func stripID(n *xmlNode) {
	if _, ok := n.AttrMap["id"]; !ok {
		return
	}
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
