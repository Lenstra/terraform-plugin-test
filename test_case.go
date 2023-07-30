package test

import (
	"io/fs"
	"path/filepath"
	"sort"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type IgnoreChangeFunc func(name, key, value string) bool

type TestOptions struct {
	IgnoreChange IgnoreChangeFunc
}

func Test(t *testing.T, path string, f func(*testing.T, *resource.TestCase), opts *TestOptions) {
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
				f(t, &testCase)
			}
			resource.Test(t, testCase)
		})
	}
}

func LoadCase(t *testing.T, path string, opts *TestOptions) resource.TestCase {
	c := resource.TestCase{}

	steps, err := LoadTestSteps(path, nil)
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
