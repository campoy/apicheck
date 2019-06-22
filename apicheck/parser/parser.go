package parser

import (
	"bytes"
	"go/types"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/tools/go/loader"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

// Repo contains the result of parsing a whole repository.
type Repo struct {
	Base     string
	Packages map[string]*loader.PackageInfo
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
	paths, err := ListPackages(dir)
	if err != nil {
		return nil, errors.Wrapf(err, "could not list packages in %s", dir)
	}

	api := &Repo{Base: paths[0], Packages: make(map[string]*loader.PackageInfo)}
	conf := loader.Config{
		AllowErrors:         true,
		TypeChecker:         types.Config{Error: func(error) {}},
		TypeCheckFuncBodies: func(path string) bool { return strings.HasPrefix(path, api.Base) },
	}
	for _, path := range paths {
		conf.Import(path)
	}

	prog, err := conf.Load()
	if err != nil {
		log.Fatal(err)
	}

	for _, path := range paths {
		pkg := prog.Package(path)
		if pkg == nil {
			return nil, errors.Errorf("could not find %s path", err)
		}
		api.Packages[strings.TrimPrefix(path, api.Base)] = pkg
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

// ListPackages returns the list of Go packages under the given directory.
func ListPackages(path string) ([]string, error) {
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
