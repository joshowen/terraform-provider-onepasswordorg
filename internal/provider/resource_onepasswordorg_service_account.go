package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/slok/terraform-provider-onepasswordorg/internal/storage"
)

var (
	_ resource.Resource                = &serviceAccountResource{}
	_ resource.ResourceWithConfigure   = &serviceAccountResource{}
	_ resource.ResourceWithImportState = &serviceAccountResource{}
)

func NewServiceAccountResource() resource.Resource {
	return &serviceAccountResource{}
}

type serviceAccountResource struct {
	repo storage.Repository
}

func (r *serviceAccountResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_account"
}

func (r *serviceAccountResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `
Provides a Service Account resource.

A 1Password service account is a non-human account used for automated access to 1Password. Upon creation,
a bearer token is generated once and stored in Terraform state. This token cannot be retrieved again after
creation, so it must be captured from state immediately after the resource is created.

~> **Note:** Changing the name of a service account forces the creation of a new resource because 1Password
does not support renaming service accounts via the CLI.

~> **Note:** The token attribute is sensitive and stored in Terraform state. Protect your state file accordingly.
If you import an existing service account, the token will be empty because it cannot be retrieved.
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the service account.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the service account.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"token": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "The bearer token for the service account. Only available at creation time; preserved in state across subsequent plans.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *serviceAccountResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	appServices := getAppServicesFromResourceRequest(&req)
	if appServices == nil {
		return
	}

	r.repo = appServices.Repository
}

func (r *serviceAccountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var tfSA ServiceAccount
	diags := req.Plan.Get(ctx, &tfSA)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	sa := mapTfToModelServiceAccount(tfSA)
	newSA, err := r.repo.CreateServiceAccount(ctx, sa)
	if err != nil {
		resp.Diagnostics.AddError("Error creating service account", "Could not create service account, unexpected error: "+err.Error())
		return
	}

	newTfSA := mapModelToTfServiceAccount(*newSA)

	diags = resp.State.Set(ctx, newTfSA)
	resp.Diagnostics.Append(diags...)
}

func (r *serviceAccountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state (including the token we stored at creation).
	var state ServiceAccount
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	sa, err := r.repo.GetServiceAccountByID(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading service account", fmt.Sprintf("Could not get service account %q, unexpected error: %s", id, err.Error()))
		return
	}

	// The API never returns the token after creation, so preserve it from state.
	// op CLI v2.34+ also no longer returns the name on read; preserve it from state too.
	readTfSA := mapModelToTfServiceAccount(*sa)
	readTfSA.Token = state.Token
	readTfSA.Name = state.Name

	diags = resp.State.Set(ctx, readTfSA)
	resp.Diagnostics.Append(diags...)
}

// Update is intentionally a no-op: all mutable attributes use RequiresReplace,
// so any change triggers a destroy+create cycle instead.
func (r *serviceAccountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Service accounts cannot be updated", "All changes to service accounts require replacement. This is an internal error.")
}

func (r *serviceAccountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var tfSA ServiceAccount
	diags := req.State.Get(ctx, &tfSA)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := tfSA.ID.ValueString()
	err := r.repo.DeleteServiceAccount(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting service account", fmt.Sprintf("Could not delete service account %q, unexpected error: %s", id, err.Error()))
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *serviceAccountResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
