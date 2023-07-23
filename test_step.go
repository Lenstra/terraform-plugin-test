package test

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var (
	expectErrorRe = regexp.MustCompile(`^#\s+ExpectError:\s+(.*)`)
)

func LoadTestStep(path string) (resource.TestStep, error) {
	step := resource.TestStep{}

	data, err := os.ReadFile(path)
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
	step.Check = resource.ComposeAggregateTestCheckFunc(checkFuncs...)

	return step, nil
}
