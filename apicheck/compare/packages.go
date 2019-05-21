package compare

import (
	"fmt"
	"go/ast"
	"go/token"
	"log"
	"sort"
	"unicode"
	"unicode/utf8"

	"github.com/pkg/errors"
)

// DeclChange tracks the addition or removal of an identifier declaration in a package.
type DeclChange struct {
	path  string
	name  string
	added bool
}

func (c DeclChange) String() string {
	name := c.path + "." + c.name
	if c.added {
		return fmt.Sprintf("identifier %q added", name)
	}
	return fmt.Sprintf("identifier %q removed", name)
}

// Compatible returns true when the change is backwards compatible.
func (c DeclChange) Compatible() bool { return c.added }

// Packages compares two packages and returns the list of changes in its identifiers.
func Packages(importPath string, base, target *ast.Package) ([]Change, error) {
	baseDecls := declsByName(base)
	targetDecls := declsByName(target)

	declsMap := make(map[string]bool)
	for name := range baseDecls {
		declsMap[name] = true
	}
	for name := range targetDecls {
		declsMap[name] = true
	}
	names := make([]string, 0, len(declsMap))
	for name := range declsMap {
		names = append(names, name)
	}
	sort.Strings(names)

	var changes []Change

	for _, name := range names {
		if _, ok := baseDecls[name]; !ok {
			changes = append(changes, DeclChange{importPath, name, true})
			continue
		}
		if _, ok := targetDecls[name]; !ok {
			changes = append(changes, DeclChange{importPath, name, false})
			continue
		}

		cs, err := Decls(name, baseDecls[name], targetDecls[name])
		if err != nil {
			return nil, errors.Wrapf(err, "could not compare decls for %s", name)
		}
		changes = append(changes, cs...)

	}

	return changes, nil
}

func isUnexported(name string) bool {
	r, _ := utf8.DecodeRuneInString(name)
	return r == '_' || unicode.IsLower(r)
}

type valueAndType struct{ value, typ ast.Expr }

func declsByName(pkg *ast.Package) map[string]interface{} {
	decls := make(map[string]interface{})

	addDecl := func(recv, name string, decl interface{}) {
		if isUnexported(name) || isUnexported(recv) {
			return
		}
		key := name
		if recv != "" {
			key = recv + "." + name
		}
		if _, ok := decls[key]; ok {
			log.Printf("found multiple declarations of %s in %s", key, pkg.Name)
		}
		decls[key] = decl
	}

	for _, f := range pkg.Files {
		for _, d := range f.Decls {

			switch d := d.(type) {

			case *ast.FuncDecl:
				if d.Recv.NumFields() == 0 {
					addDecl("", d.Name.Name, d)
					continue
				}
				recv := d.Recv.List[0].Type
				if star, ok := recv.(*ast.StarExpr); ok {
					recv = star.X
				}
				if ident, ok := recv.(*ast.Ident); ok {
					addDecl(ident.Name, d.Name.Name, d)
				}

			case *ast.GenDecl:

				for _, s := range d.Specs {
					switch d.Tok {
					case token.CONST, token.VAR:
						vs := s.(*ast.ValueSpec)
						if len(vs.Values) == 0 {
							for _, name := range vs.Names {
								addDecl("", name.Name, vs.Type)
							}
						} else {
							for i, name := range vs.Names {
								addDecl("", name.Name, &valueAndType{vs.Values[i], vs.Type})
							}
						}

					case token.TYPE:
						ts := s.(*ast.TypeSpec)
						addDecl("", ts.Name.Name, ts.Type)

					default:
						// we ignore import declarations
					}
				}

			}
		}
	}
	return decls
}
