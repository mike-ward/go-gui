package gui

// view_markdown.go defines the Markdown view component.
// Parses markdown source and renders it using RTF views.

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/mike-ward/go-gui/gui/markdown"
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
	CodeBlockText     TextStyle
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
	CodeBlockPadding  Opt[Padding]
	CodeBlockRadius   float32
	H1Separator       bool
	H2Separator       bool
	TableBorderStyle  TableBorderStyle
	TableBorderColor  Color
	TableBorderSize   float32
	TableHeadStyle    TextStyle
	TableCellStyle    TextStyle
	TableCellPadding  Opt[Padding]
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
		CodeBlockText:     guiTheme.M5,
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
		PrefixCharWidth:   4,
		CodeBlockPadding:  Some(PadAll(10)),
		CodeBlockRadius:   3.5,
		TableBorderStyle:  TableBorderHeaderOnly,
		TableBorderColor:  guiTheme.ColorBorder,
		TableBorderSize:   1,
		TableHeadStyle:    guiTheme.B4,
		TableCellStyle:    guiTheme.N4,
		TableCellPadding:  SomeP(5, 10, 5, 10),
		HighlightBG:       RGB(199, 142, 18),
		MathDPIDisplay:    150,
		MathDPIInline:     200,
		MermaidBG:         RGBA(248, 248, 255, 255),
	}
}

// MarkdownCfg configures a Markdown View. Mode defaults to wrapped text.
type MarkdownCfg struct {
	ID                  string
	Source              string
	Style               MarkdownStyle
	IDFocus             uint32
	Mode                Opt[TextMode]
	MinWidth            float32
	Invisible           bool
	Clip                bool
	FocusSkip           bool
	Disabled            bool
	Color               Color
	ColorBorder         Color
	SizeBorder          Opt[float32]
	Radius              Opt[float32]
	Padding             Opt[Padding]
	MermaidWidth        int // max pixel width for mermaid diagrams (0 = 600)
	DisableExternalAPIs bool
}

var markdownExternalAPIsEnabled bool

// SetMarkdownExternalAPIsEnabled toggles external markdown API usage
// (CodeCogs/Kroki). Default is disabled.
func SetMarkdownExternalAPIsEnabled(enabled bool) {
	markdownExternalAPIsEnabled = enabled
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

func ensureDiagramCache(w *Window) *BoundedDiagramCache {
	if w.viewState.diagramCache == nil {
		w.viewState.diagramCache =
			NewBoundedDiagramCache(50)
	}
	return w.viewState.diagramCache
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

func markdownDiagramErrorView(
	errText string, baseStyle TextStyle,
) View {
	errStyle := baseStyle
	errStyle.Color = RGBA(200, 50, 50, 255)
	return Text(TextCfg{
		Text:      errText,
		TextStyle: errStyle,
		Mode:      TextModeWrap,
	})
}

// buildMarkdownTableData converts ParsedTable to
// TableRowCfg array.
func buildMarkdownTableData(
	parsed ParsedTable, _ MarkdownStyle,
) []TableRowCfg {
	rows := make([]TableRowCfg, 0, len(parsed.Rows)+1)

	// Header row.
	hCells := make([]TableCellCfg, 0, len(parsed.Headers))
	for _, h := range parsed.Headers {
		rt := h
		hCells = append(hCells, TableCellCfg{
			Value:    richTextPlain(h),
			HeadCell: true,
			RichText: &rt,
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
			rt := cell
			cells = append(cells, TableCellCfg{
				Value:    richTextPlain(cell),
				RichText: &rt,
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
		SizeBorder: NoBorder,
		Sizing:     FillFit,
		Content: []View{
			Text(TextCfg{
				Text:      block.MathLatex,
				TextStyle: cfg.Style.Code,
			}),
		},
	})

	if cfg.DisableExternalAPIs || !markdownExternalAPIsEnabled {
		return codeFallback
	}

	diagramHash := diagramCacheHash(
		fmt.Sprintf("display_%d", markdown.MathHash(block.MathLatex)))

	cache := ensureDiagramCache(w)
	if entry, ok := cache.Get(diagramHash); ok {
		switch entry.State {
		case DiagramLoading:
			return codeFallback
		case DiagramReady:
			return Image(ImageCfg{
				Src:    entry.PNGPath,
				Width:  entry.Width,
				Height: entry.Height,
			})
		case DiagramError:
			return markdownDiagramErrorView(
				entry.Error, cfg.Style.Code,
			)
		}
	}

	// Start async fetch.
	if cache.LoadingCount() <
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
		Color:      cfg.Style.CodeBlockBG,
		Padding:    cfg.Style.CodeBlockPadding,
		Radius:     Some(cfg.Style.CodeBlockRadius),
		SizeBorder: NoBorder,
		Sizing:     FillFit,
		Content: []View{
			RTF(RtfCfg{
				RichText: block.Content,
				Mode:     TextModeSingleLine,
			}),
		},
	})

	if cfg.DisableExternalAPIs || !markdownExternalAPIsEnabled {
		return codeFallback
	}

	diagramHash := diagramCacheHash(source)

	cache := ensureDiagramCache(w)
	if entry, ok := cache.Get(diagramHash); ok {
		switch entry.State {
		case DiagramLoading:
			return Text(TextCfg{
				Text:      "Loading diagram...",
				TextStyle: cfg.Style.Text,
			})
		case DiagramReady:
			imgW, imgH := entry.Width, entry.Height
			mw := float32(cfg.MermaidWidth)
			if mw <= 0 {
				mw = 600
			}
			if imgW > mw {
				imgH *= mw / imgW
				imgW = mw
			}
			return Image(ImageCfg{
				Src:     entry.PNGPath,
				Width:   imgW,
				Height:  imgH,
				BgColor: White,
			})
		case DiagramError:
			return markdownDiagramErrorView(
				entry.Error, cfg.Style.Code,
			)
		}
	}

	if cache.LoadingCount() <
		maxConcurrentDiagramFetches {
		reqID := nextDiagramRequestID(w)
		cache.Set(diagramHash,
			DiagramCacheEntry{
				State:     DiagramLoading,
				RequestID: reqID,
			})
		fetchMermaidAsync(w, source, diagramHash, reqID)
	}
	return codeFallback
}

// mdCopyButton builds a floating copy-to-clipboard button
// with a 2-second check-mark animation.
func mdCopyButton(
	animID string, w *Window,
	onClick func(*Layout, *Event, *Window),
) View {
	copied := w.hasAnimationLocked(animID)

	iconStyle := guiTheme.Icon5
	iconStyle.Color = Gray

	var btnContent []View
	if copied {
		checkStyle := iconStyle
		checkStyle.Color = Color{80, 200, 80, 255, true}
		btnContent = []View{
			Text(TextCfg{Text: IconCheck, TextStyle: checkStyle}),
		}
	} else {
		btnContent = []View{
			Text(TextCfg{Text: IconFile, TextStyle: iconStyle}),
		}
	}

	return Button(ButtonCfg{
		Float:        true,
		FloatAnchor:  FloatTopRight,
		FloatTieOff:  FloatTopRight,
		FloatOffsetX: -4,
		FloatOffsetY: 4,
		Radius:       Some[float32](4),
		Color:        ColorTransparent,
		SizeBorder:   Some[float32](0),
		Padding:      SomeP(2, 4, 2, 4),
		Content:      btnContent,
		OnClick:      onClick,
	})
}

// renderMdCode renders a fenced code block with a copy-to-clipboard button.
func renderMdCode(
	block MarkdownBlock, cfg MarkdownCfg, w *Window, blockIdx int,
) View {
	animID := "md_cp_" + strconv.Itoa(blockIdx)
	copyBtn := mdCopyButton(animID, w,
		func(_ *Layout, e *Event, w *Window) {
			plain := richTextPlain(block.Content)
			w.SetClipboard(plain)
			w.AnimationAdd(&Animate{
				AnimID:   animID,
				Delay:    2 * time.Second,
				Callback: func(*Animate, *Window) {},
			})
			e.IsHandled = true
		})

	return Column(ContainerCfg{
		Color:      cfg.Style.CodeBlockBG,
		Padding:    cfg.Style.CodeBlockPadding,
		Radius:     Some(cfg.Style.CodeBlockRadius),
		SizeBorder: NoBorder,
		Sizing:     FillFit,
		Clip:       true,
		Content: []View{
			RTF(RtfCfg{
				RichText: block.Content,
				Mode:     TextModeSingleLine,
			}),
			copyBtn,
		},
	})
}

// Markdown creates a markdown view. Method on *Window to
// access viewState for caching.
func (w *Window) Markdown(cfg MarkdownCfg) View {
	if cfg.Invisible {
		return invisibleContainerView()
	}
	mode := cfg.Mode.Get(TextModeWrap)

	// Cache lookup; invalidate on theme change.
	hash := int64(markdown.MathHash(cfg.Source))
	themeName := guiTheme.Name
	if w.viewState.markdownCache == nil ||
		w.viewState.markdownTheme != themeName {
		w.viewState.markdownCache =
			NewBoundedMap[int64, []MarkdownBlock](100)
		w.viewState.markdownTheme = themeName
	}
	blocks, ok := w.viewState.markdownCache.Get(hash)
	if !ok {
		blocks = markdownToBlocks(cfg.Source, cfg.Style)
		w.viewState.markdownCache.Set(hash, blocks)
	}

	allowExternalAPIs := markdownExternalAPIsEnabled &&
		!cfg.DisableExternalAPIs

	// Trigger inline math fetches.
	if allowExternalAPIs {
		markdownWarnExternalAPIOnce(w)
		inlineCache := ensureDiagramCache(w)
		for _, block := range blocks {
			for _, run := range block.Content.Runs {
				if run.MathID == "" {
					continue
				}
				mhash := diagramCacheHash(run.MathID)
				if _, ok := inlineCache.Get(mhash); ok {
					continue
				}
				if inlineCache.LoadingCount() >=
					maxConcurrentDiagramFetches {
					continue
				}
				reqID := nextDiagramRequestID(w)
				inlineCache.Set(mhash,
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
				Sizing:     FillFit,
				Padding:    NoPadding,
				SizeBorder: NoBorder,
				Spacing:    Some(cfg.Style.BlockSpacing / 2),
				Content:    listItems,
			}))
			listItems = nil
		}

		switch {
		case block.IsMath:
			content = append(content, Column(ContainerCfg{
				Sizing:     FillFit,
				HAlign:     HAlignCenter,
				SizeBorder: NoBorder,
				Content: []View{
					renderMdMath(block, cfg, w),
				},
			}))

		case block.IsCode:
			if block.CodeLanguage == "mermaid" {
				content = append(content, Column(ContainerCfg{
					Sizing:     FillFit,
					HAlign:     HAlignCenter,
					SizeBorder: NoBorder,
					Content: []View{
						renderMdMermaid(block, cfg, w),
					},
				}))
			} else {
				content = append(content,
					renderMdCode(block, cfg, w, i))
			}

		case block.IsTable:
			if block.TableData != nil {
				content = append(content,
					Column(ContainerCfg{
						Sizing:  FillFit,
						Padding: NoPadding,
						Clip:    true,
						Content: []View{
							w.Table(TableCfg{
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
				Sizing:     FillFit,
				Padding:    SomeP(0, 0, 0, leftMargin),
				SizeBorder: NoBorder,
				Content: []View{
					Rectangle(RectangleCfg{
						Sizing: FixedFill,
						Width:  3,
						Color:  cfg.Style.BlockquoteBorder,
					}),
					Column(ContainerCfg{
						Color:      cfg.Style.BlockquoteBG,
						Sizing:     FillFit,
						Padding:    NoPadding,
						SizeBorder: NoBorder,
						Content: []View{
							RTF(RtfCfg{
								RichText:      block.Content,
								Mode:          mode,
								BaseTextStyle: &block.BaseStyle,
							}),
						},
					}),
				},
			}))

		case block.IsImage:
			content = append(content, Image(ImageCfg{
				Src:    block.ImageSrc,
				Width:  block.ImageWidth,
				Height: block.ImageHeight,
			}))

		case block.HeaderLevel > 0:
			if block.HeaderLevel == 1 {
				content = append(content, Rectangle(RectangleCfg{
					Sizing: FillFixed,
					Height: 3,
				}))
			}
			headingContent := []View{
				RTF(RtfCfg{
					ID:            block.AnchorSlug,
					RichText:      block.Content,
					Mode:          mode,
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
				Padding:  NoPadding,
				A11YRole: AccessRoleHeading,
				A11Y:     &AccessInfo{},
				Content:  headingContent,
			}))

		case block.IsDefTerm:
			content = append(content, RTF(RtfCfg{
				RichText:      block.Content,
				Mode:          mode,
				BaseTextStyle: &block.BaseStyle,
			}))

		case block.IsDefValue:
			content = append(content, Row(ContainerCfg{
				Sizing: FillFit,
				Padding: SomeP(
					0, 0, 0, cfg.Style.NestIndent),
				Content: []View{
					RTF(RtfCfg{
						RichText:      block.Content,
						Mode:          mode,
						BaseTextStyle: &block.BaseStyle,
					}),
				},
			}))

		case block.IsList:
			indentW := float32(block.ListIndent) *
				cfg.Style.NestIndent
			prefixW := float32(len(block.ListPrefix)) *
				cfg.Style.PrefixCharWidth
			if block.ListPrefix == "• " {
				prefixW /= 2
			} else if block.ListIndent > 0 {
				indentW += 4
			}
			listItems = append(listItems, Row(ContainerCfg{
				Sizing:     FillFit,
				Padding:    SomeP(0, 0, 0, indentW),
				SizeBorder: NoBorder,
				Content: []View{
					Column(ContainerCfg{
						Sizing:     FixedFit,
						Padding:    NoPadding,
						SizeBorder: NoBorder,
						Width:      prefixW,
						Content: []View{
							Text(TextCfg{
								Text:      block.ListPrefix,
								TextStyle: cfg.Style.Text,
							}),
						},
					}),
					Column(ContainerCfg{
						Sizing:     FillFit,
						Padding:    NoPadding,
						SizeBorder: NoBorder,
						Content: []View{
							RTF(RtfCfg{
								RichText:      block.Content,
								Mode:          mode,
								BaseTextStyle: &block.BaseStyle,
							}),
						},
					}),
				},
			}))
			// Flush if last block.
			if i == len(blocks)-1 && len(listItems) > 0 {
				content = append(content, Column(ContainerCfg{
					Sizing:     FillFit,
					Padding:    NoPadding,
					SizeBorder: NoBorder,
					Spacing:    Some(cfg.Style.BlockSpacing / 2),
					Content:    listItems,
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
				Mode:          mode,
				RichText:      block.Content,
				BaseTextStyle: &block.BaseStyle,
			}))
		}
	}

	sizing := FitFit
	if mode == TextModeWrap ||
		mode == TextModeWrapKeepSpaces {
		sizing = FillFit
	}

	// Document-level copy button.
	docAnimID := "md_cp_doc"
	source := cfg.Source
	content = append(content, mdCopyButton(docAnimID, w,
		func(_ *Layout, e *Event, w *Window) {
			w.SetClipboard(source)
			w.AnimationAdd(&Animate{
				AnimID:   docAnimID,
				Delay:    2 * time.Second,
				Callback: func(*Animate, *Window) {},
			})
			e.IsHandled = true
		}))

	return Column(ContainerCfg{
		A11YRole:    AccessRoleGroup,
		Color:       cfg.Color,
		ColorBorder: cfg.ColorBorder,
		SizeBorder:  cfg.SizeBorder,
		Radius:      cfg.Radius,
		Padding:     cfg.Padding,
		Spacing:     Some(cfg.Style.BlockSpacing),
		Sizing:      sizing,
		Content:     content,
	})
}
