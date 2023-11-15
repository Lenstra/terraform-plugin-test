# terraform-plugin-test

`terraform-plugin-test` simplifies writing tests for Terraform plugin provider.

Tests that must be written as Go code like this:

```go
func TestAccCoffeesDataSource(t *testing.T) {
    resource.Test(t, resource.TestCase{
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            // Read testing
            {
                Config: providerConfig + `data "hashicups_coffees" "test" {}`,
                Check: resource.ComposeAggregateTestCheckFunc(
                    // Verify number of coffees returned
                    resource.TestCheckResourceAttr("data.hashicups_coffees.test", "coffees.#", "9"),
                    // Verify the first coffee to ensure all attributes are set
                    resource.TestCheckResourceAttr("data.hashicups_coffees.test", "coffees.0.description", ""),
                    resource.TestCheckResourceAttr("data.hashicups_coffees.test", "coffees.0.id", "1"),
                    resource.TestCheckResourceAttr("data.hashicups_coffees.test", "coffees.0.image", "/hashicorp.png"),
                    resource.TestCheckResourceAttr("data.hashicups_coffees.test", "coffees.0.ingredients.#", "1"),
                    resource.TestCheckResourceAttr("data.hashicups_coffees.test", "coffees.0.ingredients.0.id", "6"),
                    resource.TestCheckResourceAttr("data.hashicups_coffees.test", "coffees.0.name", "HCP Aeropress"),
                    resource.TestCheckResourceAttr("data.hashicups_coffees.test", "coffees.0.price", "200"),
                    resource.TestCheckResourceAttr("data.hashicups_coffees.test", "coffees.0.teaser", "Automation in a cup"),
                    // Verify placeholder id attribute
                    resource.TestCheckResourceAttr("data.hashicups_coffees.test", "id", "placeholder"),
                ),
            },
        },
    })
}
```

can now be written in Terraform:

```terraform
data "hashicups_coffees" "test" {}
```

and imported in the provider tests:

```go
func TestAccCoffeesDataSource(t *testing.T) {
    // Each file in test_dir/ will be converted to make one test step in the
    // final test case
    test.Test(t, "./test_dir/", func(t *testing.T, dir string, tc *resource.TestCase) {
		tc.PreCheck = func() { testAccPreCheck(t) }
		tc.ProtoV6ProviderFactories = testAccProtoV6ProviderFactories
	}, nil)
}
```

## Test cases

The `terraform-plugin-test` library expect each test case to be in a separate
folder. Each `*.tf` file in the folders will be interpreted as a single step
in the test case. Here's an example file structure:

```
tests/
├── hashicups_coffees/
│   ├── missing-argument.tf
│   ├── datasource.json
│   └── datasource.tf
└── hashicups_toppings/
    ├── not-found.tf
    ├── new-topping.tf
    └── new-topping.json
```

## Test configuration

Each test step can be customized with statements in the Terraform configuration.
At least one `Check` or `ExpectError` statement must be present.

### Checking the resulting state

The `Check` keyword will automatically verify that the Terraform state after
applying the configuration matches the information present in the associated
JSON file. The text after `Check:` is the address of the resource whose state
must be check. This keyword can be used multiple time to check multiple resources
in a single test step.

Example:

```terraform
# Check: dummy_resource.test
resource "dummy_resource" "test" {}

# Check: data.hashicups_toppings.test
data "hashicups_toppings" "test" {}
```

The JSON files can be refreshed automatically using the `TFTEST_REFRESH_STATE`
environment variable.

### Expecting an error

The `ExpectError` keyword will construct test cases that is expected to fail
with an error. The text after `ExpectError:` will be the regexp to look for in
the error message

Example:

```terraform
# ExpectError: no dummy_resource could be found
resource "dummy_resource" "test" {}
```

## Automatic refresh of the Terraform state files

The Terraform states in the JSON files can be automatically refreshed by
setting the `TFTEST_REFRESH_STATE` environment variable to a non-empty value
when running the tests. This makes it convenient to update the tests or write
a new one:

```shell-session
$ TF_ACC=1 TFTEST_REFRESH_STATE=1 go test
```

All the JSON files updated during this step can be reviewed manually before
committing them to version control.

## Functions

### func [DefaultIgnoreChangeFunc](/test_case.go#L19)

`func DefaultIgnoreChangeFunc(name, key, value string) bool`

DefaultIgnoreChangeFunc is the default IgnoreChangeFunc that will be used if
one is not given by the user. It will ignore any attribute that could be
an UUID or a time string.

### func [LoadCase](/test_case.go#L75)

`func LoadCase(t *testing.T, path string, opts *TestOptions) resource.TestCase`

LoadCase loads a resource.TestCase from the given folder path.

### func [Test](/test_case.go#L50)

`func Test(t *testing.T, path string, f func(*testing.T, string, *resource.TestCase), opts *TestOptions)`

Test is the main entrypoint of terraform-plugin-test. The user can
specify a function f to customize the TestCases before they are run and
optionaly set TestOptions to control how the attributes are compared to the
expected state file.
