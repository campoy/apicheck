package parser

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/tools/go/packages"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

// Repo contains the result of parsing a whole repository.
type Repo struct {
	Base     string
	Packages map[string]*packages.Package
}

// CloneAndParse clones the a repo into a temporary directory,
// checks out the given tag, and parses its contents.
// The temporary directory is removed automatically.
func CloneAndParse(repo, tag string) (*Repo, error) {
	dir, err := CloneRepo(repo, tag)
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)
	return ParseRepo(dir)
}

// ParseRepo fetches the given repository at the specified tag, parses its contents,
// and returns the corresponding Repo.
func ParseRepo(dir string) (*Repo, error) {
	basePkg, err := packages.Load(&packages.Config{Dir: dir}, ".")
	if err != nil {
		return nil, errors.Wrapf(err, "could not load base package")
	}
	api := &Repo{
		Base:     basePkg[0].PkgPath,
		Packages: make(map[string]*packages.Package),
	}

	pkgs, err := packages.Load(&packages.Config{Dir: dir, Mode: packages.LoadSyntax}, "./...")
	if err != nil {
		return nil, err
	}
	if len(pkgs) == 0 {
		return nil, errors.Errorf("no packages found in %q", dir)
	}

	for _, pkg := range pkgs {
		api.Packages[strings.TrimPrefix(pkg.PkgPath, api.Base)] = pkg
	}

	return api, nil
}

// CloneRepo clones the given git repository and checks out the given tag
// into a temporary directory whose path is returned.
// The caller is reponsible for deleting the temp directory if needed.
func CloneRepo(repo, tag string) (string, error) {
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
