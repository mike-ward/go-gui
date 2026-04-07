package highlight

import (
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

// curatedLangs lists the chroma lexer IDs bundled by this package.
// Importing specific lexer subpackages would be ideal for binary
// size, but chroma registers all lexers via lexers.Get regardless;
// the curated list exists to define which languages this package
// officially supports and to validate user input.
var curatedLangs = map[string]struct{}{
	"go":         {},
	"golang":     {},
	"python":     {},
	"py":         {},
	"javascript": {},
	"js":         {},
	"typescript": {},
	"ts":         {},
	"rust":       {},
	"c":          {},
	"cpp":        {},
	"c++":        {},
	"java":       {},
	"ruby":       {},
	"rb":         {},
	"bash":       {},
	"sh":         {},
	"shell":      {},
	"html":       {},
	"css":        {},
	"json":       {},
	"yaml":       {},
	"yml":        {},
	"toml":       {},
	"sql":        {},
	"markdown":   {},
	"md":         {},
	"diff":       {},
	"dockerfile": {},
	"docker":     {},
	"make":       {},
	"makefile":   {},
}

// lookupLexer returns a chroma lexer for the given language tag.
// Returns nil if the language is not in the curated set and cannot
// be resolved by chroma.
func lookupLexer(lang string) chroma.Lexer {
	if lang == "" {
		return nil
	}
	key := strings.ToLower(lang)
	if _, ok := curatedLangs[key]; !ok {
		return nil
	}
	lx := lexers.Get(key)
	if lx == nil {
		return nil
	}
	return chroma.Coalesce(lx)
}
