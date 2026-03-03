package markdown

import (
	"strings"
	"testing"
)

// --- tokenizeCode ---

func tokenize(code string, lang CodeLanguage) []CodeToken {
	return tokenizeCode(code, lang, maxCodeBlockHighlightBytes)
}

func tokenText(code string, tok CodeToken) string {
	return code[tok.Start:tok.End]
}

func findToken(code string, tokens []CodeToken, kind CodeTokenKind, text string) bool {
	for _, tok := range tokens {
		if tok.Kind == kind && tokenText(code, tok) == text {
			return true
		}
	}
	return false
}

func requireToken(t *testing.T, code string, tokens []CodeToken, kind CodeTokenKind, text string) {
	t.Helper()
	if !findToken(code, tokens, kind, text) {
		t.Errorf("expected %d token %q in %q", kind, text, code)
	}
}

func TestTokenizeEmpty(t *testing.T) {
	if toks := tokenize("", LangGo); toks != nil {
		t.Errorf("empty code should return nil, got %d tokens", len(toks))
	}
}

func TestTokenizeExceedsMaxBytes(t *testing.T) {
	code := strings.Repeat("x", maxCodeBlockHighlightBytes+1)
	if toks := tokenizeCode(code, LangGo, maxCodeBlockHighlightBytes); toks != nil {
		t.Error("code exceeding maxBytes should return nil")
	}
}

func TestTokenizeCoverage(t *testing.T) {
	code := "func main() { x := 42 }"
	toks := tokenize(code, LangGo)
	if len(toks) == 0 {
		t.Fatal("expected tokens")
	}
	// Reassemble should equal original.
	var sb strings.Builder
	for _, tok := range toks {
		sb.WriteString(tokenText(code, tok))
	}
	if sb.String() != code {
		t.Errorf("reassembly mismatch:\ngot  %q\nwant %q",
			sb.String(), code)
	}
}

// --- Go ---

func TestTokenizeGoKeywords(t *testing.T) {
	code := "func main() {\n\tvar x int\n\treturn\n}"
	toks := tokenize(code, LangGo)
	requireToken(t, code, toks, TokenKeyword, "func")
	requireToken(t, code, toks, TokenKeyword, "var")
	requireToken(t, code, toks, TokenKeyword, "int")
	requireToken(t, code, toks, TokenKeyword, "return")
}

func TestTokenizeGoString(t *testing.T) {
	code := `var s = "hello world"`
	toks := tokenize(code, LangGo)
	requireToken(t, code, toks, TokenString, `"hello world"`)
}

func TestTokenizeGoBacktickString(t *testing.T) {
	code := "var s = `raw string`"
	toks := tokenize(code, LangGo)
	requireToken(t, code, toks, TokenString, "`raw string`")
}

func TestTokenizeGoCharLiteral(t *testing.T) {
	code := "var c = 'a'"
	toks := tokenize(code, LangGo)
	requireToken(t, code, toks, TokenString, "'a'")
}

func TestTokenizeGoNumber(t *testing.T) {
	code := "x := 42"
	toks := tokenize(code, LangGo)
	requireToken(t, code, toks, TokenNumber, "42")
}

func TestTokenizeGoHexNumber(t *testing.T) {
	code := "x := 0xFF"
	toks := tokenize(code, LangGo)
	requireToken(t, code, toks, TokenNumber, "0xFF")
}

func TestTokenizeGoFloatNumber(t *testing.T) {
	code := "x := 3.14"
	toks := tokenize(code, LangGo)
	requireToken(t, code, toks, TokenNumber, "3.14")
}

func TestTokenizeGoExponentNumber(t *testing.T) {
	code := "x := 1e10"
	toks := tokenize(code, LangGo)
	requireToken(t, code, toks, TokenNumber, "1e10")
}

func TestTokenizeGoExponentSign(t *testing.T) {
	code := "x := 1e-5"
	toks := tokenize(code, LangGo)
	requireToken(t, code, toks, TokenNumber, "1e-5")
}

func TestTokenizeGoDotNumber(t *testing.T) {
	code := "x := .5"
	toks := tokenize(code, LangGo)
	requireToken(t, code, toks, TokenNumber, ".5")
}

func TestTokenizeGoLineComment(t *testing.T) {
	code := "x := 1 // comment"
	toks := tokenize(code, LangGo)
	requireToken(t, code, toks, TokenComment, "// comment")
}

func TestTokenizeGoBlockComment(t *testing.T) {
	code := "x := 1 /* block */ y"
	toks := tokenize(code, LangGo)
	requireToken(t, code, toks, TokenComment, "/* block */")
}

func TestTokenizeGoBlockCommentUnclosed(t *testing.T) {
	code := "x /* unclosed"
	toks := tokenize(code, LangGo)
	requireToken(t, code, toks, TokenComment, "/* unclosed")
}

func TestTokenizeGoOperators(t *testing.T) {
	code := "x := a + b"
	toks := tokenize(code, LangGo)
	requireToken(t, code, toks, TokenOperator, ":=")
	requireToken(t, code, toks, TokenOperator, "+")
}

func TestTokenizeGoEscapedString(t *testing.T) {
	code := `var s = "hello\nworld"`
	toks := tokenize(code, LangGo)
	requireToken(t, code, toks, TokenString, `"hello\nworld"`)
}

func TestTokenizeGoIdentifier(t *testing.T) {
	code := "myVar := 1"
	toks := tokenize(code, LangGo)
	// Non-keyword idents are TokenPlain, may merge with
	// adjacent whitespace; check containment.
	found := false
	for _, tok := range toks {
		if tok.Kind == TokenPlain &&
			strings.Contains(tokenText(code, tok), "myVar") {
			found = true
		}
	}
	if !found {
		t.Error("expected plain token containing myVar")
	}
}

func TestTokenizeGoAllBuiltins(t *testing.T) {
	keywords := []string{
		"break", "case", "chan", "const", "continue",
		"default", "defer", "else", "fallthrough", "for",
		"func", "go", "goto", "if", "import", "interface",
		"map", "package", "range", "return", "select",
		"struct", "switch", "type", "var",
	}
	for _, kw := range keywords {
		code := kw + " "
		toks := tokenize(code, LangGo)
		if !findToken(code, toks, TokenKeyword, kw) {
			t.Errorf("Go keyword %q not recognized", kw)
		}
	}
}

// --- Python ---

func TestTokenizePythonKeywords(t *testing.T) {
	code := "def hello():\n    return True"
	toks := tokenize(code, LangPython)
	requireToken(t, code, toks, TokenKeyword, "def")
	requireToken(t, code, toks, TokenKeyword, "return")
	requireToken(t, code, toks, TokenKeyword, "True")
}

func TestTokenizePythonHashComment(t *testing.T) {
	code := "x = 1 # comment"
	toks := tokenize(code, LangPython)
	requireToken(t, code, toks, TokenComment, "# comment")
}

func TestTokenizePythonString(t *testing.T) {
	code := `x = "hello"`
	toks := tokenize(code, LangPython)
	requireToken(t, code, toks, TokenString, `"hello"`)
}

func TestTokenizePythonSingleQuote(t *testing.T) {
	code := "x = 'hello'"
	toks := tokenize(code, LangPython)
	requireToken(t, code, toks, TokenString, "'hello'")
}

func TestTokenizePythonTripleQuote(t *testing.T) {
	code := `x = """triple"""` //nolint:gocritic
	toks := tokenize(code, LangPython)
	requireToken(t, code, toks, TokenString, `"""triple"""`)
}

func TestTokenizePythonTripleSingleQuote(t *testing.T) {
	code := "x = '''triple'''"
	toks := tokenize(code, LangPython)
	requireToken(t, code, toks, TokenString, "'''triple'''")
}

func TestTokenizePythonTripleQuoteUnclosed(t *testing.T) {
	code := `x = """unclosed`
	toks := tokenize(code, LangPython)
	// Should consume to end.
	found := false
	for _, tok := range toks {
		if tok.Kind == TokenString &&
			strings.Contains(tokenText(code, tok), "unclosed") {
			found = true
		}
	}
	if !found {
		t.Error("unclosed triple-quote should be string token")
	}
}

func TestTokenizePythonNoBlockComment(t *testing.T) {
	// Python has no /* */ block comments.
	code := "x = 1 /* not a comment */"
	toks := tokenize(code, LangPython)
	for _, tok := range toks {
		if tok.Kind == TokenComment &&
			strings.Contains(tokenText(code, tok), "/*") {
			t.Error("Python should not have /* block comments")
		}
	}
}

// --- JavaScript ---

func TestTokenizeJSKeywords(t *testing.T) {
	code := "const x = async () => { return await fetch(); }"
	toks := tokenize(code, LangJavaScript)
	requireToken(t, code, toks, TokenKeyword, "const")
	requireToken(t, code, toks, TokenKeyword, "async")
	requireToken(t, code, toks, TokenKeyword, "return")
	requireToken(t, code, toks, TokenKeyword, "await")
}

func TestTokenizeJSTemplateString(t *testing.T) {
	code := "const s = `template`"
	toks := tokenize(code, LangJavaScript)
	requireToken(t, code, toks, TokenString, "`template`")
}

func TestTokenizeJSDollarIdent(t *testing.T) {
	code := "$el = 1"
	toks := tokenize(code, LangJavaScript)
	found := false
	for _, tok := range toks {
		if tok.Kind == TokenPlain &&
			strings.Contains(tokenText(code, tok), "$el") {
			found = true
		}
	}
	if !found {
		t.Error("expected plain token containing $el")
	}
}

func TestTokenizeJSLineComment(t *testing.T) {
	code := "x = 1 // js comment"
	toks := tokenize(code, LangJavaScript)
	requireToken(t, code, toks, TokenComment, "// js comment")
}

func TestTokenizeJSBlockComment(t *testing.T) {
	code := "x = 1 /* block */ y"
	toks := tokenize(code, LangJavaScript)
	requireToken(t, code, toks, TokenComment, "/* block */")
}

// --- TypeScript ---

func TestTokenizeTSKeywords(t *testing.T) {
	code := "interface Foo { readonly x: number }"
	toks := tokenize(code, LangTypeScript)
	requireToken(t, code, toks, TokenKeyword, "interface")
	requireToken(t, code, toks, TokenKeyword, "readonly")
	requireToken(t, code, toks, TokenKeyword, "number")
}

func TestTokenizeTSDollarIdent(t *testing.T) {
	// $ is ident start in TS; "var" is a keyword but
	// "$var" as a whole is not.
	code := "$x := 1"
	toks := tokenize(code, LangTypeScript)
	found := false
	for _, tok := range toks {
		if tok.Kind == TokenPlain &&
			strings.Contains(tokenText(code, tok), "$x") {
			found = true
		}
	}
	if !found {
		t.Error("expected plain token containing $x")
	}
}

// --- Rust ---

func TestTokenizeRustKeywords(t *testing.T) {
	code := "fn main() {\n    let mut x: i32 = 0;\n}"
	toks := tokenize(code, LangRust)
	requireToken(t, code, toks, TokenKeyword, "fn")
	requireToken(t, code, toks, TokenKeyword, "let")
	requireToken(t, code, toks, TokenKeyword, "mut")
	requireToken(t, code, toks, TokenKeyword, "i32")
}

func TestTokenizeRustNestedBlockComment(t *testing.T) {
	code := "x /* outer /* inner */ still comment */ y"
	toks := tokenize(code, LangRust)
	requireToken(t, code, toks, TokenComment,
		"/* outer /* inner */ still comment */")
}

func TestTokenizeRustLineComment(t *testing.T) {
	code := "let x = 1; // rust comment"
	toks := tokenize(code, LangRust)
	requireToken(t, code, toks, TokenComment, "// rust comment")
}

// --- C/C++ ---

func TestTokenizeCKeywords(t *testing.T) {
	code := "int main() { return 0; }"
	toks := tokenize(code, LangC)
	requireToken(t, code, toks, TokenKeyword, "int")
	requireToken(t, code, toks, TokenKeyword, "return")
}

func TestTokenizeCBlockComment(t *testing.T) {
	code := "int x; /* comment */ int y;"
	toks := tokenize(code, LangC)
	requireToken(t, code, toks, TokenComment, "/* comment */")
}

func TestTokenizeCBlockCommentNotNested(t *testing.T) {
	// C does not nest block comments.
	code := "/* outer /* inner */ end"
	toks := tokenize(code, LangC)
	// The first */ closes, so "end" should not be comment.
	for _, tok := range toks {
		if tok.Kind == TokenComment &&
			strings.Contains(tokenText(code, tok), "end") {
			t.Error("C block comments should not nest")
		}
	}
}

// --- Shell ---

func TestTokenizeShellKeywords(t *testing.T) {
	code := "if [ -f file ]; then\n  echo hello\nfi"
	toks := tokenize(code, LangShell)
	requireToken(t, code, toks, TokenKeyword, "if")
	requireToken(t, code, toks, TokenKeyword, "then")
	requireToken(t, code, toks, TokenKeyword, "echo")
	requireToken(t, code, toks, TokenKeyword, "fi")
}

func TestTokenizeShellHashComment(t *testing.T) {
	code := "# shell comment\necho hi"
	toks := tokenize(code, LangShell)
	requireToken(t, code, toks, TokenComment, "# shell comment")
}

func TestTokenizeShellString(t *testing.T) {
	code := `echo "hello world"`
	toks := tokenize(code, LangShell)
	requireToken(t, code, toks, TokenString, `"hello world"`)
}

func TestTokenizeShellNoBlockComment(t *testing.T) {
	code := "echo /* not a comment */"
	toks := tokenize(code, LangShell)
	for _, tok := range toks {
		if tok.Kind == TokenComment &&
			strings.Contains(tokenText(code, tok), "/*") {
			t.Error("Shell should not have /* block comments")
		}
	}
}

// --- JSON ---

func TestTokenizeJSONKeywords(t *testing.T) {
	code := `{"key": true, "b": false, "c": null}`
	toks := tokenize(code, LangJSON)
	requireToken(t, code, toks, TokenKeyword, "true")
	requireToken(t, code, toks, TokenKeyword, "false")
	requireToken(t, code, toks, TokenKeyword, "null")
}

func TestTokenizeJSONString(t *testing.T) {
	code := `{"key": "value"}`
	toks := tokenize(code, LangJSON)
	requireToken(t, code, toks, TokenString, `"key"`)
	requireToken(t, code, toks, TokenString, `"value"`)
}

func TestTokenizeJSONNumber(t *testing.T) {
	code := `{"n": 42}`
	toks := tokenize(code, LangJSON)
	requireToken(t, code, toks, TokenNumber, "42")
}

func TestTokenizeJSONOnlyDoubleQuote(t *testing.T) {
	// JSON only supports double quotes.
	code := "{'key': 'val'}"
	toks := tokenize(code, LangJSON)
	for _, tok := range toks {
		if tok.Kind == TokenString {
			t.Error("JSON should not recognize single-quote strings")
		}
	}
}

// --- HTML ---

func TestTokenizeHTMLKeywords(t *testing.T) {
	code := "<div class=\"container\">text</div>"
	toks := tokenize(code, LangHTML)
	requireToken(t, code, toks, TokenKeyword, "div")
}

func TestTokenizeHTMLComment(t *testing.T) {
	code := "<!-- comment -->"
	toks := tokenize(code, LangHTML)
	requireToken(t, code, toks, TokenComment, "<!-- comment -->")
}

func TestTokenizeHTMLCommentUnclosed(t *testing.T) {
	code := "<!-- unclosed"
	toks := tokenize(code, LangHTML)
	found := false
	for _, tok := range toks {
		if tok.Kind == TokenComment {
			found = true
		}
	}
	if !found {
		t.Error("unclosed HTML comment should still be comment token")
	}
}

func TestTokenizeHTMLDashIdent(t *testing.T) {
	// HTML identifiers can contain hyphens.
	code := "font-size"
	toks := tokenize(code, LangHTML)
	requireToken(t, code, toks, TokenPlain, "font-size")
}

func TestTokenizeHTMLString(t *testing.T) {
	code := `class="foo"`
	toks := tokenize(code, LangHTML)
	requireToken(t, code, toks, TokenString, `"foo"`)
}

// --- V ---

func TestTokenizeVKeywords(t *testing.T) {
	code := "fn main() {\n\tmut x := 0\n\treturn\n}"
	toks := tokenize(code, LangV)
	requireToken(t, code, toks, TokenKeyword, "fn")
	requireToken(t, code, toks, TokenKeyword, "mut")
	requireToken(t, code, toks, TokenKeyword, "return")
}

func TestTokenizeVNestedBlockComment(t *testing.T) {
	code := "/* outer /* inner */ still */ x"
	toks := tokenize(code, LangV)
	requireToken(t, code, toks, TokenComment,
		"/* outer /* inner */ still */")
}

// --- Generic ---

func TestTokenizeGenericLineComment(t *testing.T) {
	code := "x = 1 // generic comment"
	toks := tokenize(code, LangGeneric)
	requireToken(t, code, toks, TokenComment, "// generic comment")
}

func TestTokenizeGenericHashComment(t *testing.T) {
	code := "# generic hash comment"
	toks := tokenize(code, LangGeneric)
	requireToken(t, code, toks, TokenComment,
		"# generic hash comment")
}

func TestTokenizeGenericBlockComment(t *testing.T) {
	code := "x /* block */ y"
	toks := tokenize(code, LangGeneric)
	requireToken(t, code, toks, TokenComment, "/* block */")
}

func TestTokenizeGenericNestedBlockComment(t *testing.T) {
	code := "/* a /* b */ c */ d"
	toks := tokenize(code, LangGeneric)
	requireToken(t, code, toks, TokenComment,
		"/* a /* b */ c */")
}

func TestTokenizeGenericString(t *testing.T) {
	code := `x = "hello"`
	toks := tokenize(code, LangGeneric)
	requireToken(t, code, toks, TokenString, `"hello"`)
}

func TestTokenizeGenericBacktickString(t *testing.T) {
	code := "x = `raw`"
	toks := tokenize(code, LangGeneric)
	requireToken(t, code, toks, TokenString, "`raw`")
}

func TestTokenizeGenericDollarIdent(t *testing.T) {
	code := "$foo = 1"
	toks := tokenize(code, LangGeneric)
	found := false
	for _, tok := range toks {
		if tok.Kind == TokenPlain &&
			strings.Contains(tokenText(code, tok), "$foo") {
			found = true
		}
	}
	if !found {
		t.Error("expected plain token containing $foo")
	}
}

// --- Token merging ---

func TestTokenizeMergesAdjacentSameKind(t *testing.T) {
	// Whitespace tokens should be merged.
	code := "x    y"
	toks := tokenize(code, LangGo)
	for _, tok := range toks {
		if tok.Kind == TokenPlain &&
			tokenText(code, tok) == "    " {
			return // merged whitespace found
		}
	}
	// Check that there's no single-space token sequence.
	spaceCount := 0
	for _, tok := range toks {
		if tok.Kind == TokenPlain &&
			strings.TrimSpace(tokenText(code, tok)) == "" {
			spaceCount++
		}
	}
	if spaceCount > 1 {
		t.Error("adjacent whitespace should be merged")
	}
}

// --- Reassembly for all languages ---

func TestTokenizeReassemblyAllLangs(t *testing.T) {
	samples := map[CodeLanguage]string{
		LangGo:         "func main() { fmt.Println(\"hi\") }",
		LangPython:     "def foo():\n    print('hello')\n",
		LangJavaScript: "const f = () => { return 42; };",
		LangTypeScript: "interface Foo { x: number; }",
		LangRust:       "fn main() { let x: i32 = 42; }",
		LangC:          "int main() { return 0; }",
		LangShell:      "echo \"hello\" # comment",
		LangJSON:       `{"key": "value", "n": 42}`,
		LangHTML:       "<div>text</div><!-- comment -->",
		LangV:          "fn main() { mut x := 0 }",
		LangGeneric:    "x = 1 + 2 // comment",
	}
	for lang, code := range samples {
		toks := tokenize(code, lang)
		if toks == nil {
			t.Errorf("lang %d: nil tokens for %q", lang, code)
			continue
		}
		var sb strings.Builder
		for _, tok := range toks {
			sb.WriteString(tokenText(code, tok))
		}
		if sb.String() != code {
			t.Errorf("lang %d reassembly mismatch:\ngot  %q\nwant %q",
				lang, sb.String(), code)
		}
	}
}

// --- normalizeLanguageHint ---

func TestNormalizeLanguageHint(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"", ""},
		{"  ", ""},
		{"Go", "go"},
		{"PYTHON", "py"},
		{"javascript info", "js"},
		{"ts\textra", "ts"},
		{"vlang", "v"},
		{"rust", "rust"},
		{"cpp", "c"},
		{"c++", "c"},
		{"bash", "shell"},
		{"zsh", "shell"},
		{"htm", "html"},
		{"css", "html"},
		{"svg", "html"},
		{"latex", "math"},
		{"mermaid", "mermaid"},
		{"unknown", "unknown"},
	}
	for _, tc := range tests {
		got := normalizeLanguageHint(tc.input)
		if got != tc.want {
			t.Errorf("normalizeLanguageHint(%q): got %q, want %q",
				tc.input, got, tc.want)
		}
	}
}

// --- isKeyword ---

func TestIsKeywordPerLanguage(t *testing.T) {
	tests := []struct {
		lang CodeLanguage
		word string
		want bool
	}{
		{LangGo, "func", true},
		{LangGo, "notakeyword", false},
		{LangPython, "def", true},
		{LangPython, "func", false},
		{LangJavaScript, "function", true},
		{LangTypeScript, "interface", true},
		{LangRust, "fn", true},
		{LangC, "int", true},
		{LangShell, "echo", true},
		{LangHTML, "div", true},
		{LangJSON, "true", true},
		{LangJSON, "undefined", false},
		{LangV, "fn", true},
	}
	for _, tc := range tests {
		got := isKeyword(tc.word, tc.lang)
		if got != tc.want {
			t.Errorf("isKeyword(%q, %d): got %v, want %v",
				tc.word, tc.lang, got, tc.want)
		}
	}
}

// --- isOperatorChar ---

func TestIsOperatorChar(t *testing.T) {
	ops := "+-*/%=&|^!<>?:.,;()[]{}~"
	for _, ch := range ops {
		if !isOperatorChar(byte(ch)) {
			t.Errorf("expected %q to be operator", string(ch))
		}
	}
	nonOps := "abcXYZ019_ \t\n"
	for _, ch := range nonOps {
		if isOperatorChar(byte(ch)) {
			t.Errorf("expected %q to NOT be operator",
				string(ch))
		}
	}
}

// --- isCodeWhitespace ---

func TestIsCodeWhitespace(t *testing.T) {
	for _, ch := range " \t\n\r" {
		if !isCodeWhitespace(byte(ch)) {
			t.Errorf("expected %q to be whitespace", string(ch))
		}
	}
	if isCodeWhitespace('a') {
		t.Error("'a' should not be whitespace")
	}
}

// --- isIdentifierStart / isIdentifierContinue ---

func TestIsIdentifierStart(t *testing.T) {
	// Letters and underscore always valid.
	for _, ch := range "abcABC_" {
		if !isIdentifierStart(byte(ch), LangGo) {
			t.Errorf("expected %q ident start for Go", string(ch))
		}
	}
	// Digits not valid start.
	if isIdentifierStart('0', LangGo) {
		t.Error("digit should not be ident start")
	}
	// $ valid in JS/TS/Generic.
	if !isIdentifierStart('$', LangJavaScript) {
		t.Error("$ should be ident start in JS")
	}
	if !isIdentifierStart('$', LangTypeScript) {
		t.Error("$ should be ident start in TS")
	}
	if !isIdentifierStart('$', LangGeneric) {
		t.Error("$ should be ident start in Generic")
	}
	if isIdentifierStart('$', LangGo) {
		t.Error("$ should NOT be ident start in Go")
	}
	// - valid in HTML.
	if !isIdentifierStart('-', LangHTML) {
		t.Error("- should be ident start in HTML")
	}
	if isIdentifierStart('-', LangGo) {
		t.Error("- should NOT be ident start in Go")
	}
}

func TestIsIdentifierContinue(t *testing.T) {
	if !isIdentifierContinue('5', LangGo) {
		t.Error("digit should be ident continue")
	}
	if !isIdentifierContinue('-', LangHTML) {
		t.Error("- should be ident continue in HTML")
	}
	if isIdentifierContinue('-', LangPython) {
		t.Error("- should NOT be ident continue in Python")
	}
}

// --- isNumberStart ---

func TestIsNumberStart(t *testing.T) {
	if !isNumberStart("42", 0) {
		t.Error("digit should be number start")
	}
	if !isNumberStart(".5", 0) {
		t.Error("dot-digit should be number start")
	}
	if isNumberStart(".x", 0) {
		t.Error("dot-letter should NOT be number start")
	}
	if isNumberStart("abc", 0) {
		t.Error("letter should NOT be number start")
	}
}

// --- scanNumber ---

func TestScanNumber(t *testing.T) {
	tests := []struct {
		code string
		want string
	}{
		{"42 ", "42"},
		{"3.14 ", "3.14"},
		{"0xFF ", "0xFF"},
		{"0b1010 ", "0b1010"},
		{"0o77 ", "0o77"},
		{"1e10 ", "1e10"},
		{"1E+5 ", "1E+5"},
		{"1_000 ", "1_000"},
	}
	for _, tc := range tests {
		end, ok := scanNumber(tc.code, 0)
		if !ok {
			t.Errorf("scanNumber(%q): not ok", tc.code)
			continue
		}
		got := tc.code[:end]
		if got != tc.want {
			t.Errorf("scanNumber(%q): got %q, want %q",
				tc.code, got, tc.want)
		}
	}
}

// --- isStringDelim ---

func TestIsStringDelim(t *testing.T) {
	// JSON: only double quote.
	if !isStringDelim('"', LangJSON) {
		t.Error("JSON should accept double quote")
	}
	if isStringDelim('\'', LangJSON) {
		t.Error("JSON should reject single quote")
	}
	if isStringDelim('`', LangJSON) {
		t.Error("JSON should reject backtick")
	}
	// Go: double, single, backtick.
	if !isStringDelim('`', LangGo) {
		t.Error("Go should accept backtick")
	}
	// Python: double, single only.
	if isStringDelim('`', LangPython) {
		t.Error("Python should reject backtick")
	}
}

// --- hasLineCommentStart ---

func TestHasLineCommentStart(t *testing.T) {
	if !hasLineCommentStart("// c", 0, LangGo) {
		t.Error("Go should have // comment")
	}
	if hasLineCommentStart("// c", 0, LangPython) {
		t.Error("Python should not have // comment")
	}
	if !hasLineCommentStart("# c", 0, LangPython) {
		t.Error("Python should have # comment")
	}
	if hasLineCommentStart("# c", 0, LangGo) {
		t.Error("Go should not have # comment")
	}
	if !hasLineCommentStart("# c", 0, LangShell) {
		t.Error("Shell should have # comment")
	}
}

// --- hasBlockCommentStart ---

func TestHasBlockCommentStart(t *testing.T) {
	if !hasBlockCommentStart("/* c", 0, LangGo) {
		t.Error("Go should have /* comment")
	}
	if hasBlockCommentStart("/* c", 0, LangPython) {
		t.Error("Python should not have /* comment")
	}
	if !hasBlockCommentStart("<!-- c", 0, LangHTML) {
		t.Error("HTML should have <!-- comment")
	}
	if hasBlockCommentStart("<!-- c", 0, LangGo) {
		t.Error("Go should not have <!-- comment")
	}
}

// --- blockCommentsNested ---

func TestBlockCommentsNested(t *testing.T) {
	if !blockCommentsNested(LangV) {
		t.Error("V should nest block comments")
	}
	if !blockCommentsNested(LangRust) {
		t.Error("Rust should nest block comments")
	}
	if !blockCommentsNested(LangGeneric) {
		t.Error("Generic should nest block comments")
	}
	if blockCommentsNested(LangGo) {
		t.Error("Go should NOT nest block comments")
	}
	if blockCommentsNested(LangC) {
		t.Error("C should NOT nest block comments")
	}
}

// --- lineCommentPrefixLen ---

func TestLineCommentPrefixLen(t *testing.T) {
	if lineCommentPrefixLen("// c", 0, LangGo) != 2 {
		t.Error("Go // should have prefix len 2")
	}
	if lineCommentPrefixLen("# c", 0, LangPython) != 1 {
		t.Error("Python # should have prefix len 1")
	}
}

// --- multiline code ---

func TestTokenizeMultilineGo(t *testing.T) {
	code := "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"hello\")\n}"
	toks := tokenize(code, LangGo)
	requireToken(t, code, toks, TokenKeyword, "package")
	requireToken(t, code, toks, TokenKeyword, "import")
	requireToken(t, code, toks, TokenKeyword, "func")
	requireToken(t, code, toks, TokenString, `"fmt"`)
	requireToken(t, code, toks, TokenString, `"hello"`)
}
