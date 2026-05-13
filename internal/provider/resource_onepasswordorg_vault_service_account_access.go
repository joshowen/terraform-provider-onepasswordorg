package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/slok/terraform-provider-onepasswordorg/internal/model"
	"github.com/slok/terraform-provider-onepasswordorg/internal/storage"
)

var (
	_ resource.Resource                = &vaultServiceAccountAccessResource{}
	_ resource.ResourceWithConfigure   = &vaultServiceAccountAccessResource{}
	_ resource.ResourceWithImportState = &vaultServiceAccountAccessResource{}
)

func NewVaultServiceAccountAccessResource() resource.Resource {
	return &vaultServiceAccountAccessResource{}
}

type vaultServiceAccountAccessResource struct {
	repo storage.Repository
}

func (r *vaultServiceAccountAccessResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vault_service_account_access"
}

func (r *vaultServiceAccountAccessResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `
Provides vault access for a service account.
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"vault_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				Description: "The vault ID.",
			},
			"service_account_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				Description: "The service account ID.",
			},
			"permissions": permissionsAttribute,
		},
	}
}

func (r *vaultServiceAccountAccessResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	appServices := getAppServicesFromResourceRequest(&req)
	if appServices == nil {
		return
	}

	r.repo = appServices.Repository
}

func (r *vaultServiceAccountAccessResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var tfAccess VaultServiceAccountAccess
	diags := req.Plan.Get(ctx, &tfAccess)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	v, err := mapTfToModelVaultServiceAccountAccess(tfAccess)
	if err != nil {
		resp.Diagnostics.AddError("Error mapping access", "Could not map access: "+err.Error())
		return
	}

	err = r.repo.EnsureVaultServiceAccountAccess(ctx, *v)
	if err != nil {
		resp.Diagnostics.AddError("Error creating access", "Could not create access, unexpected error: "+err.Error())
		return
	}

	newTfAccess, err := mapModelToTfVaultServiceAccountAccess(*v)
	if err != nil {
		resp.Diagnostics.AddError("Error mapping access", "Could not map access: "+err.Error())
		return
	}

	diags = resp.State.Set(ctx, newTfAccess)
	resp.Diagnostics.Append(diags...)
}

func (r *vaultServiceAccountAccessResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var tfAccess VaultServiceAccountAccess
	diags := req.State.Get(ctx, &tfAccess)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := tfAccess.ID.ValueString()
	vaultID, saID, err := unpackVaultServiceAccountAccessID(id)
	if err != nil {
		resp.Diagnostics.AddError("Error getting access ID", "Could not get access ID: "+err.Error())
		return
	}

	access, err := r.repo.GetVaultServiceAccountAccessByID(ctx, vaultID, saID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading access", fmt.Sprintf("Could not get access %q, unexpected error: %s", id, err.Error()))
		return
	}

	readTfAccess, err := mapModelToTfVaultServiceAccountAccess(*access)
	if err != nil {
		resp.Diagnostics.AddError("Error mapping access", "Could not map access: "+err.Error())
		return
	}

	diags = resp.State.Set(ctx, readTfAccess)
	resp.Diagnostics.Append(diags...)
}

func (r *vaultServiceAccountAccessResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan VaultServiceAccountAccess
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state VaultServiceAccountAccess
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = state.ID

	v, err := mapTfToModelVaultServiceAccountAccess(plan)
	if err != nil {
		resp.Diagnostics.AddError("Error mapping access", "Could not map access: "+err.Error())
		return
	}

	err = r.repo.EnsureVaultServiceAccountAccess(ctx, *v)
	if err != nil {
		resp.Diagnostics.AddError("Error updating access", "Could not update access, unexpected error: "+err.Error())
		return
	}

	newTfAccess, err := mapModelToTfVaultServiceAccountAccess(*v)
	if err != nil {
		resp.Diagnostics.AddError("Error mapping access", "Could not map access: "+err.Error())
		return
	}

	diags = resp.State.Set(ctx, newTfAccess)
	resp.Diagnostics.Append(diags...)
}

func (r *vaultServiceAccountAccessResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var tfAccess VaultServiceAccountAccess
	diags := req.State.Get(ctx, &tfAccess)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := tfAccess.ID.ValueString()
	vaultID, saID, err := unpackVaultServiceAccountAccessID(id)
	if err != nil {
		resp.Diagnostics.AddError("Error getting access ID", "Could not get access ID: "+err.Error())
		return
	}

	err = r.repo.DeleteVaultServiceAccountAccess(ctx, vaultID, saID)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting access", fmt.Sprintf("Could not delete access %q, unexpected error: %s", id, err.Error()))
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *vaultServiceAccountAccessResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func mapTfToModelVaultServiceAccountAccess(m VaultServiceAccountAccess) (*model.VaultServiceAccountAccess, error) {
	saID := m.ServiceAccountID.ValueString()
	vaultID := m.VaultID.ValueString()

	if m.ID.ValueString() != "" {
		vid, sid, err := unpackVaultServiceAccountAccessID(m.ID.ValueString())
		if err != nil {
			return nil, err
		}
		if sid != saID {
			return nil, fmt.Errorf("resource id is wrong based on service account ID")
		}
		if vid != vaultID {
			return nil, fmt.Errorf("resource id is wrong based on vault ID")
		}
	}

	return &model.VaultServiceAccountAccess{
		VaultID:          vaultID,
		ServiceAccountID: saID,
		Permissions:      mapTfToModelAccessPermissions(*m.Permissions),
	}, nil
}

func mapModelToTfVaultServiceAccountAccess(m model.VaultServiceAccountAccess) (*VaultServiceAccountAccess, error) {
	id := packVaultServiceAccountAccessID(m.VaultID, m.ServiceAccountID)

	return &VaultServiceAccountAccess{
		ID:               types.StringValue(id),
		ServiceAccountID: types.StringValue(m.ServiceAccountID),
		VaultID:          types.StringValue(m.VaultID),
		Permissions:      mapModelToTfAccessPermissions(m.Permissions),
	}, nil
}

func packVaultServiceAccountAccessID(vaultID, serviceAccountID string) string {
	return vaultID + "/" + serviceAccountID
}

func unpackVaultServiceAccountAccessID(id string) (vaultID, serviceAccountID string, err error) {
	s := strings.SplitN(id, "/", 2)
	if len(s) != 2 {
		return "", "", fmt.Errorf(
			"invalid vault service account access ID format: %s (expected <VAULT ID>/<SERVICE ACCOUNT ID>)", id)
	}

	return s[0], s[1], nil
}
