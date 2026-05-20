package onepasswordcli_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/slok/terraform-provider-onepasswordorg/internal/model"
	"github.com/slok/terraform-provider-onepasswordorg/internal/storage/onepasswordcli"
	"github.com/slok/terraform-provider-onepasswordorg/internal/storage/onepasswordcli/onepasswordclimock"
)

func TestRepositoryEnsureVaultServiceAccountAccess(t *testing.T) {
	tests := map[string]struct {
		access model.VaultServiceAccountAccess
		mock   func(m *onepasswordclimock.OpCli)
		expErr bool
	}{
		"Granting service account access to a vault should issue op vault user grant.": {
			access: model.VaultServiceAccountAccess{
				VaultID:          "vault-001",
				ServiceAccountID: "sa-001",
				Permissions: model.AccessPermissions{
					ViewItems:            true,
					ViewAndCopyPasswords: true,
					ViewItemHistory:      true,
				},
			},
			mock: func(m *onepasswordclimock.OpCli) {
				// First revokes everything.
				expCmd := `vault user revoke --vault vault-001 --user sa-001`
				m.On("RunOpCmd", mock.Anything, strings.Fields(expCmd)).Once().Return("", "", nil)

				// Then grants the specified permissions.
				expCmd = `vault user grant --vault vault-001 --user sa-001 --no-input --permissions view_items,view_and_copy_passwords,view_item_history`
				m.On("RunOpCmd", mock.Anything, strings.Fields(expCmd)).Once().Return("", "", nil)
			},
		},

		"Having an error while calling op vault user grant should fail.": {
			access: model.VaultServiceAccountAccess{
				VaultID:          "vault-001",
				ServiceAccountID: "sa-001",
			},
			mock: func(m *onepasswordclimock.OpCli) {
				expCmd := `vault user revoke --vault vault-001 --user sa-001`
				m.On("RunOpCmd", mock.Anything, strings.Fields(expCmd)).Once().Return("", "", nil)

				expCmd = `vault user grant --vault vault-001 --user sa-001 --no-input`
				m.On("RunOpCmd", mock.Anything, strings.Fields(expCmd)).Once().Return("", "", fmt.Errorf("something"))
			},
			expErr: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			require := require.New(t)
			assert := assert.New(t)

			mc := &onepasswordclimock.OpCli{}
			test.mock(mc)

			repo, err := onepasswordcli.NewRepository(mc)
			require.NoError(err)

			err = repo.EnsureVaultServiceAccountAccess(context.TODO(), test.access)

			if test.expErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}

			mc.AssertExpectations(t)
		})
	}
}

func TestRepositoryDeleteVaultServiceAccountAccess(t *testing.T) {
	tests := map[string]struct {
		vaultID          string
		serviceAccountID string
		mock             func(m *onepasswordclimock.OpCli)
		expErr           bool
	}{
		"Revoking service account vault access should issue op vault user revoke.": {
			vaultID:          "vault-001",
			serviceAccountID: "sa-001",
			mock: func(m *onepasswordclimock.OpCli) {
				expCmd := `vault user revoke --vault vault-001 --user sa-001`
				m.On("RunOpCmd", mock.Anything, strings.Fields(expCmd)).Once().Return("", "", nil)
			},
		},

		"Having an error while calling op vault user revoke should fail.": {
			vaultID:          "vault-001",
			serviceAccountID: "sa-001",
			mock: func(m *onepasswordclimock.OpCli) {
				expCmd := `vault user revoke --vault vault-001 --user sa-001`
				m.On("RunOpCmd", mock.Anything, strings.Fields(expCmd)).Once().Return("", "", fmt.Errorf("something"))
			},
			expErr: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			require := require.New(t)
			assert := assert.New(t)

			mc := &onepasswordclimock.OpCli{}
			test.mock(mc)

			repo, err := onepasswordcli.NewRepository(mc)
			require.NoError(err)

			err = repo.DeleteVaultServiceAccountAccess(context.TODO(), test.vaultID, test.serviceAccountID)

			if test.expErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}

			mc.AssertExpectations(t)
		})
	}
}

func TestRepositoryGetVaultServiceAccountAccessByID(t *testing.T) {
	tests := map[string]struct {
		vaultID          string
		serviceAccountID string
		mock             func(m *onepasswordclimock.OpCli)
		expAccess        *model.VaultServiceAccountAccess
		expErr           bool
	}{
		"Getting an existing service account access should return its permissions.": {
			vaultID:          "vault-001",
			serviceAccountID: "sa-001",
			mock: func(m *onepasswordclimock.OpCli) {
				expCmd := `vault user list vault-001 --format json`
				accesses := []map[string]interface{}{
					{"id": "sa-other", "permissions": []string{"view_items"}},
					{"id": "sa-001", "permissions": []string{"view_items", "view_and_copy_passwords", "manage_vault"}},
				}
				out, _ := json.Marshal(accesses)
				m.On("RunOpCmd", mock.Anything, strings.Fields(expCmd)).Once().Return(string(out), "", nil)
			},
			expAccess: &model.VaultServiceAccountAccess{
				VaultID:          "vault-001",
				ServiceAccountID: "sa-001",
				Permissions: model.AccessPermissions{
					ViewItems:            true,
					ViewAndCopyPasswords: true,
					ManageVault:          true,
				},
			},
		},

		"When the service account is not in the vault user list, should return an error.": {
			vaultID:          "vault-001",
			serviceAccountID: "sa-missing",
			mock: func(m *onepasswordclimock.OpCli) {
				expCmd := `vault user list vault-001 --format json`
				accesses := []map[string]interface{}{
					{"id": "sa-other", "permissions": []string{"view_items"}},
				}
				out, _ := json.Marshal(accesses)
				m.On("RunOpCmd", mock.Anything, strings.Fields(expCmd)).Once().Return(string(out), "", nil)
			},
			expErr: true,
		},

		"Having an error while calling op vault user list should fail.": {
			vaultID:          "vault-001",
			serviceAccountID: "sa-001",
			mock: func(m *onepasswordclimock.OpCli) {
				expCmd := `vault user list vault-001 --format json`
				m.On("RunOpCmd", mock.Anything, strings.Fields(expCmd)).Once().Return("", "", fmt.Errorf("something"))
			},
			expErr: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			require := require.New(t)
			assert := assert.New(t)

			mc := &onepasswordclimock.OpCli{}
			test.mock(mc)

			repo, err := onepasswordcli.NewRepository(mc)
			require.NoError(err)

			gotAccess, err := repo.GetVaultServiceAccountAccessByID(context.TODO(), test.vaultID, test.serviceAccountID)

			if test.expErr {
				assert.Error(err)
			} else {
				require.NoError(err)
				assert.Equal(test.expAccess, gotAccess)
			}

			mc.AssertExpectations(t)
		})
	}
}
