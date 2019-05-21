package apicheck

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

// A Change is the basic abstraction of all changes, backwards compatible or not.
type Change interface {
	String() string
	Compatible() bool
}

type packageChange struct {
	path  string
	added bool
}

func (c packageChange) String() string {
	if c.added {
		return fmt.Sprintf("package %q added", c.path)
	}
	return fmt.Sprintf("package %q removed", c.path)
}

func (c packageChange) Compatible() bool { return c.added }

func compare(base, target *API) ([]Change, error) {
	pathsMap := make(map[string]bool)
	for path := range base.Packages {
		pathsMap[path] = true
	}
	for path := range target.Packages {
		pathsMap[path] = true
	}
	paths := make([]string, 0, len(pathsMap))
	for path := range pathsMap {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	var changes []Change

	for _, path := range paths {
		if _, ok := base.Packages[path]; !ok {
			changes = append(changes, packageChange{path, true})
			continue
		}
		if _, ok := target.Packages[path]; !ok {
			changes = append(changes, packageChange{path, false})
			continue
		}

		// the package appears in both sides
		cs, err := comparePkgs(path, base.Packages[path], target.Packages[path])
		if err != nil {
			return nil, errors.Wrapf(err, "could not compare %s", path)
		}
		changes = append(changes, cs...)
	}

	return changes, nil
}

type declChange struct {
	path  string
	name  string
	added bool
}

func (c declChange) String() string {
	name := c.path + "." + c.name
	if c.added {
		return fmt.Sprintf("identifier %q added", name)
	}
	return fmt.Sprintf("identifier %q removed", name)
}

func (c declChange) Compatible() bool { return c.added }

func comparePkgs(importPath string, base, target *ast.Package) ([]Change, error) {
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
			changes = append(changes, declChange{importPath, name, true})
			continue
		}
		if _, ok := targetDecls[name]; !ok {
			changes = append(changes, declChange{importPath, name, false})
			continue
		}

		cs, err := compareDecls(baseDecls[name], targetDecls[name])
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
			log.Printf("found %s declared twice in %s", key, pkg.Name)
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
						for _, name := range vs.Names {
							addDecl("", name.Name, vs.Type)
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

func compareDecls(target, base interface{}) ([]Change, error) {
	return nil, nil
}
