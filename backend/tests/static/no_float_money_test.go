//go:build static

// Static AST check — no float for money fields.
//
// Doctrine refs: Rule 25, Rule 46 (REJ-13).
// Mitigates: CLAUDE.md cross-portfolio invariant violation.
//
// Allowlist: internal/domain/emissions uses float64 for kgCO2e (legitimate
// physical quantity, not money).

package static_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var moneyFieldRegex = regexp.MustCompile(`(?i)(amount|price|cost|euro|fee|charge|invoice|payment|payable|receivable|refund)`)

var allowedFloatFiles = []string{
	"backend/internal/domain/emissions/", // legitimate physical quantity
	"backend/cmd/simulator/",             // dev-only RNG
}

func TestNoFloatForMoneyFields(t *testing.T) {
	t.Skip("scaffold — implement once internal/domain/ split lands; walk ast.File for fields named (amount|price|cost|...)")

	root := "../../"
	fset := token.NewFileSet()

	violations := []string{}
	walk := func(path string, info interface{}, err error) error {
		_ = info
		if err != nil || !strings.HasSuffix(path, ".go") {
			return err
		}
		for _, allow := range allowedFloatFiles {
			if strings.Contains(path, allow) {
				return nil
			}
		}
		f, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if err != nil {
			return nil
		}
		ast.Inspect(f, func(n ast.Node) bool {
			st, ok := n.(*ast.StructType)
			if !ok {
				return true
			}
			for _, fld := range st.Fields.List {
				for _, name := range fld.Names {
					if !moneyFieldRegex.MatchString(name.Name) {
						continue
					}
					if id, ok := fld.Type.(*ast.Ident); ok {
						if id.Name == "float32" || id.Name == "float64" {
							violations = append(violations, path+": "+name.Name+" "+id.Name)
						}
					}
				}
			}
			return true
		})
		return nil
	}

	if err := filepathWalk(root, walk); err != nil {
		t.Fatalf("walk: %v", err)
	}
	if len(violations) > 0 {
		t.Errorf("float-for-money violations:\n%s", strings.Join(violations, "\n"))
	}
}

func filepathWalk(root string, fn filepath.WalkFunc) error {
	return filepath.Walk(root, fn)
}
