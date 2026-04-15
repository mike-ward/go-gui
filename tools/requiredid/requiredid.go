// Package requiredid provides a go/analysis pass that flags composite
// literals of struct types whose fields are tagged `gui:"required"`
// when the required field is absent or set to an empty string literal.
package requiredid

import (
	"go/ast"
	"go/token"
	"go/types"
	"reflect"
	"slices"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// ignoreDirective is the line-comment marker that suppresses the
// diagnostic for a single composite literal. Place it on the same
// line as the literal's opening brace.
const ignoreDirective = "requiredid:ignore"

// Analyzer is the go/analysis pass.
var Analyzer = &analysis.Analyzer{
	Name: "requiredid",
	Doc:  "flags gui Cfg literals missing a required tagged field",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	for _, f := range pass.Files {
		ignored := ignoredLines(pass.Fset, f)
		parents := parentCalls(f)
		ast.Inspect(f, func(n ast.Node) bool {
			lit, ok := n.(*ast.CompositeLit)
			if !ok {
				return true
			}
			t := pass.TypesInfo.TypeOf(lit)
			if t == nil {
				return true
			}
			named, ok := t.(*types.Named)
			if !ok {
				return true
			}
			st, ok := named.Underlying().(*types.Struct)
			if !ok {
				return true
			}
			if !isFactoryArg(lit, named.Obj().Name(), parents) {
				return true
			}
			required := requiredFields(st)
			if len(required) == 0 {
				return true
			}
			if ignored[pass.Fset.Position(lit.Pos()).Line] {
				return true
			}
			for _, name := range required {
				if !hasNonEmptyField(lit, name) {
					pass.Reportf(lit.Pos(),
						"%s.%s is required (gui:\"required\") and must be non-empty",
						named.Obj().Name(), name)
				}
			}
			return true
		})
	}
	return nil, nil
}

// parentCalls maps each CompositeLit directly used as a call
// argument to that enclosing CallExpr.
func parentCalls(f *ast.File) map[*ast.CompositeLit]*ast.CallExpr {
	out := map[*ast.CompositeLit]*ast.CallExpr{}
	ast.Inspect(f, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		for _, a := range call.Args {
			if lit, ok := a.(*ast.CompositeLit); ok {
				out[lit] = call
			}
		}
		return true
	})
	return out
}

func isFactoryArg(lit *ast.CompositeLit, typeName string, parents map[*ast.CompositeLit]*ast.CallExpr) bool {
	call, ok := parents[lit]
	if !ok {
		return false
	}
	want := strings.TrimSuffix(typeName, "Cfg")
	if want == typeName {
		return false
	}
	var name string
	switch fn := call.Fun.(type) {
	case *ast.Ident:
		name = fn.Name
	case *ast.SelectorExpr:
		name = fn.Sel.Name
	default:
		return false
	}
	return name == want
}

func ignoredLines(fset *token.FileSet, f *ast.File) map[int]bool {
	out := map[int]bool{}
	for _, cg := range f.Comments {
		for _, c := range cg.List {
			if strings.Contains(c.Text, ignoreDirective) {
				out[fset.Position(c.Slash).Line] = true
			}
		}
	}
	return out
}

func requiredFields(st *types.Struct) []string {
	var out []string
	for i := range st.NumFields() {
		tag := reflect.StructTag(st.Tag(i)).Get("gui")
		if tag == "" {
			continue
		}
		if slices.Contains(strings.Split(tag, ","), "required") {
			out = append(out, st.Field(i).Name())
		}
	}
	return out
}

func hasNonEmptyField(lit *ast.CompositeLit, name string) bool {
	for _, e := range lit.Elts {
		kv, ok := e.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		id, ok := kv.Key.(*ast.Ident)
		if !ok || id.Name != name {
			continue
		}
		return !isEmptyString(kv.Value)
	}
	return false
}

func isEmptyString(e ast.Expr) bool {
	lit, ok := e.(*ast.BasicLit)
	return ok && (lit.Value == `""` || lit.Value == "``")
}
