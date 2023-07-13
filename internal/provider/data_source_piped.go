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

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	api "github.com/pipe-cd/pipecd/pkg/app/server/service/apiservice"
)

var (
	_ datasource.DataSource              = &pipedDataSource{}
	_ datasource.DataSourceWithConfigure = &pipedDataSource{}
)

func NewPipedDataSource() datasource.DataSource {
	return &pipedDataSource{}
}

type pipedDataSource struct {
	c APIClient
}

type (
	pipedDataSourceModel struct {
		ID                types.String                           `tfsdk:"id"`
		Name              types.String                           `tfsdk:"name"`
		Description       types.String                           `tfsdk:"description"`
		ProjectID         types.String                           `tfsdk:"project_id"`
		Repositories      []pipedDataSourceRepositoryModel       `tfsdk:"repositories"`
		PlatformProviders []pipedDataSourcePlatformProviderModel `tfsdk:"platform_providers"`
	}

	pipedDataSourceRepositoryModel struct {
		ID     types.String `tfsdk:"id"`
		Remote types.String `tfsdk:"remote"`
		Branch types.String `tfsdk:"branch"`
	}

	pipedDataSourcePlatformProviderModel struct {
		Name types.String `tfsdk:"name"`
		Type types.String `tfsdk:"type"`
	}
)

func (p *pipedDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_piped"
}

func (p *pipedDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "PipeCD pied data source.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required: true,
			},
			"name": schema.StringAttribute{
				Computed: true,
			},
			"description": schema.StringAttribute{
				Computed: true,
			},
			"project_id": schema.StringAttribute{
				Computed: true,
			},
			"repositories": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed: true,
						},
						"remote": schema.StringAttribute{
							Computed: true,
						},
						"branch": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
			"platform_providers": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed: true,
						},
						"type": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (p *pipedDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	p.c = req.ProviderData.(APIClient)
}

func (p *pipedDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state pipedDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	getReq := &api.GetPipedRequest{
		PipedId: state.ID.ValueString(),
	}
	getResp, err := p.c.GetPiped(ctx, getReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read PipeCD piped",
			err.Error(),
		)
		return
	}

	repos := make([]pipedDataSourceRepositoryModel, 0, len(getResp.Piped.Repositories))
	for _, r := range getResp.Piped.Repositories {
		repos = append(repos, pipedDataSourceRepositoryModel{
			ID:     types.StringValue(r.Id),
			Remote: types.StringValue(r.Remote),
			Branch: types.StringValue(r.Branch),
		})
	}

	providers := make([]pipedDataSourcePlatformProviderModel, 0, len(getResp.Piped.PlatformProviders))
	for _, p := range getResp.Piped.PlatformProviders {
		providers = append(providers, pipedDataSourcePlatformProviderModel{
			Name: types.StringValue(p.Name),
			Type: types.StringValue(p.Type),
		})
	}

	state = pipedDataSourceModel{
		ID:                types.StringValue(getResp.Piped.Id),
		Name:              types.StringValue(getResp.Piped.Name),
		ProjectID:         types.StringValue(getResp.Piped.ProjectId),
		Description:       types.StringValue(getResp.Piped.Desc),
		Repositories:      repos,
		PlatformProviders: providers,
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
