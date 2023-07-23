package test

import (
	"io/fs"
	"path/filepath"
	"sort"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func Test(t *testing.T, path string, f func(*testing.T, *resource.TestCase)) {
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
			testCase := LoadCase(t, dir)
			if f != nil {
				f(t, &testCase)
			}
			resource.Test(t, testCase)
		})
	}
}

func LoadCase(t *testing.T, path string) resource.TestCase {
	c := resource.TestCase{}

	files := find(path)
	sort.Strings(files)

	for _, filename := range files {
		step, err := LoadTestStep(filename)
		if err != nil {
			t.Fatalf("failed to load step: %v", err)
		}
		c.Steps = append(c.Steps, step)
	}

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
