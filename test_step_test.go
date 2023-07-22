package test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestLoadTestStep(t *testing.T) {
	tests := []struct {
		path                string
		want                resource.TestStep
		expectCheckFunction bool
	}{
		{
			path:                "tests/test.tf",
			expectCheckFunction: true,
			want: resource.TestStep{
				Config:      "# ExpectError: error we will look for\nresource \"dummy_resource\" \"test\" {}\n",
				ExpectError: regexp.MustCompile("error we will look for"),
			},
		},
		{
			path: "tests/missing-state.tf",
			want: resource.TestStep{
				Config: "resource \"dummy_resource\" \"test\" {}\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got, err := LoadTestStep(tt.path)
			require.NoError(t, err)
			if tt.expectCheckFunction {
				require.NotNil(t, got.Check)
				got.Check = nil
			} else {
				require.Nil(t, got.Check)
			}
			require.Equal(t, got, tt.want)
		})
	}
}
