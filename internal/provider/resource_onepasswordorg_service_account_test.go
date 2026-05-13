package provider_test

import (
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/slok/terraform-provider-onepasswordorg/internal/model"
	"github.com/slok/terraform-provider-onepasswordorg/internal/provider"
)

// TestAccServiceAccountCreateDelete checks that a service account is created and deleted.
func TestAccServiceAccountCreateDelete(t *testing.T) {
	tests := map[string]struct {
		config string
		expSA  model.ServiceAccount
		expErr *regexp.Regexp
	}{
		"A correct configuration should create the service account.": {
			config: `
resource "onepasswordorg_service_account" "test_sa" {
  name = "ci-bot"
}
`,
			expSA: model.ServiceAccount{
				ID:    "ci-bot",
				Name:  "ci-bot",
				Token: "ops_fake_token_ci-bot",
			},
		},

		"A non-set name should fail.": {
			config: `
resource "onepasswordorg_service_account" "test_sa" {
}
`,
			expErr: regexp.MustCompile("Missing required argument"),
		},

		"An empty name should fail.": {
			config: `
resource "onepasswordorg_service_account" "test_sa" {
  name = ""
}
`,
			expErr: regexp.MustCompile("Attribute name string length must be at least 1, got: 0"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// Prepare fake storage.
			path, deleteFn := getFakeRepoTmpFile("TestAccServiceAccountCreateDelete")
			defer deleteFn()
			_ = os.Setenv(provider.EnvVarOpFakeStoragePath, path)

			var checks resource.TestCheckFunc
			if test.expErr == nil {
				checks = resource.ComposeAggregateTestCheckFunc(
					assertServiceAccountOnFakeStorage(t, &test.expSA),
					resource.TestCheckResourceAttr("onepasswordorg_service_account.test_sa", "id", test.expSA.ID),
					resource.TestCheckResourceAttr("onepasswordorg_service_account.test_sa", "name", test.expSA.Name),
					// Token is sensitive; we can still check it exists in state via the helper.
				)
			}

			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { testAccPreCheck(t) },
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				CheckDestroy:             assertServiceAccountDeletedOnFakeStorage(t, test.expSA.ID),
				Steps: []resource.TestStep{
					{
						Config:      test.config,
						Check:       checks,
						ExpectError: test.expErr,
					},
				},
			})
		})
	}
}

// TestAccServiceAccountNameChangeReplace checks that changing the name recreates the resource.
func TestAccServiceAccountNameChangeReplace(t *testing.T) {
	// Prepare fake storage.
	path, deleteFn := getFakeRepoTmpFile("TestAccServiceAccountNameChangeReplace")
	defer deleteFn()
	_ = os.Setenv(provider.EnvVarOpFakeStoragePath, path)

	configCreate := `
resource "onepasswordorg_service_account" "test_sa" {
  name = "ci-bot"
}
`
	configUpdate := `
resource "onepasswordorg_service_account" "test_sa" {
  name = "ci-bot-renamed"
}
`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: configCreate,
				Check: resource.ComposeAggregateTestCheckFunc(
					assertServiceAccountOnFakeStorage(t, &model.ServiceAccount{
						ID:    "ci-bot",
						Name:  "ci-bot",
						Token: "ops_fake_token_ci-bot",
					}),
					resource.TestCheckResourceAttr("onepasswordorg_service_account.test_sa", "name", "ci-bot"),
				),
			},
			{
				Config: configUpdate,
				Check: resource.ComposeAggregateTestCheckFunc(
					assertServiceAccountOnFakeStorage(t, &model.ServiceAccount{
						ID:    "ci-bot-renamed",
						Name:  "ci-bot-renamed",
						Token: "ops_fake_token_ci-bot-renamed",
					}),
					assertServiceAccountDeletedOnFakeStorage(t, "ci-bot"),
					resource.TestCheckResourceAttr("onepasswordorg_service_account.test_sa", "name", "ci-bot-renamed"),
				),
			},
		},
	})
}
