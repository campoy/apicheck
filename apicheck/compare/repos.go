package compare

import (
	"fmt"

	"github.com/campoy/apicheck/apicheck/internal/util"
	"github.com/campoy/apicheck/apicheck/parser"
	"github.com/pkg/errors"
)

// A Change is the basic abstraction of all changes, backwards compatible or not.
type Change interface {
	String() string
	Compatible() bool
}

// packageChange tracks the addition or removal of a package in a repository.
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

// Repos compares two given Repos.
func Repos(base, target *parser.Repo) ([]Change, error) {
	paths := util.SortUnique(
		util.KeysFromMap(base.Packages, nil, nil),
		util.KeysFromMap(target.Packages, nil, nil))

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
		cs, err := Packages(base.Packages[path], target.Packages[path])
		if err != nil {
			return nil, errors.Wrapf(err, "could not compare %s", path)
		}
		changes = append(changes, cs...)
	}

	return changes, nil
}
