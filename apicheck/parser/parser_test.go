package parser_test

import (
	"sort"
	"testing"

	"github.com/campoy/apicheck/apicheck/parser"
)

func TestParse(t *testing.T) {
	repo, err := parser.ParseRepo(".")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(repo.Packages) != 1 {
		t.Fatalf("expected one package, got %v", len(repo.Packages))
	}
	const this = "github.com/campoy/apicheck/apicheck/parser"
	pkg, ok := repo.Packages[this]
	if !ok {
		t.Fatalf("expected to find package %s, got %v", this, repo.Packages)
	}

	var names []string
	for _, obj := range pkg.Defs {
		if obj != nil && obj.Exported() {
			names = append(names, obj.Name())
		}
	}
	sort.Strings(names)

	if len(names) != 6 {
		t.Fatalf("expected 6 exported identifiers from this package, got %v", len(names))
	}

	expects := []string{
		"CloneAndParse",
		"CloneRepo",
		"ListPackages",
		"Packages",
		"ParseRepo",
		"Repo",
	}

	for i, name := range names {
		if name != expects[i] {
			t.Errorf("expected %d-th name to be %s, got %s", i, expects[i], name)
		}
	}
}
