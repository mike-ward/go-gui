package markdown

// walker.go converts markdown source to []Block using
// goldmark as the parser backend.

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	east "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
	emoji "github.com/yuin/goldmark-emoji"
	emast "github.com/yuin/goldmark-emoji/ast"
)

// Parse converts markdown source to []Block.
func Parse(source string, hardLineBreaks bool) []Block {
	abbrDefs := collectAbbrDefs(source)
	footnoteDefs := collectFootnoteDefs(source)
	source = preprocessSource(source)
	src := []byte(source)

	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.DefinitionList,
			emoji.Emoji,
		),
		goldmark.WithParserOptions(
			parser.WithInlineParsers(
				util.Prioritized(&mathInlineParser{}, 100),
				util.Prioritized(&highlightParser{}, 100),
				util.Prioritized(&underlineParser{}, 100),
				util.Prioritized(&superscriptParser{}, 100),
				util.Prioritized(&subscriptParser{}, 100),
			),
		),
	)

	doc := md.Parser().Parse(text.NewReader(src))

	w := &mdWalker{
		source:       src,
		footnoteDefs: footnoteDefs,
		hardBreaks:   hardLineBreaks,
	}
	w.walkDocument(doc)

	for i := range w.blocks {
		w.blocks[i].Runs = mergeAdjacentRuns(
			w.blocks[i].Runs)
		if !w.blocks[i].IsCode {
			applyTypography(w.blocks[i].Runs)
		}
		if len(footnoteDefs) > 0 {
			w.blocks[i].Runs = applyFootnoteRefs(
				w.blocks[i].Runs, footnoteDefs)
		}
		if len(abbrDefs) > 0 {
			w.blocks[i].Runs = replaceAbbreviations(
				w.blocks[i].Runs, abbrDefs)
		}
	}
	return w.blocks
}

// mdWalker walks a goldmark AST producing Blocks.
type mdWalker struct {
	source       []byte
	footnoteDefs map[string]string
	hardBreaks   bool
	blocks       []Block
	bqDepth      int
	listDepth    int
}

// inlineState tracks formatting context during inline walking.
type inlineState struct {
	format        Format
	strikethrough bool
	highlight     bool
	underline     bool
	superscript   bool
	subscript     bool
	link          string
	tooltip       string
}

func (w *mdWalker) walkDocument(doc ast.Node) {
	for c := doc.FirstChild(); c != nil; c = c.NextSibling() {
		w.walkBlock(c)
	}
}

func (w *mdWalker) walkBlock(node ast.Node) {
	switch node.Kind() {
	case ast.KindParagraph:
		w.walkParagraph(node)
	case ast.KindHeading:
		w.walkHeading(node.(*ast.Heading))
	case ast.KindThematicBreak:
		w.blocks = append(w.blocks, Block{IsHR: true})
	case ast.KindFencedCodeBlock:
		w.walkFencedCode(node.(*ast.FencedCodeBlock))
	case ast.KindCodeBlock:
		w.walkCodeBlock(node.(*ast.CodeBlock))
	case ast.KindBlockquote:
		w.walkBlockquote(node.(*ast.Blockquote))
	case ast.KindList:
		w.walkList(node.(*ast.List))
	case ast.KindHTMLBlock:
		// Skip raw HTML blocks.
	default:
		switch node.Kind() {
		case east.KindTable:
			w.walkTable(node.(*east.Table))
		case east.KindDefinitionList:
			w.walkDefList(node)
		default:
			for c := node.FirstChild(); c != nil; c = c.NextSibling() {
				w.walkBlock(c)
			}
		}
	}
}

func (w *mdWalker) walkParagraph(node ast.Node) {
	// Standalone image → image block.
	if node.ChildCount() == 1 &&
		node.FirstChild().Kind() == ast.KindImage {
		w.walkImage(node.FirstChild().(*ast.Image))
		return
	}
	// Standalone display math → math block.
	if node.ChildCount() == 1 &&
		node.FirstChild().Kind() == NodeKindMathDisplay {
		dm := node.FirstChild().(*nodeMathDisplay)
		w.blocks = append(w.blocks, Block{
			IsMath:    true,
			MathLatex: dm.Latex,
		})
		return
	}
	runs := w.collectRuns(node, inlineState{})
	runs = trimTrailingBreaks(runs)
	if len(runs) > 0 {
		w.blocks = append(w.blocks, Block{Runs: runs})
	}
}

func (w *mdWalker) walkHeading(h *ast.Heading) {
	runs := w.collectRuns(h, inlineState{})
	runs = trimTrailingBreaks(runs)
	slug := HeadingSlug(RunsToText(runs))
	w.blocks = append(w.blocks, Block{
		HeaderLevel: h.Level,
		AnchorSlug:  slug,
		Runs:        runs,
	})
}

func (w *mdWalker) walkFencedCode(fcb *ast.FencedCodeBlock) {
	langHint := normalizeLanguageHint(
		string(fcb.Language(w.source)))
	codeText := w.collectLines(fcb)

	if langHint == "math" {
		w.blocks = append(w.blocks, Block{
			IsMath:    true,
			MathLatex: codeText,
		})
		return
	}
	w.emitCodeBlock(codeText, langHint)
}

func (w *mdWalker) walkCodeBlock(cb *ast.CodeBlock) {
	w.emitCodeBlock(w.collectLines(cb), "")
}

func (w *mdWalker) emitCodeBlock(code, langHint string) {
	lang := LangFromHint(langHint)
	tokens := tokenizeCode(code, lang,
		maxCodeBlockHighlightBytes)
	var runs []Run
	if len(tokens) == 0 {
		runs = []Run{{Text: code, Format: FormatCode}}
	} else {
		runs = make([]Run, len(tokens))
		for i, tok := range tokens {
			runs[i] = Run{
				Text:      code[tok.Start:tok.End],
				Format:    FormatCode,
				CodeToken: tok.Kind,
			}
		}
	}
	w.blocks = append(w.blocks, Block{
		IsCode:       true,
		CodeLanguage: langHint,
		Runs:         runs,
	})
}

// collectLines extracts text from a block node's Lines().
func (w *mdWalker) collectLines(node ast.Node) string {
	type hasLines interface {
		Lines() *text.Segments
	}
	hl, ok := node.(hasLines)
	if !ok {
		return ""
	}
	segs := hl.Lines()
	var sb strings.Builder
	for i := 0; i < segs.Len(); i++ {
		seg := segs.At(i)
		sb.Write(seg.Value(w.source))
	}
	return strings.TrimRight(sb.String(), "\n")
}

func (w *mdWalker) walkBlockquote(bq *ast.Blockquote) {
	w.bqDepth++
	depth := w.bqDepth
	var runs []Run
	var nested []*ast.Blockquote
	for c := bq.FirstChild(); c != nil; c = c.NextSibling() {
		if c.Kind() == ast.KindParagraph {
			cr := w.collectRuns(c, inlineState{})
			if len(runs) > 0 && len(cr) > 0 {
				runs = append(runs, Run{Text: "\n"})
			}
			runs = append(runs, cr...)
		} else if c.Kind() == ast.KindBlockquote {
			nested = append(nested, c.(*ast.Blockquote))
		} else {
			w.walkBlock(c)
		}
	}
	runs = trimTrailingBreaks(runs)
	if len(runs) > 0 {
		w.blocks = append(w.blocks, Block{
			IsBlockquote:    true,
			BlockquoteDepth: depth,
			Runs:            runs,
		})
	}
	// Recurse after emitting parent; bqDepth still at
	// current level so nested depths are correct.
	for _, nbq := range nested {
		w.walkBlockquote(nbq)
	}
	w.bqDepth--
}

func (w *mdWalker) walkList(list *ast.List) {
	startNum := list.Start
	for c := list.FirstChild(); c != nil; c = c.NextSibling() {
		if c.Kind() != ast.KindListItem {
			continue
		}
		li := c.(*ast.ListItem)

		// Determine prefix.
		var prefix string
		if list.IsOrdered() {
			prefix = fmt.Sprintf("%d. ", startNum)
			startNum++
		} else {
			prefix = "• "
		}

		// Check for task checkbox.
		if li.HasChildren() {
			p := li.FirstChild()
			if p.HasChildren() {
				fc := p.FirstChild()
				if fc.Kind() == east.KindTaskCheckBox {
					cb := fc.(*east.TaskCheckBox)
					if cb.IsChecked {
						prefix = "☑ "
					} else {
						prefix = "☐ "
					}
				}
			}
		}

		// Collect runs from inline children; defer nested
		// lists so the parent block is emitted first.
		var runs []Run
		var nested []*ast.List
		for ic := li.FirstChild(); ic != nil; ic = ic.NextSibling() {
			switch ic.Kind() {
			case ast.KindParagraph, ast.KindTextBlock:
				runs = append(runs,
					w.collectRuns(ic, inlineState{})...)
			case ast.KindList:
				nested = append(nested,
					ic.(*ast.List))
			default:
				w.walkBlock(ic)
			}
		}
		runs = trimTrailingBreaks(runs)
		w.blocks = append(w.blocks, Block{
			IsList:     true,
			ListPrefix: prefix,
			ListIndent: w.listDepth,
			Runs:       runs,
		})
		for _, nl := range nested {
			w.listDepth++
			w.walkList(nl)
			w.listDepth--
		}
	}
}

func (w *mdWalker) walkImage(img *ast.Image) {
	alt := w.collectText(img)
	src, width, height := parseImageSrc(
		string(img.Destination))
	if !isSafeImagePath(src) {
		src = ""
	}
	w.blocks = append(w.blocks, Block{
		IsImage:     true,
		ImageSrc:    src,
		ImageAlt:    alt,
		ImageWidth:  width,
		ImageHeight: height,
	})
}

func (w *mdWalker) walkTable(tbl *east.Table) {
	var aligns []Align
	for _, a := range tbl.Alignments {
		aligns = append(aligns, convertAlign(a))
	}
	var headers [][]Run
	var rows [][][]Run
	for c := tbl.FirstChild(); c != nil; c = c.NextSibling() {
		switch c.Kind() {
		case east.KindTableHeader:
			for cell := c.FirstChild(); cell != nil; cell = cell.NextSibling() {
				if cell.Kind() == east.KindTableCell {
					headers = append(headers,
						w.collectRuns(cell, inlineState{}))
				}
			}
		case east.KindTableRow:
			var row [][]Run
			for cell := c.FirstChild(); cell != nil; cell = cell.NextSibling() {
				if cell.Kind() == east.KindTableCell {
					row = append(row,
						w.collectRuns(cell, inlineState{}))
				}
			}
			rows = append(rows, row)
		}
	}
	colCount := len(headers)
	if colCount == 0 && len(rows) > 0 {
		colCount = len(rows[0])
	}
	w.blocks = append(w.blocks, Block{
		IsTable: true,
		TableData: &Table{
			Headers:    headers,
			Alignments: aligns,
			Rows:       rows,
			ColCount:   colCount,
		},
	})
}

func convertAlign(a east.Alignment) Align {
	switch a {
	case east.AlignLeft:
		return AlignLeft
	case east.AlignRight:
		return AlignRight
	case east.AlignCenter:
		return AlignCenter
	default:
		return AlignStart
	}
}

func (w *mdWalker) walkDefList(node ast.Node) {
	for c := node.FirstChild(); c != nil; c = c.NextSibling() {
		switch c.Kind() {
		case east.KindDefinitionTerm:
			runs := w.collectRuns(c,
				inlineState{format: FormatBold})
			runs = trimTrailingBreaks(runs)
			w.blocks = append(w.blocks, Block{
				IsDefTerm: true,
				Runs:      runs,
			})
		case east.KindDefinitionDescription:
			var runs []Run
			for dc := c.FirstChild(); dc != nil; dc = dc.NextSibling() {
				if dc.Kind() == ast.KindTextBlock ||
				dc.Kind() == ast.KindParagraph {
					runs = append(runs,
						w.collectRuns(dc, inlineState{})...)
				}
			}
			runs = trimTrailingBreaks(runs)
			w.blocks = append(w.blocks, Block{
				IsDefValue: true,
				Runs:       runs,
			})
		}
	}
}

// --- Inline walking ---

// collectRuns walks inline children producing flat Run slices.
func (w *mdWalker) collectRuns(
	node ast.Node, state inlineState,
) []Run {
	var runs []Run
	for c := node.FirstChild(); c != nil; c = c.NextSibling() {
		runs = append(runs, w.walkInline(c, state)...)
	}
	return runs
}

func (w *mdWalker) walkInline(
	node ast.Node, state inlineState,
) []Run {
	switch node.Kind() {
	case ast.KindText:
		return w.walkText(node.(*ast.Text), state)
	case ast.KindString:
		s := node.(*ast.String)
		v := string(s.Value)
		if len(v) == 0 {
			return nil
		}
		return []Run{w.makeRun(v, state)}
	case ast.KindEmphasis:
		return w.walkEmphasis(node.(*ast.Emphasis), state)
	case ast.KindCodeSpan:
		t := w.collectText(node)
		return []Run{{Text: t, Format: FormatCode,
			Link: state.link}}
	case ast.KindLink:
		return w.walkLink(node.(*ast.Link), state)
	case ast.KindAutoLink:
		return w.walkAutoLink(node.(*ast.AutoLink), state)
	case ast.KindImage:
		alt := w.collectText(node)
		return []Run{{Text: alt, Format: state.format}}
	case ast.KindRawHTML:
		return nil
	case emast.KindEmoji:
		e := node.(*emast.Emoji)
		if len(e.Value.Unicode) > 0 {
			return []Run{w.makeRun(
				string(e.Value.Unicode[0]), state)}
		}
		return []Run{w.makeRun(
			":"+string(e.ShortName)+":", state)}
	default:
		return w.walkInlineExt(node, state)
	}
}

func (w *mdWalker) walkText(
	t *ast.Text, state inlineState,
) []Run {
	seg := t.Segment
	v := string(seg.Value(w.source))
	var runs []Run
	if len(v) > 0 {
		runs = append(runs, w.makeRun(v, state))
	}
	if t.HardLineBreak() ||
		(w.hardBreaks && t.SoftLineBreak()) {
		runs = append(runs, Run{Text: "\n"})
	} else if t.SoftLineBreak() {
		runs = append(runs,
			Run{Text: " ", Format: state.format})
	}
	return runs
}

func (w *mdWalker) walkEmphasis(
	em *ast.Emphasis, state inlineState,
) []Run {
	ns := state
	if em.Level == 1 {
		ns.format = mergeFormat(ns.format, FormatItalic)
	} else {
		ns.format = mergeFormat(ns.format, FormatBold)
	}
	return w.collectRuns(em, ns)
}

func (w *mdWalker) walkLink(
	link *ast.Link, state inlineState,
) []Run {
	url := string(link.Destination)
	if !IsSafeURL(url) {
		url = ""
	}
	ns := state
	ns.link = url
	return w.collectRuns(link, ns)
}

func (w *mdWalker) walkAutoLink(
	al *ast.AutoLink, state inlineState,
) []Run {
	url := string(al.URL(w.source))
	label := string(al.Label(w.source))
	if !IsSafeURL(url) {
		return []Run{{Text: label, Format: state.format}}
	}
	return []Run{{Text: label, Format: state.format,
		Link: url}}
}

func (w *mdWalker) walkInlineExt(
	node ast.Node, state inlineState,
) []Run {
	kind := node.Kind()
	switch {
	case kind == east.KindStrikethrough:
		ns := state
		ns.strikethrough = true
		return w.collectRuns(node, ns)
	case kind == east.KindTaskCheckBox:
		return nil
	case kind == NodeKindMathInline:
		mi := node.(*nodeMathInline)
		id := fmt.Sprintf("math_%x", MathHash(mi.Latex))
		return []Run{{MathID: id, MathLatex: mi.Latex,
			Format: state.format}}
	case kind == NodeKindMathDisplay:
		md := node.(*nodeMathDisplay)
		id := fmt.Sprintf("math_%x", MathHash(md.Latex))
		return []Run{{MathID: id, MathLatex: md.Latex,
			Format: state.format}}
	case kind == NodeKindHighlight:
		ns := state
		ns.highlight = true
		return w.collectRuns(node, ns)
	case kind == NodeKindUnderline:
		ns := state
		ns.underline = true
		return w.collectRuns(node, ns)
	case kind == NodeKindSuperscript:
		ns := state
		ns.superscript = true
		return w.collectRuns(node, ns)
	case kind == NodeKindSubscript:
		ns := state
		ns.subscript = true
		return w.collectRuns(node, ns)
	default:
		return w.collectRuns(node, state)
	}
}

func (w *mdWalker) makeRun(
	text string, s inlineState,
) Run {
	return Run{
		Text:          text,
		Format:        s.format,
		Strikethrough: s.strikethrough,
		Highlight:     s.highlight,
		Underline:     s.underline,
		Superscript:   s.superscript,
		Subscript:     s.subscript,
		Link:          s.link,
		Tooltip:       s.tooltip,
	}
}

// collectText extracts plain text from a node tree.
func (w *mdWalker) collectText(node ast.Node) string {
	var sb strings.Builder
	_ = ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch n.Kind() {
		case ast.KindText:
			t := n.(*ast.Text)
			seg := t.Segment
			sb.Write(seg.Value(w.source))
		case ast.KindString:
			sb.Write(n.(*ast.String).Value)
		}
		return ast.WalkContinue, nil
	})
	return sb.String()
}

// mergeFormat combines formatting levels.
func mergeFormat(parent, child Format) Format {
	if parent == FormatCode || child == FormatCode {
		return FormatCode
	}
	switch parent {
	case FormatPlain:
		return child
	case FormatBold:
		if child == FormatItalic ||
			child == FormatBoldItalic {
			return FormatBoldItalic
		}
		return parent
	case FormatItalic:
		if child == FormatBold ||
			child == FormatBoldItalic {
			return FormatBoldItalic
		}
		return parent
	case FormatBoldItalic:
		return FormatBoldItalic
	}
	return parent
}

// --- Pre-processing ---

// reImageDims matches image links with dimension syntax:
// ](url =WxH) → ](url#dim=WxH)
// Goldmark splits on the space, so encode dims as a fragment.
var reImageDims = regexp.MustCompile(
	`(\]\([^\s)]+) (=\d+x\d+\))`)

// preprocessSource strips metadata definitions, converts
// multi-line $$...$$ to ```math code fences, and encodes
// image dimension syntax into URL fragments.
func preprocessSource(source string) string {
	source = reImageDims.ReplaceAllString(
		source, "${1}#dim${2}")
	lines := strings.Split(source, "\n")
	result := make([]string, 0, len(lines))
	i := 0
	for i < len(lines) {
		trimmed := strings.TrimSpace(lines[i])

		// Strip abbreviation definitions.
		if isAbbrDef(trimmed) {
			result = append(result, "")
			i++
			continue
		}

		// Strip footnote definitions + continuations.
		if isFootnoteDef(trimmed) {
			result = append(result, "")
			i++
			for i < len(lines) {
				next := lines[i]
				if len(next) == 0 {
					if i+1 < len(lines) &&
						len(lines[i+1]) > 0 &&
						(lines[i+1][0] == ' ' ||
							lines[i+1][0] == '\t') {
						result = append(result, "")
						i++
						continue
					}
					break
				}
				if next[0] != ' ' && next[0] != '\t' {
					break
				}
				result = append(result, "")
				i++
			}
			continue
		}

		// Multi-line $$...$$ → ```math fences.
		if trimmed == "$$" {
			result = append(result, "```math")
			i++
			for i < len(lines) {
				if strings.TrimSpace(lines[i]) == "$$" {
					result = append(result, "```")
					i++
					break
				}
				result = append(result, lines[i])
				i++
			}
			continue
		}

		result = append(result, lines[i])
		i++
	}
	return strings.Join(result, "\n")
}

// --- Metadata pre-scanning ---

func collectAbbrDefs(source string) map[string]string {
	defs := map[string]string{}
	for _, line := range strings.Split(source, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "*[") {
			continue
		}
		idx := strings.Index(trimmed, "]:")
		if idx < 3 {
			continue
		}
		abbr := trimmed[2:idx]
		expansion := strings.TrimSpace(trimmed[idx+2:])
		if len(abbr) > 0 && len(expansion) > 0 {
			defs[abbr] = expansion
			if len(defs) >= maxAbbreviationDefs {
				break
			}
		}
	}
	return defs
}

func collectFootnoteDefs(source string) map[string]string {
	defs := map[string]string{}
	lines := strings.Split(source, "\n")
	i := 0
	for i < len(lines) {
		trimmed := strings.TrimSpace(lines[i])
		if !strings.HasPrefix(trimmed, "[^") {
			i++
			continue
		}
		idx := strings.Index(trimmed, "]:")
		if idx < 3 {
			i++
			continue
		}
		id := trimmed[2:idx]
		content := strings.TrimSpace(trimmed[idx+2:])
		i++
		contCount := 0
		for i < len(lines) &&
			contCount < maxFootnoteContinuationLines {
			next := lines[i]
			if len(next) == 0 {
				if i+1 < len(lines) &&
					len(lines[i+1]) > 0 &&
					(lines[i+1][0] == ' ' ||
						lines[i+1][0] == '\t') {
					content += "\n\n"
					i++
					continue
				}
				break
			}
			if next[0] != ' ' && next[0] != '\t' {
				break
			}
			content += " " + strings.TrimSpace(next)
			contCount++
			i++
		}
		if len(id) > 0 && len(content) > 0 {
			defs[id] = content
			if len(defs) >= maxFootnoteDefs {
				break
			}
		}
	}
	return defs
}

func isAbbrDef(line string) bool {
	return strings.HasPrefix(line, "*[") &&
		strings.Contains(line, "]:")
}

func isFootnoteDef(line string) bool {
	if !strings.HasPrefix(line, "[^") {
		return false
	}
	return strings.Index(line, "]:") >= 3
}

// --- Post-processing ---

// applyFootnoteRefs replaces [^id] patterns in run text
// with superscript tooltip runs.
func applyFootnoteRefs(
	runs []Run, defs map[string]string,
) []Run {
	if len(defs) == 0 {
		return runs
	}
	var result []Run
	changed := false
	for _, run := range runs {
		if run.Link != "" || run.Tooltip != "" ||
			run.MathID != "" {
			result = append(result, run)
			continue
		}
		nr := splitRunForFootnotes(run, defs)
		if len(nr) != 1 || nr[0].Tooltip != "" {
			changed = true
		}
		result = append(result, nr...)
	}
	if !changed {
		return runs
	}
	return result
}

func splitRunForFootnotes(
	run Run, defs map[string]string,
) []Run {
	t := run.Text
	var result []Run
	pos := 0
	lastPos := 0
	for pos < len(t) {
		idx := strings.Index(t[pos:], "[^")
		if idx < 0 {
			break
		}
		start := pos + idx
		end := strings.Index(t[start+2:], "]")
		if end < 0 {
			pos = start + 2
			continue
		}
		end = start + 2 + end
		id := t[start+2 : end]
		content, ok := defs[id]
		if !ok {
			pos = end + 1
			continue
		}
		if start > lastPos {
			r := run
			r.Text = t[lastPos:start]
			result = append(result, r)
		}
		result = append(result, Run{
			Text:        id,
			Format:      run.Format,
			Superscript: true,
			Tooltip:     content,
		})
		lastPos = end + 1
		pos = lastPos
	}
	if lastPos == 0 {
		return []Run{run}
	}
	if lastPos < len(t) {
		r := run
		r.Text = t[lastPos:]
		result = append(result, r)
	}
	return result
}

// replaceAbbreviations scans runs for abbreviation occurrences
// and splits/marks them with tooltips.
func replaceAbbreviations(
	runs []Run, defs map[string]string,
) []Run {
	if len(defs) == 0 {
		return runs
	}
	abbrs := make([]string, 0, len(defs))
	for k := range defs {
		abbrs = append(abbrs, k)
	}
	sort.Slice(abbrs, func(i, j int) bool {
		return len(abbrs[i]) > len(abbrs[j])
	})
	var result []Run
	for _, run := range runs {
		if run.Link != "" || run.Tooltip != "" ||
			run.MathID != "" {
			result = append(result, run)
			continue
		}
		result = append(result,
			splitRunForAbbrs(run, abbrs, defs)...)
	}
	return result
}

func splitRunForAbbrs(
	run Run, abbrs []string, defs map[string]string,
) []Run {
	t := run.Text
	if len(t) == 0 {
		return []Run{run}
	}
	var firstChars [256]bool
	for _, a := range abbrs {
		if len(a) > 0 {
			firstChars[a[0]] = true
		}
	}
	var result []Run
	pos := 0
	lastPos := 0
	for pos < len(t) {
		if !firstChars[t[pos]] {
			pos++
			continue
		}
		matched := false
		for _, abbr := range abbrs {
			if pos+len(abbr) > len(t) ||
				t[pos:pos+len(abbr)] != abbr {
				continue
			}
			if !isWordBoundary(t, pos-1) ||
				!isWordBoundary(t, pos+len(abbr)) {
				continue
			}
			if pos > lastPos {
				r := run
				r.Text = t[lastPos:pos]
				result = append(result, r)
			}
			result = append(result, Run{
				Text:          abbr,
				Format:        run.Format,
				Strikethrough: run.Strikethrough,
				Highlight:     run.Highlight,
				Superscript:   run.Superscript,
				Subscript:     run.Subscript,
				Tooltip:       defs[abbr],
			})
			pos += len(abbr)
			lastPos = pos
			matched = true
			break
		}
		if !matched {
			pos++
		}
	}
	if lastPos == 0 {
		return []Run{run}
	}
	if lastPos < len(t) {
		r := run
		r.Text = t[lastPos:]
		result = append(result, r)
	}
	return result
}

func isWordBoundary(text string, pos int) bool {
	if pos < 0 || pos >= len(text) {
		return true
	}
	c := text[pos]
	return !((c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') || c == '_')
}

// --- Helpers ---

// mergeAdjacentRuns combines consecutive runs with identical
// formatting into single runs. Needed because goldmark may
// split text across multiple Text nodes (e.g. [^1] becomes
// "[" and "^1]" separately).
func mergeAdjacentRuns(runs []Run) []Run {
	if len(runs) <= 1 {
		return runs
	}
	result := make([]Run, 0, len(runs))
	cur := runs[0]
	for _, r := range runs[1:] {
		if canMergeRuns(cur, r) {
			cur.Text += r.Text
		} else {
			result = append(result, cur)
			cur = r
		}
	}
	result = append(result, cur)
	return result
}

func canMergeRuns(a, b Run) bool {
	return a.Format == b.Format &&
		a.Strikethrough == b.Strikethrough &&
		a.Highlight == b.Highlight &&
		a.Superscript == b.Superscript &&
		a.Subscript == b.Subscript &&
		a.Link == b.Link &&
		a.Tooltip == b.Tooltip &&
		a.MathID == "" && b.MathID == "" &&
		a.CodeToken == b.CodeToken &&
		a.Underline == b.Underline
}

// applyTypography replaces --- with em dash, -- with en dash,
// and ... with ellipsis in non-code runs.
// Must replace --- before --.
func applyTypography(runs []Run) {
	for i := range runs {
		if runs[i].Format == FormatCode || runs[i].MathID != "" {
			continue
		}
		t := runs[i].Text
		t = strings.ReplaceAll(t, "---", "\u2014")
		t = strings.ReplaceAll(t, "--", "\u2013")
		t = strings.ReplaceAll(t, "...", "\u2026")
		runs[i].Text = t
	}
}

func trimTrailingBreaks(runs []Run) []Run {
	for len(runs) > 0 &&
		runs[len(runs)-1].Text == "\n" &&
		runs[len(runs)-1].Link == "" {
		runs = runs[:len(runs)-1]
	}
	return runs
}

// RunsToText concatenates run text into a single string.
func RunsToText(runs []Run) string {
	var sb strings.Builder
	for _, r := range runs {
		sb.WriteString(r.Text)
	}
	return sb.String()
}

// parseImageSrc parses dimension suffixes from image URLs.
// Handles both "path.png =WxH" (original) and
// "path.png#dim=WxH" (preprocessed for goldmark).
func parseImageSrc(raw string) (string, float32, float32) {
	raw = strings.TrimSpace(raw)
	// Check preprocessed fragment form first.
	if idx := strings.LastIndex(raw, "#dim="); idx >= 0 {
		src := raw[:idx]
		dims := raw[idx+5:]
		if w, h, ok := parseDims(dims); ok {
			return src, w, h
		}
	}
	// Original space-separated form.
	if idx := strings.LastIndex(raw, " ="); idx >= 0 {
		dims := raw[idx+2:]
		if w, h, ok := parseDims(dims); ok {
			return strings.TrimSpace(raw[:idx]), w, h
		}
	}
	return raw, 0, 0
}

func parseDims(s string) (float32, float32, bool) {
	xIdx := strings.IndexByte(s, 'x')
	if xIdx < 0 {
		return 0, 0, false
	}
	w := parseFloat32(s[:xIdx])
	h := parseFloat32(s[xIdx+1:])
	return w, h, w > 0 || h > 0
}

func parseFloat32(s string) float32 {
	var v float32
	for _, c := range s {
		if c >= '0' && c <= '9' {
			v = v*10 + float32(c-'0')
		} else {
			return 0
		}
	}
	return v
}

// MathHash computes a simple hash of a string.
func MathHash(s string) uint64 {
	var h uint64
	for i := 0; i < len(s); i++ {
		h = h*31 + uint64(s[i])
	}
	return h
}
