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
	_ datasource.DataSource              = &applicationDataSource{}
	_ datasource.DataSourceWithConfigure = &applicationDataSource{}
)

func NewApplicationDataSource() datasource.DataSource {
	return &applicationDataSource{}
}

type applicationDataSource struct {
	c APIClient
}

type (
	applicationDataSourceModel struct {
		ID               types.String                       `tfsdk:"id"`
		Name             types.String                       `tfsdk:"name"`
		PipedID          types.String                       `tfsdk:"piped_id"`
		ProjectID        types.String                       `tfsdk:"project_id"`
		Kind             types.String                       `tfsdk:"kind"`
		PlatformProvider types.String                       `tfsdk:"platform_provider"`
		Plugins          []applicationDataSourcePluginModel `tfsdk:"plugins"`
		Description      types.String                       `tfsdk:"description"`
		Git              *applicationDataSourceGitModel     `tfsdk:"git"`
	}

	applicationDataSourceGitModel struct {
		RepositoryID types.String `tfsdk:"repository_id"`
		Remote       types.String `tfsdk:"remote"`
		Branch       types.String `tfsdk:"branch"`
		Path         types.String `tfsdk:"path"`
		Filename     types.String `tfsdk:"filename"`
	}

	applicationDataSourcePluginModel struct {
		Name          types.String   `tfsdk:"name"`
		DeployTargets []types.String `tfsdk:"deploy_targets"`
	}
)

func (a *applicationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application"
}

func (a *applicationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "PipeCD application resource.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of this Application.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "The application name.",
				Computed:    true,
			},
			"piped_id": schema.StringAttribute{
				Description: "The ID of piped that should handle this application.",
				Computed:    true,
			},
			"project_id": schema.StringAttribute{
				Computed: true,
			},
			"kind": schema.StringAttribute{
				Description: "The kind of application.",
				Computed:    true,
			},
			"platform_provider": schema.StringAttribute{
				Description:        "The platform provider name. One of the registered providers in the piped configuration. The previous name of this field is cloud-provider.",
				Computed:           true,
				DeprecationMessage: "Use `plugins` instead. This field will be removed in the next major version.",
			},
			"plugins": schema.ListNestedAttribute{
				Description: "The list of plugins that this application uses.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "The name of the plugin.",
							Computed:    true,
						},
						"deploy_targets": schema.ListAttribute{
							Description: "The list of deploy targets that this plugin uses.",
							Computed:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
			"description": schema.StringAttribute{
				Description: "The description of the application.",
				Computed:    true,
			},
			"git": schema.SingleNestedAttribute{
				Description: "Git path for the application.",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"repository_id": schema.StringAttribute{
						Description: "The repository ID. One the registered repositories in the piped configuration.",
						Computed:    true,
					},
					"remote": schema.StringAttribute{
						Computed: true,
					},
					"branch": schema.StringAttribute{
						Computed: true,
					},
					"path": schema.StringAttribute{
						Description: "The relative path from the root of repository to the application directory.",
						Computed:    true,
					},
					"filename": schema.StringAttribute{
						Description: "The configuration file name. (default \"app.pipecd.yaml\")",
						Computed:    true,
					},
				},
			},
		},
	}
}

func (a *applicationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	a.c = req.ProviderData.(APIClient)
}

func (a *applicationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state applicationDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	getReq := &api.GetApplicationRequest{
		ApplicationId: state.ID.ValueString(),
	}
	getResp, err := a.c.GetApplication(ctx, getReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read PipeCD application",
			err.Error(),
		)
		return
	}

	state = applicationDataSourceModel{
		ID:          types.StringValue(getResp.Application.Id),
		Name:        types.StringValue(getResp.Application.Name),
		PipedID:     types.StringValue(getResp.Application.PipedId),
		ProjectID:   types.StringValue(getResp.Application.ProjectId),
		Kind:        types.StringValue(getResp.Application.Kind.String()),
		Description: types.StringValue(getResp.Application.Description),
		Git: &applicationDataSourceGitModel{
			RepositoryID: types.StringValue(getResp.Application.GitPath.Repo.Id),
			Remote:       types.StringValue(getResp.Application.GitPath.Repo.Remote),
			Branch:       types.StringValue(getResp.Application.GitPath.Repo.Branch),
			Path:         types.StringValue(getResp.Application.GitPath.Path),
			Filename:     types.StringValue(getResp.Application.GitPath.ConfigFilename),
		},
	}

	if getResp.Application.PlatformProvider != "" {
		state.PlatformProvider = types.StringValue(getResp.Application.PlatformProvider)
	}

	if len(getResp.Application.DeployTargetsByPlugin) != 0 {
		state.Plugins = make([]applicationDataSourcePluginModel, 0, len(getResp.Application.DeployTargetsByPlugin))
		for k, v := range getResp.Application.DeployTargetsByPlugin {
			deployTargets := make([]types.String, 0, len(v.DeployTargets))
			for _, dt := range v.DeployTargets {
				deployTargets = append(deployTargets, types.StringValue(dt))
			}
			state.Plugins = append(state.Plugins, applicationDataSourcePluginModel{
				Name:          types.StringValue(k),
				DeployTargets: deployTargets,
			})
		}
		sort.Slice(state.Plugins, func(i, j int) bool {
			return state.Plugins[i].Name.ValueString() < state.Plugins[j].Name.ValueString()
		})
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
