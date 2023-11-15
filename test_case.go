package test

import (
	"io/fs"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type IgnoreChangeFunc func(name, key, value string) bool

// DefaultIgnoreChangeFunc is the default IgnoreChangeFunc that will be used if
// one is not given by the user. It will ignore any attribute that could be
// an UUID or a time string.
func DefaultIgnoreChangeFunc(name, key, value string) bool {
	if _, err := uuid.ParseUUID(value); err == nil {
		return true
	}

	layouts := []string{
		time.RFC3339,
		time.RFC3339Nano,
		time.DateTime,
		time.DateOnly,
		time.TimeOnly,
		"2006-01-02 15:04:05.999999999 -0700 MST",
	}
	for _, layout := range layouts {
		if _, err := time.Parse(layout, value); err == nil {
			return true
		}
	}

	return false
}

// TestOptions can be used to customize the behavior of terraform-plugin-test.
type TestOptions struct {
	IgnoreChange IgnoreChangeFunc
}

// Test is the main entrypoint of terraform-plugin-test. The user can
// specify a function f to customize the TestCases before they are run and
// optionaly set TestOptions to control how the attributes are compared to the
// expected state file.
func Test(t *testing.T, path string, f func(*testing.T, string, *resource.TestCase), opts *TestOptions) {
	files := find(path)
	dirs := map[string]struct{}{}
	for _, path := range files {
		dirs[filepath.Dir(path)] = struct{}{}
	}
	files = []string{}
	for dir := range dirs {
		files = append(files, dir)
	}

	sort.Strings(files)

	for _, dir := range files {
		t.Run(dir, func(t *testing.T) {
			testCase := LoadCase(t, dir, opts)
			if f != nil {
				f(t, dir, &testCase)
			}
			resource.Test(t, testCase)
		})
	}
}

// LoadCase loads a resource.TestCase from the given folder path.
func LoadCase(t *testing.T, path string, opts *TestOptions) resource.TestCase {
	c := resource.TestCase{}

	steps, err := loadTestSteps(path, opts)
	if err != nil {
		t.Fatal(err.Error())
	}
	c.Steps = steps

	return c
}

func find(root string) []string {
	var a []string
	filepath.WalkDir(root, func(s string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if filepath.Ext(d.Name()) == ".tf" {
			a = append(a, s)
		}
		return nil
	})
	return a
}
