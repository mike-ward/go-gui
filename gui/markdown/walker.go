package markdown

// walker.go converts markdown source to []Block using
// goldmark as the parser backend.

import (
	"cmp"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"sync"

	"github.com/yuin/goldmark"
	emoji "github.com/yuin/goldmark-emoji"
	emast "github.com/yuin/goldmark-emoji/ast"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	east "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

var (
	mdInstance     goldmark.Markdown
	mdInstanceOnce sync.Once
)

func getMarkdownParser() goldmark.Markdown {
	mdInstanceOnce.Do(func() {
		mdInstance = goldmark.New(
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
	})
	return mdInstance
}

// Parse converts markdown source to []Block.
func Parse(source string, hardLineBreaks bool) []Block {
	// Normalize CRLF/CR to LF. The WASM text backend's
	// Intl.Segmenter treats \r\n as a single grapheme cluster,
	// preventing newline recognition in code blocks.
	if strings.ContainsRune(source, '\r') {
		source = strings.ReplaceAll(source, "\r\n", "\n")
		source = strings.ReplaceAll(source, "\r", "\n")
	}
	source, abbrDefs, footnoteDefs := scanSource(source)
	abbrMatcher := buildAbbrMatcher(abbrDefs)
	src := []byte(source)

	md := getMarkdownParser()
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
				w.blocks[i].Runs, abbrMatcher)
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

type abbrMatcher struct {
	abbrs      []string
	firstChars [256]bool
	defs       map[string]string
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
	for i := range segs.Len() {
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
	runs := make([]Run, 0, node.ChildCount())
	return w.collectRunsInto(runs, node, state)
}

func (w *mdWalker) collectRunsInto(
	dst []Run, node ast.Node, state inlineState,
) []Run {
	for c := node.FirstChild(); c != nil; c = c.NextSibling() {
		dst = w.walkInline(dst, c, state)
	}
	return dst
}

func (w *mdWalker) walkInline(
	dst []Run, node ast.Node, state inlineState,
) []Run {
	switch node.Kind() {
	case ast.KindText:
		return w.walkText(dst, node.(*ast.Text), state)
	case ast.KindString:
		s := node.(*ast.String)
		v := string(s.Value)
		if len(v) == 0 {
			return dst
		}
		return append(dst, w.makeRun(v, state))
	case ast.KindEmphasis:
		return w.walkEmphasis(dst, node.(*ast.Emphasis), state)
	case ast.KindCodeSpan:
		t := w.collectText(node)
		return append(dst, Run{
			Text: t, Format: FormatCode, Link: state.link,
		})
	case ast.KindLink:
		return w.walkLink(dst, node.(*ast.Link), state)
	case ast.KindAutoLink:
		return w.walkAutoLink(dst, node.(*ast.AutoLink), state)
	case ast.KindImage:
		alt := w.collectText(node)
		return append(dst, Run{Text: alt, Format: state.format})
	case ast.KindRawHTML:
		return dst
	case emast.KindEmoji:
		e := node.(*emast.Emoji)
		if len(e.Value.Unicode) > 0 {
			return append(dst, w.makeRun(
				string(e.Value.Unicode), state))
		}
		return append(dst, w.makeRun(
			":"+string(e.ShortName)+":", state))
	default:
		return w.walkInlineExt(dst, node, state)
	}
}

func (w *mdWalker) walkText(
	dst []Run, t *ast.Text, state inlineState,
) []Run {
	seg := t.Segment
	v := string(seg.Value(w.source))
	if len(v) > 0 {
		dst = append(dst, w.makeRun(v, state))
	}
	if t.HardLineBreak() ||
		(w.hardBreaks && t.SoftLineBreak()) {
		dst = append(dst, Run{Text: "\n"})
	} else if t.SoftLineBreak() {
		dst = append(dst, Run{Text: " ", Format: state.format})
	}
	return dst
}

func (w *mdWalker) walkEmphasis(
	dst []Run, em *ast.Emphasis, state inlineState,
) []Run {
	ns := state
	if em.Level == 1 {
		ns.format = mergeFormat(ns.format, FormatItalic)
	} else {
		ns.format = mergeFormat(ns.format, FormatBold)
	}
	return w.collectRunsInto(dst, em, ns)
}

func (w *mdWalker) walkLink(
	dst []Run, link *ast.Link, state inlineState,
) []Run {
	url := string(link.Destination)
	if !IsSafeURL(url) {
		url = ""
	}
	ns := state
	ns.link = url
	return w.collectRunsInto(dst, link, ns)
}

func (w *mdWalker) walkAutoLink(
	dst []Run, al *ast.AutoLink, state inlineState,
) []Run {
	url := string(al.URL(w.source))
	label := string(al.Label(w.source))
	if !IsSafeURL(url) {
		return append(dst, Run{
			Text: label, Format: state.format,
		})
	}
	return append(dst, Run{
		Text: label, Format: state.format, Link: url,
	})
}

func (w *mdWalker) walkInlineExt(
	dst []Run, node ast.Node, state inlineState,
) []Run {
	switch kind := node.Kind(); kind {
	case east.KindStrikethrough:
		ns := state
		ns.strikethrough = true
		return w.collectRunsInto(dst, node, ns)
	case east.KindTaskCheckBox:
		return dst
	case NodeKindMathInline:
		mi := node.(*nodeMathInline)
		id := fmt.Sprintf("math_%x", MathHash(mi.Latex))
		return append(dst, Run{
			MathID: id, MathLatex: mi.Latex, Format: state.format,
		})
	case NodeKindMathDisplay:
		md := node.(*nodeMathDisplay)
		id := fmt.Sprintf("math_%x", MathHash(md.Latex))
		return append(dst, Run{
			MathID: id, MathLatex: md.Latex, Format: state.format,
		})
	case NodeKindHighlight:
		ns := state
		ns.highlight = true
		return w.collectRunsInto(dst, node, ns)
	case NodeKindUnderline:
		ns := state
		ns.underline = true
		return w.collectRunsInto(dst, node, ns)
	case NodeKindSuperscript:
		ns := state
		ns.superscript = true
		return w.collectRunsInto(dst, node, ns)
	case NodeKindSubscript:
		ns := state
		ns.subscript = true
		return w.collectRunsInto(dst, node, ns)
	default:
		return w.collectRunsInto(dst, node, state)
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
	if node == nil {
		return ""
	}
	var sb strings.Builder
	var local [16]ast.Node
	stack := local[:0]
	for c := node.LastChild(); c != nil; c = c.PreviousSibling() {
		stack = append(stack, c)
	}
	for len(stack) > 0 {
		n := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		switch n.Kind() {
		case ast.KindText:
			t := n.(*ast.Text)
			seg := t.Segment
			sb.Write(seg.Value(w.source))
		case ast.KindString:
			sb.Write(n.(*ast.String).Value)
		}
		for c := n.LastChild(); c != nil; c = c.PreviousSibling() {
			stack = append(stack, c)
		}
	}
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

func scanSource(source string) (string, map[string]string, map[string]string) {
	source = reImageDims.ReplaceAllString(source, "${1}#dim${2}")
	lines := strings.Split(source, "\n")
	abbrDefs := map[string]string{}
	footnoteDefs := map[string]string{}
	result := make([]string, 0, len(lines))
	i := 0
	for i < len(lines) {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		if isAbbrDef(trimmed) {
			if len(abbrDefs) < maxAbbreviationDefs {
				idx := strings.Index(trimmed, "]:")
				if idx >= 3 {
					abbr := trimmed[2:idx]
					expansion := strings.TrimSpace(trimmed[idx+2:])
					if len(abbr) > 0 && len(expansion) > 0 {
						abbrDefs[abbr] = expansion
					}
				}
			}
			result = append(result, "")
			i++
			continue
		}

		if isFootnoteDef(trimmed) {
			var id, content string
			idx := strings.Index(trimmed, "]:")
			if idx >= 3 {
				id = trimmed[2:idx]
				content = strings.TrimSpace(trimmed[idx+2:])
			}
			result = append(result, "")
			i++
			contCount := 0
			for i < len(lines) {
				next := lines[i]
				if len(next) == 0 {
					if i+1 < len(lines) &&
						len(lines[i+1]) > 0 &&
						(lines[i+1][0] == ' ' ||
							lines[i+1][0] == '\t') {
						if contCount < maxFootnoteContinuationLines {
							content += "\n\n"
						}
						result = append(result, "")
						i++
						continue
					}
					break
				}
				if next[0] != ' ' && next[0] != '\t' {
					break
				}
				if contCount < maxFootnoteContinuationLines {
					content += " " + strings.TrimSpace(next)
					contCount++
				}
				result = append(result, "")
				i++
			}
			if len(footnoteDefs) < maxFootnoteDefs &&
				len(id) > 0 && len(content) > 0 {
				footnoteDefs[id] = content
			}
			continue
		}

		// Multi-line $$...$$ → ```math fences.
		if trimmed == "$$" {
			result = append(result, "```math")
			i++
			const maxMathLines = 200
			mathLines := 0
			for i < len(lines) && mathLines < maxMathLines {
				if strings.TrimSpace(lines[i]) == "$$" {
					result = append(result, "```")
					i++
					break
				}
				result = append(result, lines[i])
				i++
				mathLines++
			}
			continue
		}

		result = append(result, line)
		i++
	}
	return strings.Join(result, "\n"), abbrDefs, footnoteDefs
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
	match := footnoteMatchFunc(defs)
	result := make([]Run, 0, len(runs))
	changed := false
	for _, run := range runs {
		if run.Link != "" || run.Tooltip != "" ||
			run.MathID != "" {
			result = append(result, run)
			continue
		}
		nr, split := splitRunByMatches(run, match)
		if split {
			changed = true
			result = append(result, nr...)
			continue
		}
		result = append(result, run)
	}
	if !changed {
		return runs
	}
	return result
}

// runMatchFunc finds the next match at or after pos in text.
// Returns start/end indices, the replacement run, and whether
// a match was found.
type runMatchFunc func(
	text string, pos int, base Run,
) (start, end int, replacement Run, found bool)

// splitRunByMatches splits a run using match to find and
// replace substrings. Returns nil, false if no matches.
func splitRunByMatches(
	run Run, match runMatchFunc,
) ([]Run, bool) {
	t := run.Text
	if len(t) == 0 {
		return nil, false
	}
	var result []Run
	pos := 0
	lastPos := 0
	split := false
	for pos < len(t) {
		start, end, repl, found := match(t, pos, run)
		if !found {
			break
		}
		if !split {
			result = make([]Run, 0, 3)
			split = true
		}
		if start > lastPos {
			r := run
			r.Text = t[lastPos:start]
			result = append(result, r)
		}
		result = append(result, repl)
		lastPos = end
		pos = end
	}
	if !split {
		return nil, false
	}
	if lastPos < len(t) {
		r := run
		r.Text = t[lastPos:]
		result = append(result, r)
	}
	return result, true
}

func footnoteMatchFunc(
	defs map[string]string,
) runMatchFunc {
	return func(
		text string, pos int, base Run,
	) (int, int, Run, bool) {
		for pos < len(text) {
			idx := strings.Index(text[pos:], "[^")
			if idx < 0 {
				return 0, 0, Run{}, false
			}
			start := pos + idx
			end := strings.Index(text[start+2:], "]")
			if end < 0 {
				pos = start + 2
				continue
			}
			end = start + 2 + end
			id := text[start+2 : end]
			content, ok := defs[id]
			if !ok {
				pos = end + 1
				continue
			}
			return start, end + 1, Run{
				Text:        id,
				Format:      base.Format,
				Superscript: true,
				Tooltip:     content,
			}, true
		}
		return 0, 0, Run{}, false
	}
}

func abbrMatchFunc(matcher *abbrMatcher) runMatchFunc {
	return func(
		text string, pos int, base Run,
	) (int, int, Run, bool) {
		for pos < len(text) {
			if !matcher.firstChars[text[pos]] {
				pos++
				continue
			}
			for _, abbr := range matcher.abbrs {
				if pos+len(abbr) > len(text) ||
					text[pos:pos+len(abbr)] != abbr {
					continue
				}
				if !isWordBoundary(text, pos-1) ||
					!isWordBoundary(text, pos+len(abbr)) {
					continue
				}
				return pos, pos + len(abbr), Run{
					Text:          abbr,
					Format:        base.Format,
					Strikethrough: base.Strikethrough,
					Highlight:     base.Highlight,
					Superscript:   base.Superscript,
					Subscript:     base.Subscript,
					Tooltip:       matcher.defs[abbr],
				}, true
			}
			pos++
		}
		return 0, 0, Run{}, false
	}
}

// replaceAbbreviations scans runs for abbreviation occurrences
// and splits/marks them with tooltips.
func replaceAbbreviations(
	runs []Run, matcher *abbrMatcher,
) []Run {
	if matcher == nil {
		return runs
	}
	match := abbrMatchFunc(matcher)
	result := make([]Run, 0, len(runs))
	for _, run := range runs {
		if run.Link != "" || run.Tooltip != "" ||
			run.MathID != "" {
			result = append(result, run)
			continue
		}
		nr, split := splitRunByMatches(run, match)
		if split {
			result = append(result, nr...)
			continue
		}
		result = append(result, run)
	}
	return result
}

func buildAbbrMatcher(defs map[string]string) *abbrMatcher {
	if len(defs) == 0 {
		return nil
	}
	abbrs := make([]string, 0, len(defs))
	var firstChars [256]bool
	for k := range defs {
		if len(k) == 0 {
			continue
		}
		abbrs = append(abbrs, k)
		firstChars[k[0]] = true
	}
	slices.SortFunc(abbrs, func(a, b string) int {
		return cmp.Compare(len(b), len(a))
	})
	return &abbrMatcher{
		abbrs:      abbrs,
		firstChars: firstChars,
		defs:       defs,
	}
}

func isWordBoundary(text string, pos int) bool {
	if pos < 0 || pos >= len(text) {
		return true
	}
	c := text[pos]
	return (c < 'a' || c > 'z') &&
		(c < 'A' || c > 'Z') &&
		(c < '0' || c > '9') && c != '_'
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
	var sb strings.Builder
	merging := false
	for _, r := range runs[1:] {
		if canMergeRuns(cur, r) {
			if !merging {
				sb.WriteString(cur.Text)
				merging = true
			}
			sb.WriteString(r.Text)
		} else {
			if merging {
				cur.Text = sb.String()
				sb.Reset()
				merging = false
			}
			result = append(result, cur)
			cur = r
		}
	}
	if merging {
		cur.Text = sb.String()
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
		if !strings.Contains(t, "--") && !strings.Contains(t, "...") {
			continue
		}
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
	before, after, found := strings.Cut(s, "x")
	if !found {
		return 0, 0, false
	}
	w := parseFloat32(before)
	h := parseFloat32(after)
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

// MathHash computes a FNV-1a hash of a string.
func MathHash(s string) uint64 {
	h := uint64(14695981039346656037) // FNV offset basis
	for i := range len(s) {
		h ^= uint64(s[i])
		h *= 1099511628211 // FNV prime
	}
	return h
}
