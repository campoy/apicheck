package compare

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"log"
	"unicode"
	"unicode/utf8"

	"github.com/campoy/apicheck/apicheck/internal/util"
	"github.com/pkg/errors"
	"golang.org/x/tools/go/loader"
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

func findObj(pkg *loader.PackageInfo, name string) types.Object {
	for _, obj := range pkg.Defs {
		if obj != nil && obj.Name() == name {
			return obj
		}
	}
	return nil
}

// Packages compares two packages and returns the list of changes in its identifiers.
func Packages(base, target *loader.PackageInfo) ([]Change, error) {
	getName := func(v interface{}) string { return v.(*ast.Ident).Name }
	isExported := func(v interface{}) bool { return v != nil && v.(types.Object).Exported() }
	names := util.SortUnique(
		util.KeysFromMap(base.Defs, getName, isExported),
		util.KeysFromMap(target.Defs, getName, isExported))

	var changes []Change
	for _, name := range names {
		baseObj := findObj(base, name)
		if baseObj == nil {
			changes = append(changes, DeclChange{base.Pkg.Path(), name, true})
			continue
		}

		targetObj := findObj(target, name)
		if targetObj == nil {
			changes = append(changes, DeclChange{target.Pkg.Path(), name, false})
			continue
		}

		cs, err := Decls(name, baseObj, targetObj)
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
