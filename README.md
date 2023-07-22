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

can be written in Terraform:

```terraform
data "hashicups_coffees" "test" {}
```

and imported in the provider tests:

```go
func TestAccCoffeesDataSource(t *testing.T) {
    // Each file in test_dir/ will be converted to make one test step in the
    // final test case
    testCase := test.Load(t, "./test_dir/")
    resource.Test(t, testCase)
}
```

## Test configuration

Each test step can be customized with statements in the Terraform configuration.

### Expecting an error

The `ExpectError` keyword will construct test cases that is expected to fail
with an error. The text after `ExpectError:` will be the regexp to look for in
the error message

Example:

```terraform
# ExpectError: error we will look for
resource "dummy_resource" "test" {}
```
