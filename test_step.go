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

const (
	refreshStateEnv = "TFTEST_REFRESH_STATE"
)

var (
	expectErrorRe = regexp.MustCompile(`^#\s*ExpectError:\s+(.*)`)
	checkRe       = regexp.MustCompile(`^#\s*Check:\s+(.*)`)
)

func LoadTestSteps(path string, opts *TestOptions) ([]resource.TestStep, error) {
	var steps []resource.TestStep

	files := find(path)
	sort.Strings(files)

	for _, filename := range files {
		step, err := loadTestStep(filename, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to load step: %w", err)
		}
		steps = append(steps, step)
	}

	return steps, nil
}

func loadTestStep(path string, opts *TestOptions) (resource.TestStep, error) {
	if opts == nil {
		opts = &TestOptions{
			IgnoreChange: DefaultIgnoreChangeFunc,
		}
	}

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

	names := map[string]struct{}{}
	for _, token := range tokens {
		if token.Type == hclsyntax.TokenComment {
			if expectErrorRe.Match(token.Bytes) {
				matches := expectErrorRe.FindSubmatch(token.Bytes)
				if len(matches) != 0 {
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

			if checkRe.Match(token.Bytes) {
				matches := checkRe.FindSubmatch(token.Bytes)
				if len(matches) != 0 {
					name := strings.TrimSpace(string(matches[1]))
					names[name] = struct{}{}
				}
			}
		}
	}

	if step.ExpectError == nil && len(names) == 0 {
		return step, errors.New("neither Check or ExpectError statements have been found in Terraform configuration")
	}

	stateFilePath := strings.TrimSuffix(path, filepath.Ext(path)) + ".json"

	if os.Getenv(refreshStateEnv) != "" {
		step.Check = refreshStateFunc(stateFilePath, names, opts.IgnoreChange)
		return step, nil
	}

	data, err = os.ReadFile(stateFilePath)
	if errors.Is(err, os.ErrNotExist) {
		return step, nil
	} else if err != nil {
		return step, err
	}

	var state map[string]map[string]string
	if err = json.Unmarshal(data, &state); err != nil {
		return step, err
	}

	var checkFuncs []resource.TestCheckFunc
	for name, content := range state {
		for key, value := range content {
			if _, found := names[name]; !found {
				continue
			}

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

func refreshStateFunc(path string, names map[string]struct{}, ignoreChange IgnoreChangeFunc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		state := map[string]map[string]string{}

		for name := range names {
			ms := s.RootModule()
			rs, ok := ms.Resources[name]
			if !ok {
				return fmt.Errorf("not found: %s in %s", name, ms.Path)
			}

			is := rs.Primary
			if is == nil {
				return fmt.Errorf("no primary instance: %s in %s", name, ms.Path)
			}

			state[name] = map[string]string{}
			for key, value := range is.Attributes {
				if key == "%" {
					continue
				}

				if ignoreChange != nil && ignoreChange(name, key, value) {
					state[name][key] = "<set>"
				} else {
					state[name][key] = value
				}
			}
		}

		f, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("failed to open file %q: %w", path, err)
		}
		defer f.Close()

		encoder := json.NewEncoder(f)
		encoder.SetEscapeHTML(false)
		encoder.SetIndent("", "  ")
		err = encoder.Encode(state)
		if err != nil {
			return fmt.Errorf("failed to marshal state: %w", err)
		}

		return nil
	}
}
