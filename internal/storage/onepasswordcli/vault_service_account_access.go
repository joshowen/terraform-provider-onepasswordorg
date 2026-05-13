package onepasswordcli

import (
"context"
"encoding/json"
"fmt"

"github.com/slok/terraform-provider-onepasswordorg/internal/model"
)

func (r *Repository) EnsureVaultServiceAccountAccess(ctx context.Context, saAccess model.VaultServiceAccountAccess) error {
_ = r.DeleteVaultServiceAccountAccess(ctx, saAccess.VaultID, saAccess.ServiceAccountID)

ps := mapModelToOpPermissions(saAccess.Permissions)
cmdArgs := &onePasswordCliCmd{}
cmdArgs.VaultArg().ServiceAccountArg().GrantArg().VaultFlag(saAccess.VaultID).ServiceAccountFlag(saAccess.ServiceAccountID).NoInputFlag().PermissionsFlag(ps)

_, stderr, err := r.cli.RunOpCmd(ctx, cmdArgs.GetArgs())
if err != nil {
return fmt.Errorf("op cli command failed: %w: %s", err, stderr)
}

return nil
}

func (r *Repository) DeleteVaultServiceAccountAccess(ctx context.Context, vaultID string, serviceAccountID string) error {
cmdArgs := &onePasswordCliCmd{}
cmdArgs.VaultArg().ServiceAccountArg().RevokeArg().VaultFlag(vaultID).ServiceAccountFlag(serviceAccountID)

_, stderr, err := r.cli.RunOpCmd(ctx, cmdArgs.GetArgs())
if err != nil {
return fmt.Errorf("op cli command failed: %w: %s", err, stderr)
}

return nil
}

func (r *Repository) GetVaultServiceAccountAccessByID(ctx context.Context, vaultID string, serviceAccountID string) (*model.VaultServiceAccountAccess, error) {
cmdArgs := &onePasswordCliCmd{}
cmdArgs.VaultArg().ServiceAccountArg().ListArg().RawStrArg(vaultID).FormatJSONFlag()

stdout, stderr, err := r.cli.RunOpCmd(ctx, cmdArgs.GetArgs())
if err != nil {
return nil, fmt.Errorf("op cli command failed: %w: %s", err, stderr)
}

accesses := []opVaultServiceAccountAccess{}
err = json.Unmarshal([]byte(stdout), &accesses)
if err != nil {
return nil, fmt.Errorf("could not unmarshal op cli stdout: %w", err)
}

var access *opVaultServiceAccountAccess
for _, a := range accesses {
if a.ServiceAccountID == serviceAccountID {
access = &a
break
}
}

if access == nil {
return nil, fmt.Errorf("service account access %q in vault %q not found", serviceAccountID, vaultID)
}

return &model.VaultServiceAccountAccess{
VaultID:          vaultID,
ServiceAccountID: serviceAccountID,
Permissions:      mapOpToModelPermissions(access.Permissions),
}, nil
}

type opVaultServiceAccountAccess struct {
ServiceAccountID string   `json:"id"`
Permissions      []string `json:"permissions"`
}
