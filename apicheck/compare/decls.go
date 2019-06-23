package compare

import (
	"fmt"
	"go/types"
)

// IdentChange tracks changes in the definition of an identifier.
type IdentChange struct {
	Name                 string
	BaseType, TargetType types.Type
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
func Decls(name string, base, target types.Object) ([]Change, error) {
	if types.Identical(base.Type(), target.Type()) {
		return nil, nil
	}
	if types.AssignableTo(base.Type(), target.Type()) {
		return nil, nil
	}
	if types.ConvertibleTo(base.Type(), target.Type()) {
		return nil, nil
	}
	// log.Printf("comparing %s:\n\t%s\n\tvs.\nt\t%s\n\n", name, base, target)
	return []Change{IdentChange{name, base.Type(), target.Type()}}, nil
}
