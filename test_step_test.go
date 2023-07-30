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
		expectError         string
		expectCheckFunction bool
	}{
		{
			path:                "tests/test-case/test.tf",
			expectCheckFunction: true,
			want: resource.TestStep{
				Config:      "# ExpectError: error we will look for\nresource \"dummy_resource\" \"test\" {}\n",
				ExpectError: regexp.MustCompile("error we will look for"),
			},
		},
		{
			path: "tests/test-case/missing-state.tf",
			want: resource.TestStep{
				Config: "# Check: dummy_resource.test\nresource \"dummy_resource\" \"test\" {}\n",
			},
		},
		{
			path:        "tests/missing-comment.tf",
			expectError: "neither Check or ExpectError statements have been found",
			want:        resource.TestStep{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got, err := loadTestStep(tt.path, nil)

			if tt.expectError == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tt.expectError)
			}

			if tt.expectCheckFunction {
				require.NotNil(t, got.Check)
				got.Check = nil
			} else {
				require.Nil(t, got.Check)
			}

			if tt.expectError == "" {
				require.Equal(t, got, tt.want)
			}
		})
	}
}
