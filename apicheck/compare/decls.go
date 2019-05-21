package compare

import (
	"fmt"
	"go/ast"
	"log"
	"reflect"
)

// IdentChange tracks changes in the definition of an identifier.
type IdentChange struct {
	Name                 string
	BaseType, TargetType reflect.Type
}

func (c IdentChange) String() string {
	return fmt.Sprintf("%s changed from %v to %v", c.Name, c.BaseType, c.TargetType)
}

// Compatible returns whether the change is backwards compatible.
func (c IdentChange) Compatible() bool {
	// TODO: some changes might be compatible. e.g. const to var
	return false
}

// Decls compares two different declarations.
func Decls(name string, base, target interface{}) ([]Change, error) {
	if bt, tt := reflect.TypeOf(base), reflect.TypeOf(target); bt != tt {
		return []Change{IdentChange{name, bt, tt}}, nil
	}

	switch b := base.(type) {
	case *ast.Ident:
		return Idents(name, b, target.(*ast.Ident))
	case *ast.FuncDecl:
		return FuncDecls(name, b, target.(*ast.FuncDecl))
	case *ast.FuncType:
		return FuncTypes(name, b, target.(*ast.FuncType))
	case *ast.InterfaceType:
		return InterfaceTypes(name, b, target.(*ast.InterfaceType))
	case *ast.StarExpr:
		return StarExprs(name, b, target.(*ast.StarExpr))
	case *ast.StructType:
		return StructTypes(name, b, target.(*ast.StructType))
	case *valueAndType:
		return valueAndTypes(name, b, target.(*valueAndType))
	default:
		log.Printf("can't handle %T yet", b)
	}

	return nil, nil
}

// Idents compares two Idents and returns all found changes.
func Idents(name string, base, target *ast.Ident) ([]Change, error) {
	return nil, nil
}

// FuncDecls compares two FuncDecls and returns all found changes.
func FuncDecls(name string, base, target *ast.FuncDecl) ([]Change, error) {
	return nil, nil
}

// FuncTypes compares two FuncTypes and returns all found changes.
func FuncTypes(name string, base, target *ast.FuncType) ([]Change, error) {
	return nil, nil
}

// InterfaceTypes compares two InterfaceTypes and returns all found changes.
func InterfaceTypes(name string, base, target *ast.InterfaceType) ([]Change, error) {
	return nil, nil
}

// StarExprs compares two StarExprs and returns all found changes.
func StarExprs(name string, base, target *ast.StarExpr) ([]Change, error) {
	return nil, nil
}

// StructTypes compares two StructTypes and returns all found changes.
func StructTypes(name string, base, target *ast.StructType) ([]Change, error) {
	return nil, nil
}

// valueAndTypes compares two valueAndTypes and returns all found changes.
func valueAndTypes(name string, base, target *valueAndType) ([]Change, error) {
	return nil, nil
}
