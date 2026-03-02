package gui

// markdown_types.go defines styled markdown block types.
// These are the output of the styling bridge that converts
// parser MdBlocks into GUI-ready MarkdownBlocks.

// MarkdownBlock is a parsed, styled block of markdown.
type MarkdownBlock struct {
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
	BaseStyle       TextStyle
	Content         RichText
	TableData       *ParsedTable
}

// ParsedTable is a parsed, styled markdown table.
type ParsedTable struct {
	Headers    []RichText
	Alignments []HorizontalAlign
	Rows       [][]RichText
}
