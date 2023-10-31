// Copyright 2023 The PipeCD Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	api "github.com/pipe-cd/pipecd/pkg/app/server/service/apiservice"
	"github.com/pipe-cd/pipecd/pkg/model"
)

var (
	_ resource.Resource                = &ApplicationResource{}
	_ resource.ResourceWithImportState = &ApplicationResource{}
)

func NewApplicationResource() resource.Resource {
	return &ApplicationResource{}
}

type ApplicationResource struct {
	c APIClient
}

type (
	applicationResourceModel struct {
		ID               types.String                `tfsdk:"id"`
		Name             types.String                `tfsdk:"name"`
		PipedID          types.String                `tfsdk:"piped_id"`
		Kind             types.String                `tfsdk:"kind"`
		PlatformProvider types.String                `tfsdk:"platform_provider"`
		Description      types.String                `tfsdk:"description"`
		Git              applicationResourceGitModel `tfsdk:"git"`
	}

	applicationResourceGitModel struct {
		RepositoryID types.String `tfsdk:"repository_id"`
		Path         types.String `tfsdk:"path"`
		Filename     types.String `tfsdk:"filename"`
	}
)

func (a *ApplicationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	getReq := &api.GetApplicationRequest{
		ApplicationId: req.ID,
	}
	getResp, err := a.c.GetApplication(ctx, getReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading application",
			"Could not read application, unexpected error: "+err.Error(),
		)
		return
	}

	state := applicationResourceModel{
		ID:               types.StringValue(req.ID),
		Name:             types.StringValue(getResp.Application.Name),
		PipedID:          types.StringValue(getResp.Application.PipedId),
		Kind:             types.StringValue(getResp.Application.Kind.String()),
		PlatformProvider: types.StringValue(getResp.Application.PlatformProvider),
		Description:      types.StringValue(getResp.Application.Description),
		Git: applicationResourceGitModel{
			RepositoryID: types.StringValue(getResp.Application.GitPath.Repo.Id),
			Path:         types.StringValue(getResp.Application.GitPath.Path),
			Filename:     types.StringValue(getResp.Application.GitPath.ConfigFilename),
		},
	}
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (a *ApplicationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application"
}

func (a *ApplicationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "PipeCD application resource.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of this Application.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The application name.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"piped_id": schema.StringAttribute{
				Description: "The ID of piped that should handle this application.",
				Required:    true,
			},
			"kind": schema.StringAttribute{
				Description: "The kind of application.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					func() validator.String {
						values := make([]string, 0, len(model.ApplicationKind_value))
						for k := range model.ApplicationKind_value {
							values = append(values, k)
						}
						return stringvalidator.OneOf(values...)
					}(),
				},
			},
			"platform_provider": schema.StringAttribute{
				Description: "The platform provider name. One of the registered providers in the piped configuration. The previous name of this field is cloud-provider.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the application.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"git": schema.SingleNestedAttribute{
				Description: "Git path for the application.",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"repository_id": schema.StringAttribute{
						Description: "The repository ID. One the registered repositories in the piped configuration.",
						Required:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"path": schema.StringAttribute{
						Description: "The relative path from the root of repository to the application directory.",
						Required:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"filename": schema.StringAttribute{
						Description: "The configuration file name. (default \"app.pipecd.yaml\")",
						Optional:    true,
						Computed:    true,
						Default:     stringdefault.StaticString("app.pipecd.yaml"),
					},
				},
			},
		},
	}
}

func (a *applicationResourceModel) application() *model.Application {
	git := &model.ApplicationGitPath{
		Repo: &model.ApplicationGitRepository{
			Id: a.Git.RepositoryID.ValueString(),
		},
		Path:           a.Git.Path.ValueString(),
		ConfigFilename: a.Git.Filename.ValueString(),
	}
	kind := model.ApplicationKind_value[a.Kind.ValueString()]
	app := &model.Application{
		Id:               a.ID.ValueString(),
		Name:             a.Name.ValueString(),
		PipedId:          a.PipedID.ValueString(),
		GitPath:          git,
		Kind:             model.ApplicationKind(kind),
		PlatformProvider: a.PlatformProvider.ValueString(),
		Description:      a.Description.ValueString(),
	}
	return app
}

func (a *ApplicationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan applicationResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	app := plan.application()
	addReq := &api.AddApplicationRequest{
		Name:             app.Name,
		PipedId:          app.PipedId,
		GitPath:          app.GitPath,
		Kind:             app.Kind,
		PlatformProvider: app.PlatformProvider,
		Description:      app.Description,
	}

	addResp, err := a.c.AddApplication(ctx, addReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating application",
			"Could not create application, unexpected error: "+err.Error(),
		)
		return
	}

	getReq := &api.GetApplicationRequest{
		ApplicationId: addResp.ApplicationId,
	}
	getResp, err := a.c.GetApplication(ctx, getReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting application",
			"Could not get application, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "AddApplication response", map[string]interface{}{"response_fields": getResp})

	state := applicationResourceModel{
		ID:               types.StringValue(addResp.ApplicationId),
		Name:             types.StringValue(getResp.Application.Name),
		PipedID:          types.StringValue(getResp.Application.PipedId),
		Kind:             types.StringValue(getResp.Application.Kind.String()),
		PlatformProvider: types.StringValue(getResp.Application.PlatformProvider),
		Description:      types.StringValue(getResp.Application.Description),
		Git: applicationResourceGitModel{
			RepositoryID: types.StringValue(getResp.Application.GitPath.Repo.Id),
			Path:         types.StringValue(getResp.Application.GitPath.Path),
			Filename:     types.StringValue(getResp.Application.GitPath.ConfigFilename),
		},
	}
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (a *ApplicationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state applicationResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (a *ApplicationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan applicationResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	updateReq := &api.UpdateApplicationRequest{
		ApplicationId:    plan.application().Id,
		PipedId:          plan.application().PipedId,
		PlatformProvider: plan.application().PlatformProvider,
		GitPath:          plan.application().GitPath,
	}
	if _, err := a.c.UpdateApplication(ctx, updateReq); err != nil {
		resp.Diagnostics.AddError(
			"Error updating application",
			"Could not update application, unexpected error: "+err.Error(),
		)
		return
	}
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (a *ApplicationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state applicationResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	delReq := &api.DeleteApplicationRequest{
		ApplicationId: state.ID.ValueString(),
	}
	_, err := a.c.DeleteApplication(ctx, delReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting PipeCD application",
			"Could not delete application, unexpected error: "+err.Error(),
		)
		return
	}
}

func (a *ApplicationResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	a.c = req.ProviderData.(APIClient)
}
