package gui

// md_types.go defines parser output types for the markdown
// pipeline. These are the intermediate types between the
// goldmark AST and the gui styling layer.

// Limits for multi-line constructs.
const (
	maxBlockquoteLines          = 100
	maxTableLines               = 500
	maxTableColumns             = 100
	maxListContinuationLines    = 50
	maxFootnoteContinuationLines = 20
	maxParagraphContinuationLines = 100
	maxCodeBlockLines           = 10000
	maxMathBlockLines           = 200
)

// Inline parsing limits.
const maxInlineNestingDepth = 16

// Source length caps for external API submissions.
const (
	maxLatexSourceLen   = 2000
	maxMermaidSourceLen = 10000
)

// Metadata collection limits.
const (
	maxAbbreviationDefs = 1000
	maxFootnoteDefs     = 10000
	maxLinkDefs         = 10000
)

// Highlight limits.
const (
	maxCodeBlockHighlightBytes  = 131072
	maxInlineCodeHighlightBytes = 2048
	maxHighlightTokensPerBlock  = 16384
	maxHighlightTokenBytes      = 4096
	maxHighlightCommentDepth    = 16
	maxHighlightStringScanBytes = 32768
	maxHighlightIdentifierBytes = 256
	maxHighlightNumberBytes     = 128
)

// MdAlign represents column alignment in tables.
type MdAlign uint8

const (
	MdAlignStart  MdAlign = iota
	MdAlignEnd
	MdAlignCenter
	MdAlignLeft
	MdAlignRight
)

// MdFormat represents inline text formatting.
type MdFormat uint8

const (
	MdFormatPlain MdFormat = iota
	MdFormatBold
	MdFormatItalic
	MdFormatBoldItalic
	MdFormatCode
)

// MdCodeTokenKind classifies syntax highlighting tokens.
type MdCodeTokenKind uint8

const (
	MdTokenPlain MdCodeTokenKind = iota
	MdTokenKeyword
	MdTokenString
	MdTokenNumber
	MdTokenComment
	MdTokenOperator
)

// MdCodeLanguage identifies a programming language for
// syntax highlighting.
type MdCodeLanguage uint8

const (
	MdLangGeneric MdCodeLanguage = iota
	MdLangV
	MdLangJavaScript
	MdLangTypeScript
	MdLangPython
	MdLangJSON
	MdLangGo
	MdLangRust
	MdLangC
	MdLangShell
	MdLangHTML
)

// MdCodeToken represents a highlighted token span.
type MdCodeToken struct {
	Kind  MdCodeTokenKind
	Start int
	End   int
}

// MdRun is a single inline text segment with format and flags.
type MdRun struct {
	Text          string
	Format        MdFormat
	Strikethrough bool
	Underline     bool
	Highlight     bool // ==text==
	Superscript   bool
	Subscript     bool
	Link          string
	Tooltip       string // abbreviation or footnote
	MathID        string
	MathLatex     string
	CodeToken     MdCodeTokenKind
}

// MdBlock is a parsed block of markdown content.
type MdBlock struct {
	HeaderLevel     int
	IsCode          bool
	IsHR            bool
	IsBlockquote    bool
	IsImage         bool
	IsTable         bool
	IsList          bool
	IsMath          bool
	IsDefTerm       bool
	IsDefValue      bool
	BlockquoteDepth int
	ListPrefix      string
	ListIndent      int
	ImageSrc        string
	ImageAlt        string
	ImageWidth      float32
	ImageHeight     float32
	CodeLanguage    string
	MathLatex       string
	AnchorSlug      string
	Runs            []MdRun
	TableData       *MdTable
}

// MdTable holds parsed table data.
type MdTable struct {
	Headers    [][]MdRun
	Alignments []MdAlign
	Rows       [][][]MdRun
	ColCount   int
}
