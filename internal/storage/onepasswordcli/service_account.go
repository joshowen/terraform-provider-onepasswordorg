package onepasswordcli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/slok/terraform-provider-onepasswordorg/internal/model"
)

// opServiceAccount is the JSON representation returned by the op CLI for service accounts.
// The Token field is only populated on creation.
type opServiceAccount struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Token string `json:"token,omitempty"`
}

func (r Repository) CreateServiceAccount(ctx context.Context, sa model.ServiceAccount) (*model.ServiceAccount, error) {
	cmdArgs := &onePasswordCliCmd{}
	cmdArgs.ServiceAccountArg().CreateArg().RawStrArg(sa.Name).CanCreateVaultsFlag().FormatJSONFlag()

	stdout, stderr, err := r.cli.RunOpCmd(ctx, cmdArgs.GetArgs())
	if err != nil {
		return nil, fmt.Errorf("op cli command failed: %w: %s", err, stderr)
	}

	var created opServiceAccount
	if err := json.Unmarshal([]byte(stdout), &created); err != nil {
		return nil, fmt.Errorf("could not unmarshal op cli stdout: %w", err)
	}

	if created.ID == "" {
		return nil, fmt.Errorf("service account create response missing id")
	}
	if created.Token == "" {
		return nil, fmt.Errorf("service account create response missing token")
	}
	// op CLI v2.34+ omits name from the create response; fall back to the input name.
	if created.Name == "" {
		created.Name = sa.Name
	}

	return &model.ServiceAccount{
		ID:    created.ID,
		Name:  created.Name,
		Token: created.Token,
	}, nil
}

func (r Repository) GetServiceAccountByID(ctx context.Context, id string) (*model.ServiceAccount, error) {
	// op CLI v2.34+ removed the 'service-account get' subcommand. Use ratelimit as an
	// existence probe: it exits non-zero if the service account does not exist.
	// The ratelimit response does not include the service account name; the resource's
	// Read function preserves it from prior Terraform state.
	cmdArgs := &onePasswordCliCmd{}
	cmdArgs.ServiceAccountArg().RatelimitArg().RawStrArg(id)

	_, stderr, err := r.cli.RunOpCmd(ctx, cmdArgs.GetArgs())
	if err != nil {
		return nil, fmt.Errorf("op cli command failed: %w: %s", err, stderr)
	}

	return &model.ServiceAccount{ID: id}, nil
}

func (r Repository) GetServiceAccountByName(ctx context.Context, name string) (*model.ServiceAccount, error) {
	return r.GetServiceAccountByID(ctx, name)
}

func (r Repository) DeleteServiceAccount(ctx context.Context, id string) error {
	cmdArgs := &onePasswordCliCmd{}
	cmdArgs.ServiceAccountArg().DeleteArg().RawStrArg(id)

	_, stderr, err := r.cli.RunOpCmd(ctx, cmdArgs.GetArgs())
	if err != nil {
		return fmt.Errorf("op cli command failed: %w: %s", err, stderr)
	}

	return nil
}
