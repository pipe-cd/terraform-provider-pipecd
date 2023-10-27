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
	"log"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	api "github.com/pipe-cd/pipecd/pkg/app/server/service/apiservice"
	"github.com/pipe-cd/pipecd/pkg/model"
)

var (
	_ resource.Resource                = &PipedResource{}
	_ resource.ResourceWithImportState = &PipedResource{}
)

func NewPipedResource() resource.Resource {
	return &PipedResource{}
}

type PipedResource struct {
	c APIClient
}

type (
	pipedResourceModel struct {
		ID          types.String `tfsdk:"id"`
		Name        types.String `tfsdk:"name"`
		Description types.String `tfsdk:"description"`
		APIKey      types.String `tfsdk:"api_key"`
	}
)

func (p *PipedResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	getReq := &api.GetPipedRequest{
		PipedId: req.ID,
	}
	getResp, err := p.c.GetPiped(ctx, getReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading piped",
			"Could not read piped, unexpected error: "+err.Error(),
		)
		return
	}

	state := pipedResourceModel{
		ID:          types.StringValue(req.ID),
		Name:        types.StringValue(getResp.Piped.Name),
		Description: types.StringValue(getResp.Piped.Desc),
		APIKey:      types.StringUnknown(),
	}
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (p *PipedResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_piped"
}

func (p *PipedResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "PipeCD piped resource.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of piped that should handle this application.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The piped name.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the piped.",
				Optional:    true,
				Computed:    true,
			},
			"api_key": schema.StringAttribute{
				Description: "The API key of the piped.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (p *pipedResourceModel) piped() *model.Piped {
	piped := &model.Piped{
		Id:   p.ID.ValueString(),
		Name: p.Name.ValueString(),
		Desc: p.Description.ValueString(),
	}
	return piped
}

func (p *PipedResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan pipedResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	piped := plan.piped()
	registerReq := &api.RegisterPipedRequest{
		Name: piped.Name,
		Desc: piped.Desc,
	}

	registerResp, err := p.c.RegisterPiped(ctx, registerReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating piped",
			"Could not create piped, unexpected error: "+err.Error(),
		)
		return
	}

	plan = pipedResourceModel{
		ID:          types.StringValue(registerResp.Id),
		Name:        types.StringValue(piped.Name),
		Description: types.StringValue(piped.Desc),
		APIKey:      types.StringValue(registerResp.Key),
	}
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (p *PipedResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state pipedResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (p *PipedResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan pipedResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	piped := plan.piped()
	updateReq := &api.UpdatePipedRequest{
		PipedId: piped.Id,
		Name:    piped.Name,
		Desc:    piped.Desc,
	}

	_, err := p.c.UpdatePiped(ctx, updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating piped",
			"Could not update piped, unexpected error: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (p *PipedResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state pipedResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	log.Printf("[WARNING] PipeCD Piped resources"+
		" cannot be deleted. The resource %s will be disabled and removed from Terraform"+
		" state, but will still be present on PipeCD Control Plane.", state.ID.ValueString())

	disableReq := &api.DisablePipedRequest{
		PipedId: state.ID.ValueString(),
	}
	_, err := p.c.DisablePiped(ctx, disableReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Disabling PipeCD piped",
			"Could not disable piped, unexpected error: "+err.Error(),
		)
		return
	}

	resp.State.RemoveResource(ctx)
}

func (p *PipedResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	p.c = req.ProviderData.(APIClient)
}
