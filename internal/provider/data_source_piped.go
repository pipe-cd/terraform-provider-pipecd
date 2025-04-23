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
	"sort"

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
		Plugins           []pipedDataSourcePluginModel           `tfsdk:"plugins"`
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

	pipedDataSourcePluginModel struct {
		Name          types.String   `tfsdk:"name"`
		DeployTargets []types.String `tfsdk:"deploy_targets"`
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
				DeprecationMessage: "Use `plugins` instead. This field will be removed in the next major version.",
			},
			"plugins": schema.ListNestedAttribute{
				Description: "The list of plugins that this piped uses.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the plugin.",
						},
						"deploy_targets": schema.ListAttribute{
							Computed:    true,
							ElementType: types.StringType,
							Description: "The list of deploy targets that this plugin uses.",
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

	state = pipedDataSourceModel{
		ID:          types.StringValue(getResp.Piped.Id),
		Name:        types.StringValue(getResp.Piped.Name),
		ProjectID:   types.StringValue(getResp.Piped.ProjectId),
		Description: types.StringValue(getResp.Piped.Desc),
	}

	if len(getResp.Piped.Repositories) != 0 {
		repos := make([]pipedDataSourceRepositoryModel, 0, len(getResp.Piped.Repositories))
		for _, r := range getResp.Piped.Repositories {
			repos = append(repos, pipedDataSourceRepositoryModel{
				ID:     types.StringValue(r.Id),
				Remote: types.StringValue(r.Remote),
				Branch: types.StringValue(r.Branch),
			})
		}
		state.Repositories = repos
	}

	if len(getResp.Piped.PlatformProviders) != 0 {
		providers := make([]pipedDataSourcePlatformProviderModel, 0, len(getResp.Piped.PlatformProviders))
		for _, p := range getResp.Piped.PlatformProviders {
			providers = append(providers, pipedDataSourcePlatformProviderModel{
				Name: types.StringValue(p.Name),
				Type: types.StringValue(p.Type),
			})
		}
		state.PlatformProviders = providers
	}

	if len(getResp.Piped.Plugins) != 0 {
		plugins := make([]pipedDataSourcePluginModel, 0, len(getResp.Piped.Plugins))
		for _, p := range getResp.Piped.Plugins {
			deployTargets := make([]types.String, 0, len(p.DeployTargets))
			for _, dt := range p.DeployTargets {
				deployTargets = append(deployTargets, types.StringValue(dt))
			}
			plugins = append(plugins, pipedDataSourcePluginModel{
				Name:          types.StringValue(p.Name),
				DeployTargets: deployTargets,
			})
		}
		sort.Slice(plugins, func(i, j int) bool {
			return plugins[i].Name.ValueString() < plugins[j].Name.ValueString()
		})
		state.Plugins = plugins
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
