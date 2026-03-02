package gui

// view_markdown.go defines the Markdown view component.
// Parses markdown source and renders it using RTF views.

import (
	"fmt"
	"log"
	"strings"
)

// MarkdownStyle controls rendered markdown appearance.
type MarkdownStyle struct {
	Text              TextStyle
	H1                TextStyle
	H2                TextStyle
	H3                TextStyle
	H4                TextStyle
	H5                TextStyle
	H6                TextStyle
	Bold              TextStyle
	Italic            TextStyle
	BoldItalic        TextStyle
	Code              TextStyle
	CodeBlockBG       Color
	CodeKeywordColor  Color
	CodeStringColor   Color
	CodeNumberColor   Color
	CodeCommentColor  Color
	CodeOperatorColor Color
	HRColor           Color
	LinkColor         Color
	BlockquoteBorder  Color
	BlockquoteBG      Color
	BlockSpacing      float32
	NestIndent        float32
	PrefixCharWidth   float32
	CodeBlockPadding  Padding
	CodeBlockRadius   float32
	H1Separator       bool
	H2Separator       bool
	TableBorderStyle  TableBorderStyle
	TableBorderColor  Color
	TableBorderSize   float32
	TableHeadStyle    TextStyle
	TableCellStyle    TextStyle
	TableCellPadding  Padding
	TableRowAlt       *Color
	HighlightBG       Color
	HardLineBreaks    bool
	MathDPIDisplay    int
	MathDPIInline     int
	MermaidBG         Color
}

// DefaultMarkdownStyle returns a MarkdownStyle using the
// current theme.
func DefaultMarkdownStyle() MarkdownStyle {
	return MarkdownStyle{
		Text:              guiTheme.N3,
		H1:                guiTheme.B1,
		H2:                guiTheme.B2,
		H3:                guiTheme.B3,
		H4:                guiTheme.B4,
		H5:                guiTheme.B5,
		H6:                guiTheme.B6,
		Bold:              guiTheme.B3,
		Italic:            guiTheme.I3,
		BoldItalic:        guiTheme.BI3,
		Code:              guiTheme.M5,
		CodeBlockBG:       RGBA(0, 0, 0, 50),
		CodeKeywordColor:  guiTheme.ColorSelect,
		CodeStringColor:   RGB(75, 125, 75),
		CodeNumberColor:   RGB(169, 114, 62),
		CodeCommentColor:  guiTheme.ColorBorder,
		CodeOperatorColor: guiTheme.N3.Color,
		HRColor:           guiTheme.ColorBorder,
		LinkColor:         guiTheme.ColorSelect,
		BlockquoteBorder:  guiTheme.ColorBorder,
		BlockquoteBG:      RGBA(128, 128, 128, 20),
		BlockSpacing:      8,
		NestIndent:        16,
		PrefixCharWidth:   8,
		CodeBlockPadding:  PadAll(10),
		CodeBlockRadius:   3.5,
		TableBorderStyle:  TableBorderHeaderOnly,
		TableBorderColor:  guiTheme.ColorBorder,
		TableBorderSize:   1,
		TableHeadStyle:    guiTheme.B3,
		TableCellStyle:    guiTheme.N3,
		TableCellPadding:  NewPadding(5, 10, 5, 10),
		HighlightBG:       RGB(199, 142, 18),
		MathDPIDisplay:    150,
		MathDPIInline:     200,
		MermaidBG:         RGBA(248, 248, 255, 255),
	}
}

// MarkdownCfg configures a Markdown View.
type MarkdownCfg struct {
	ID                  string
	Source              string
	Style               MarkdownStyle
	IDFocus             uint32
	Mode                TextMode
	MinWidth            float32
	Invisible           bool
	Clip                bool
	FocusSkip           bool
	Disabled            bool
	Color               Color
	ColorBorder         Color
	SizeBorder          float32
	Radius              float32
	Padding             Padding
	MermaidWidth        int
	DisableExternalAPIs bool
}

func richTextPlain(rt RichText) string {
	if len(rt.Runs) == 0 {
		return ""
	}
	var sb strings.Builder
	for _, r := range rt.Runs {
		sb.WriteString(r.Text)
	}
	return sb.String()
}

func nextDiagramRequestID(w *Window) uint64 {
	w.viewState.diagramRequestSeq++
	return w.viewState.diagramRequestSeq
}

func markdownWarnExternalAPIOnce(w *Window) {
	if w.viewState.externalAPIWarningLogged {
		return
	}
	w.viewState.externalAPIWarningLogged = true
	log.Println("markdown: external APIs enabled; " +
		"content may be sent to codecogs.com and kroki.io")
}

// buildMarkdownTableData converts ParsedTable to
// TableRowCfg array.
func buildMarkdownTableData(
	parsed ParsedTable, style MarkdownStyle,
) []TableRowCfg {
	rows := make([]TableRowCfg, 0, len(parsed.Rows)+1)

	// Header row.
	hCells := make([]TableCellCfg, 0, len(parsed.Headers))
	for _, h := range parsed.Headers {
		hCells = append(hCells, TableCellCfg{
			Value:    richTextPlain(h),
			HeadCell: true,
			Content: RTF(RtfCfg{
				RichText: h,
				Mode:     TextModeSingleLine,
			}),
		})
	}
	rows = append(rows, TableRowCfg{Cells: hCells})

	// Data rows.
	for _, r := range parsed.Rows {
		cells := make([]TableCellCfg, 0, len(r))
		for _, cell := range r {
			cells = append(cells, TableCellCfg{
				Value: richTextPlain(cell),
				Content: RTF(RtfCfg{
					RichText: cell,
					Mode:     TextModeSingleLine,
				}),
			})
		}
		rows = append(rows, TableRowCfg{Cells: cells})
	}
	return rows
}

// renderMdMath renders a display math block.
func renderMdMath(
	block MarkdownBlock, cfg MarkdownCfg, w *Window,
) View {
	codeFallback := Column(ContainerCfg{
		Color:      cfg.Style.CodeBlockBG,
		Padding:    cfg.Style.CodeBlockPadding,
		Radius:     Some(cfg.Style.CodeBlockRadius),
		Sizing:     FillFit,
		Content: []View{
			Text(TextCfg{
				Text:      block.MathLatex,
				TextStyle: cfg.Style.Code,
			}),
		},
	})

	if cfg.DisableExternalAPIs {
		return codeFallback
	}

	diagramHash := mathCacheHash(
		fmt.Sprintf("display_%d", mathHash(block.MathLatex)))

	if w.viewState.diagramCache != nil {
		if entry, ok := w.viewState.diagramCache.Get(
			diagramHash); ok {
			switch entry.State {
			case DiagramLoading:
				return codeFallback
			case DiagramReady:
				// Placeholder — image view not yet available.
				return Text(TextCfg{
					Text:      "[math: rendered]",
					TextStyle: cfg.Style.Code,
				})
			case DiagramError:
				errStyle := cfg.Style.Code
				errStyle.Color = RGBA(200, 50, 50, 255)
				return Text(TextCfg{
					Text:      entry.Error,
					TextStyle: errStyle,
				})
			}
		}
	}

	// Start async fetch.
	if w.viewState.diagramCache == nil {
		w.viewState.diagramCache =
			NewBoundedDiagramCache(50)
	}
	if w.viewState.diagramCache.LoadingCount() <
		maxConcurrentDiagramFetches {
		reqID := nextDiagramRequestID(w)
		w.viewState.diagramCache.Set(diagramHash,
			DiagramCacheEntry{
				State:     DiagramLoading,
				RequestID: reqID,
			})
		fetchMathAsync(w, block.MathLatex, diagramHash,
			reqID, cfg.Style.MathDPIDisplay,
			cfg.Style.Text.Color)
	}
	return codeFallback
}

// renderMdMermaid renders a mermaid diagram block.
func renderMdMermaid(
	block MarkdownBlock, cfg MarkdownCfg, w *Window,
) View {
	source := richTextPlain(block.Content)
	codeFallback := Column(ContainerCfg{
		Color:   cfg.Style.CodeBlockBG,
		Padding: cfg.Style.CodeBlockPadding,
		Radius:  Some(cfg.Style.CodeBlockRadius),
		Sizing:  FillFit,
		Content: []View{
			RTF(RtfCfg{
				RichText: block.Content,
				Mode:     TextModeSingleLine,
			}),
		},
	})

	if cfg.DisableExternalAPIs {
		return codeFallback
	}

	diagramHash := int64(
		(mathHash(source) << 32) | uint64(len(source)))

	if w.viewState.diagramCache != nil {
		if entry, ok := w.viewState.diagramCache.Get(
			diagramHash); ok {
			switch entry.State {
			case DiagramLoading:
				return Text(TextCfg{
					Text:      "Loading diagram...",
					TextStyle: cfg.Style.Text,
				})
			case DiagramReady:
				return Text(TextCfg{
					Text:      "[diagram: rendered]",
					TextStyle: cfg.Style.Code,
				})
			case DiagramError:
				errStyle := cfg.Style.Code
				errStyle.Color = RGBA(200, 50, 50, 255)
				return Text(TextCfg{
					Text:      entry.Error,
					TextStyle: errStyle,
				})
			}
		}
	}

	if w.viewState.diagramCache == nil {
		w.viewState.diagramCache =
			NewBoundedDiagramCache(50)
	}
	if w.viewState.diagramCache.LoadingCount() <
		maxConcurrentDiagramFetches {
		reqID := nextDiagramRequestID(w)
		w.viewState.diagramCache.Set(diagramHash,
			DiagramCacheEntry{
				State:     DiagramLoading,
				RequestID: reqID,
			})
		mw := cfg.MermaidWidth
		if mw <= 0 {
			mw = 500
		}
		fetchMermaidAsync(w, source, diagramHash, reqID,
			cfg.Style.MermaidBG.R,
			cfg.Style.MermaidBG.G,
			cfg.Style.MermaidBG.B)
	}
	return codeFallback
}

// renderMdCode renders a fenced code block.
func renderMdCode(
	block MarkdownBlock, cfg MarkdownCfg,
) View {
	return Column(ContainerCfg{
		Color:   cfg.Style.CodeBlockBG,
		Padding: cfg.Style.CodeBlockPadding,
		Radius:  Some(cfg.Style.CodeBlockRadius),
		Sizing:  FillFit,
		Clip:    true,
		Content: []View{
			RTF(RtfCfg{
				RichText: block.Content,
				Mode:     TextModeSingleLine,
			}),
		},
	})
}

// Markdown creates a markdown view. Method on *Window to
// access viewState for caching.
func (w *Window) Markdown(cfg MarkdownCfg) View {
	if cfg.Invisible {
		return invisibleContainerView()
	}

	// Cache lookup.
	hash := int64(mathHash(cfg.Source))
	if w.viewState.markdownCache == nil {
		w.viewState.markdownCache =
			NewBoundedMap[int64, []MarkdownBlock](100)
	}
	blocks, ok := w.viewState.markdownCache.Get(hash)
	if !ok {
		blocks = markdownToBlocks(cfg.Source, cfg.Style)
		w.viewState.markdownCache.Set(hash, blocks)
	}

	// Trigger inline math fetches.
	if !cfg.DisableExternalAPIs {
		markdownWarnExternalAPIOnce(w)
		if w.viewState.diagramCache == nil {
			w.viewState.diagramCache =
				NewBoundedDiagramCache(50)
		}
		for _, block := range blocks {
			for _, run := range block.Content.Runs {
				if run.MathID == "" {
					continue
				}
				mhash := mathCacheHash(run.MathID)
				if _, ok := w.viewState.diagramCache.Get(
					mhash); ok {
					continue
				}
				if w.viewState.diagramCache.LoadingCount() >=
					maxConcurrentDiagramFetches {
					continue
				}
				reqID := nextDiagramRequestID(w)
				w.viewState.diagramCache.Set(mhash,
					DiagramCacheEntry{
						State:     DiagramLoading,
						RequestID: reqID,
					})
				fetchMathAsync(w, run.MathLatex, mhash,
					reqID, cfg.Style.MathDPIInline,
					cfg.Style.Text.Color)
			}
		}
	}

	// Build content views from blocks.
	content := make([]View, 0, len(blocks))
	var listItems []View
	prevWasBQ := false

	for i, block := range blocks {
		// Extra space after blockquote group.
		if prevWasBQ && !block.IsBlockquote {
			content = append(content, Rectangle(RectangleCfg{
				Sizing: FillFixed,
				Height: cfg.Style.BlockSpacing,
			}))
		}
		prevWasBQ = block.IsBlockquote

		// Flush accumulated list items.
		if !block.IsList && len(listItems) > 0 {
			content = append(content, Column(ContainerCfg{
				Sizing:  FillFit,
				Spacing: Some(cfg.Style.BlockSpacing / 2),
				Content: listItems,
			}))
			listItems = nil
			content = append(content, Rectangle(RectangleCfg{
				Sizing: FillFixed,
				Height: cfg.Style.BlockSpacing,
			}))
		}

		switch {
		case block.IsMath:
			content = append(content,
				renderMdMath(block, cfg, w))

		case block.IsCode:
			if block.CodeLanguage == "mermaid" {
				content = append(content,
					renderMdMermaid(block, cfg, w))
			} else {
				content = append(content,
					renderMdCode(block, cfg))
			}

		case block.IsTable:
			if block.TableData != nil {
				content = append(content,
					Column(ContainerCfg{
						Sizing: FillFit,
						Clip:   true,
						Content: []View{
							Table(TableCfg{
								BorderStyle:      cfg.Style.TableBorderStyle,
								ColorBorder:      cfg.Style.TableBorderColor,
								SizeBorder:       cfg.Style.TableBorderSize,
								TextStyleHead:    cfg.Style.TableHeadStyle,
								TextStyle:        cfg.Style.TableCellStyle,
								CellPadding:      cfg.Style.TableCellPadding,
								ColorRowAlt:      cfg.Style.TableRowAlt,
								ColumnAlignments: block.TableData.Alignments,
								Data:             buildMarkdownTableData(*block.TableData, cfg.Style),
							}),
						},
					}))
			}

		case block.IsHR:
			content = append(content, Rectangle(RectangleCfg{
				Sizing: FillFixed,
				Height: 1,
				Color:  cfg.Style.HRColor,
			}))

		case block.IsBlockquote:
			leftMargin := float32(
				block.BlockquoteDepth-1) * cfg.Style.NestIndent
			content = append(content, Row(ContainerCfg{
				Sizing:  FillFit,
				Padding: NewPadding(0, 0, 0, leftMargin),
				Content: []View{
					Rectangle(RectangleCfg{
						Sizing: FixedFill,
						Width:  3,
						Color:  cfg.Style.BlockquoteBorder,
					}),
					Column(ContainerCfg{
						Color:  cfg.Style.BlockquoteBG,
						Sizing: FillFit,
						Content: []View{
							RTF(RtfCfg{
								RichText:      block.Content,
								Mode:          cfg.Mode,
								BaseTextStyle: &block.BaseStyle,
							}),
						},
					}),
				},
			}))

		case block.IsImage:
			// Placeholder — image view Phase 18.
			content = append(content, Text(TextCfg{
				Text: "[image: " + block.ImageSrc + "]",
				TextStyle: cfg.Style.Text,
			}))

		case block.HeaderLevel > 0:
			headingContent := []View{
				RTF(RtfCfg{
					ID:            block.AnchorSlug,
					RichText:      block.Content,
					Mode:          cfg.Mode,
					BaseTextStyle: &block.BaseStyle,
				}),
			}
			if (block.HeaderLevel == 1 &&
				cfg.Style.H1Separator) ||
				(block.HeaderLevel == 2 &&
					cfg.Style.H2Separator) {
				headingContent = append(headingContent,
					Rectangle(RectangleCfg{
						Sizing: FillFixed,
						Height: 1,
						Color:  cfg.Style.HRColor,
					}))
			}
			content = append(content, Column(ContainerCfg{
				Sizing:   FillFit,
				A11YRole: AccessRoleHeading,
				A11Y:     &AccessInfo{},
				Content:  headingContent,
			}))

		case block.IsDefTerm:
			content = append(content, RTF(RtfCfg{
				RichText:      block.Content,
				Mode:          cfg.Mode,
				BaseTextStyle: &block.BaseStyle,
			}))

		case block.IsDefValue:
			content = append(content, Row(ContainerCfg{
				Sizing: FillFit,
				Padding: NewPadding(
					0, 0, 0, cfg.Style.NestIndent),
				Content: []View{
					RTF(RtfCfg{
						RichText:      block.Content,
						Mode:          cfg.Mode,
						BaseTextStyle: &block.BaseStyle,
					}),
				},
			}))

		case block.IsList:
			indentW := float32(0)
			if block.ListIndent > 0 {
				indentW = float32(block.ListIndent-1) *
					cfg.Style.NestIndent
			}
			prefixW := float32(len(block.ListPrefix)) *
				cfg.Style.PrefixCharWidth
			if block.ListPrefix == "• " {
				prefixW /= 2
			}
			listItems = append(listItems, Row(ContainerCfg{
				Sizing:  FillFit,
				Padding: NewPadding(0, 0, 0, indentW),
				Content: []View{
					Column(ContainerCfg{
						Sizing: FixedFit,
						Width:  prefixW,
						Content: []View{
							Text(TextCfg{
								Text:      block.ListPrefix,
								TextStyle: cfg.Style.Text,
							}),
						},
					}),
					Column(ContainerCfg{
						Sizing: FillFit,
						Content: []View{
							RTF(RtfCfg{
								RichText:      block.Content,
								Mode:          cfg.Mode,
								BaseTextStyle: &block.BaseStyle,
							}),
						},
					}),
				},
			}))
			// Flush if last block.
			if i == len(blocks)-1 && len(listItems) > 0 {
				content = append(content, Column(ContainerCfg{
					Sizing:  FillFit,
					Spacing: Some(cfg.Style.BlockSpacing / 2),
					Content: listItems,
				}))
				listItems = nil
			}

		default:
			content = append(content, RTF(RtfCfg{
				ID:            cfg.ID,
				IDFocus:       cfg.IDFocus,
				Clip:          cfg.Clip,
				FocusSkip:     cfg.FocusSkip,
				Disabled:      cfg.Disabled,
				MinWidth:      cfg.MinWidth,
				Mode:          cfg.Mode,
				RichText:      block.Content,
				BaseTextStyle: &block.BaseStyle,
			}))
		}
	}

	sizing := FitFit
	if cfg.Mode == TextModeWrap ||
		cfg.Mode == TextModeWrapKeepSpaces {
		sizing = FillFit
	}

	return Column(ContainerCfg{
		A11YRole:    AccessRoleGroup,
		Color:       cfg.Color,
		ColorBorder: cfg.ColorBorder,
		SizeBorder:  Some(cfg.SizeBorder),
		Radius:      Some(cfg.Radius),
		Padding:     cfg.Padding,
		Spacing:     Some(cfg.Style.BlockSpacing),
		Sizing:      sizing,
		Content:     content,
	})
}
