package apicheck

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

// BackwardsCompatible checks whether version target of pkg is backwards
// compatible with the given base version.
// It returns a list of backwards incompatible changes otherwise.
func BackwardsCompatible(repo, base, target string) ([]Change, error) {
	baseAPI, err := FetchAPI(repo, base)
	if err != nil {
		return nil, errors.Wrapf(err, "could not load API for %s version %s", repo, base)
	}

	targetAPI, err := FetchAPI(repo, target)
	if err != nil {
		return nil, errors.Wrapf(err, "could not load API for %s version %s", repo, target)
	}

	return compare(baseAPI, targetAPI)
}

// API contains the result of parsing a whole repository.
type API struct {
	Packages map[string]*ast.Package
}

// FetchAPI fetches the given repository at the specified tag, parses its contents,
// and returns the corresponding API.
func FetchAPI(repo, tag string) (*API, error) {
	dir, err := gitClone(repo, tag)
	// defer os.RemoveAll(dir)
	if err != nil {
		return nil, err
	}

	paths, err := goList(dir)
	if err != nil {
		return nil, errors.Wrapf(err, "could not list packages in %s", dir)
	}

	api := &API{Packages: make(map[string]*ast.Package)}

	base := paths[0]
	for _, path := range paths {
		fs := token.NewFileSet()
		include := func(fi os.FileInfo) bool { return !strings.HasSuffix(fi.Name(), "_test.go") }
		rel := filepath.Join(dir, strings.TrimPrefix(path, base))
		pkgs, err := parser.ParseDir(fs, rel, include, 0)
		if err != nil {
			return nil, errors.Wrapf(err, "could not parse dir %s", dir)
		}

		for name, pkg := range pkgs {
			if strings.HasSuffix(name, "_test") {
				continue
			}
			if _, ok := api.Packages[path]; ok {
				var names []string
				for name := range pkgs {
					names = append(names, name)
				}
				log.Fatalf("found more than one non-test package in %s: %v", path, names)
			}
			api.Packages[path] = pkg
		}
		for _, pkg := range pkgs {
			api.Packages[path] = pkg
		}
	}

	return api, nil
}

func gitClone(repo, tag string) (string, error) {
	path, err := ioutil.TempDir("", "")
	if err != nil {
		return "", errors.Wrapf(err, "could not create temp directory")
	}

	r, err := git.PlainClone(path, false, &git.CloneOptions{URL: repo})
	if err != nil {
		return "", errors.Wrapf(err, "could not clone repository")
	}

	if tag == "HEAD" {
		return path, nil
	}

	tree, err := r.Worktree()
	if err != nil {
		return "", errors.Wrap(err, "could not get worktree")
	}

	err = tree.Checkout(&git.CheckoutOptions{Branch: plumbing.ReferenceName("refs/tags/" + tag)})
	if err != nil {
		return "", errors.Wrapf(err, "could not checkout %s", tag)
	}
	return path, nil
}

func goList(path string) ([]string, error) {
	stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)

	cmd := exec.Command("go", "list", "./...")
	cmd.Dir = path
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		return nil, errors.Wrapf(err, "could not list packages: %s", stderr)
	}

	return strings.Split(strings.TrimSpace(stdout.String()), "\n"), nil
}
