package test

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var (
	expectErrorRe = regexp.MustCompile(`^#\s+ExpectError:\s+(.*)`)
)

func LoadTestSteps(path string) ([]resource.TestStep, error) {
	var steps []resource.TestStep

	files := find(path)
	sort.Strings(files)

	for _, filename := range files {
		step, err := LoadTestStep(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to load step: %w", err)
		}
		steps = append(steps, step)
	}

	return steps, nil
}

func LoadTestStep(path string) (resource.TestStep, error) {
	step := resource.TestStep{}

	data, err := os.ReadFile(path)
	if err != nil {
		return step, err
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return step, err
	}

	step.Config = string(data)

	tokens, diags := hclsyntax.LexConfig(data, path, hcl.InitialPos)
	if diags.HasErrors() {
		return step, diags
	}

	for _, token := range tokens {
		if token.Type == hclsyntax.TokenComment {
			if expectErrorRe.Match(token.Bytes) {
				matches := expectErrorRe.FindSubmatch(token.Bytes)
				if len(matches) == 0 {
					continue
				}
				if step.ExpectError != nil {
					return step, errors.New("multiple ExpectError statements have been found")
				}
				expr := strings.TrimSpace(string(matches[1]))
				re, err := regexp.Compile(expr)
				if err != nil {
					return step, err
				}
				step.ExpectError = re
			}
		}
	}

	path = strings.TrimSuffix(path, filepath.Ext(path)) + ".json"
	data, err = os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return step, nil
	}

	if err != nil {
		return step, err
	}

	var state map[string]map[string]string
	if err = json.Unmarshal(data, &state); err != nil {
		return step, err
	}

	var checkFuncs []resource.TestCheckFunc
	for name, content := range state {
		for key, value := range content {
			if value == "<set>" {
				checkFuncs = append(checkFuncs, resource.TestCheckResourceAttrSet(name, key))
			} else {
				checkFuncs = append(checkFuncs, resource.TestCheckResourceAttr(name, key, value))
			}
		}
	}

	step.Check = func(state *terraform.State) error {
		err := resource.ComposeAggregateTestCheckFunc(checkFuncs...)(state)
		if err != nil {
			return fmt.Errorf("%w\nAn error occured while running test step %q\n\n", err, absPath)
		}

		return nil
	}

	return step, nil
}
