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

	return &model.ServiceAccount{
		ID:    created.ID,
		Name:  created.Name,
		Token: created.Token,
	}, nil
}

func (r Repository) GetServiceAccountByID(ctx context.Context, id string) (*model.ServiceAccount, error) {
	cmdArgs := &onePasswordCliCmd{}
	cmdArgs.ServiceAccountArg().GetArg().RawStrArg(id).FormatJSONFlag()

	stdout, stderr, err := r.cli.RunOpCmd(ctx, cmdArgs.GetArgs())
	if err != nil {
		return nil, fmt.Errorf("op cli command failed: %w: %s", err, stderr)
	}

	var got opServiceAccount
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		return nil, fmt.Errorf("could not unmarshal op cli stdout: %w", err)
	}

	return &model.ServiceAccount{
		ID:   got.ID,
		Name: got.Name,
	}, nil
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
