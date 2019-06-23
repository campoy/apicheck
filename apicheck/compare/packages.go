package compare

import (
	"fmt"
	"go/ast"
	"go/types"

	"github.com/campoy/apicheck/apicheck/internal/util"
	"github.com/pkg/errors"
	"golang.org/x/tools/go/packages"
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

func findObj(pkg *packages.Package, name string) types.Object {
	for _, obj := range pkg.TypesInfo.Defs {
		if obj != nil && obj.Name() == name {
			return obj
		}
	}
	return nil
}

// Packages compares two packages and returns the list of changes in its identifiers.
func Packages(base, target *packages.Package) ([]Change, error) {
	getName := func(v interface{}) string { return v.(*ast.Ident).Name }
	isExported := func(v interface{}) bool { return v != nil && v.(types.Object).Exported() }
	names := util.SortUnique(
		util.KeysFromMap(base.TypesInfo.Defs, getName, isExported),
		util.KeysFromMap(target.TypesInfo.Defs, getName, isExported))

	var changes []Change
	for _, name := range names {
		baseObj := findObj(base, name)
		if baseObj == nil {
			changes = append(changes, DeclChange{base.PkgPath, name, true})
			continue
		}

		targetObj := findObj(target, name)
		if targetObj == nil {
			changes = append(changes, DeclChange{target.PkgPath, name, false})
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
