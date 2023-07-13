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

//go:generate mockgen -source=$GOFILE -package=mock -destination=./mock/mock.go

package provider

import (
	"context"
	"crypto/tls"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"google.golang.org/grpc/credentials"

	api "github.com/pipe-cd/pipecd/pkg/app/server/service/apiservice"
	"github.com/pipe-cd/pipecd/pkg/rpc/rpcauth"
	"github.com/pipe-cd/pipecd/pkg/rpc/rpcclient"
)

var _ provider.Provider = &PipeCDProvider{}

type PipeCDProvider struct {
	version string
	client  APIClient
}

type pipeCDProviderModel struct {
	Host   types.String `tfsdk:"host"`
	APIKey types.String `tfsdk:"api_key"`
}

func (p *PipeCDProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "pipecd"
}

func (p *PipeCDProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interact with PipeCD.",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Optional: true,
			},
			"api_key": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
			},
		},
	}
}

func (p *PipeCDProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring PipeCD client")

	var config pipeCDProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.Host.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Unknown PipeCD API Host",
			"The provider cannot create the PipeCD API client as there is an unknown configuration value for the PipeCD API host. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the PIPECD_HOST environment variable.",
		)
	}

	if config.APIKey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Unknown PipeCD API Key",
			"The provider cannot create the PipeCD API client as there is an unknown configuration value for the PipeCD API Key. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the PIPECD_API_KEY environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	host := os.Getenv("PIPECD_HOST")
	apiKey := os.Getenv("PIPECD_API_KEY")

	if !config.Host.IsNull() {
		host = config.Host.ValueString()
	}

	if !config.APIKey.IsNull() {
		apiKey = config.APIKey.ValueString()
	}

	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing PipeCD API Host",
			"The provider cannot create the PipeCD API client as there is a missing or empty value for the PipeCD API host. "+
				"Set the host value in the configuration or use the PIPECD_HOST environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if apiKey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Missing PipeCD API Key",
			"The provider cannot create the PipeCD API client as there is a missing or empty value for the PipeCD API Key. "+
				"Set the api_key value in the configuration or use the PIPECD_API_KEY environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "pipecd_host", host)
	ctx = tflog.SetField(ctx, "pipecd_api_key", apiKey)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "pipecd_api_key")

	tflog.Debug(ctx, "Creating PipeCD client")

	if p.client == nil {
		creds := rpcclient.NewPerRPCCredentials(apiKey, rpcauth.APIKeyCredentials, true)
		tlsConfig := &tls.Config{}
		options := []rpcclient.DialOption{
			rpcclient.WithBlock(),
			rpcclient.WithPerRPCCredentials(creds),
			rpcclient.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
		}
		client, err := api.NewClient(ctx, host, options...)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Create PipeCD API Client",
				"An unexpected error occurred when creating the PipeCD API client. "+
					"If the error is not clear, please contact the provider developers.\n\n"+
					"PipeCD Client Error: "+err.Error(),
			)
			return
		}
		p.client = client
	}

	resp.DataSourceData = p.client
	resp.ResourceData = p.client

	tflog.Info(ctx, "Configured PipeCD client", map[string]any{"success": true})
}

func (p *PipeCDProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewApplicationDataSource,
		NewPipedDataSource,
	}
}

func (p *PipeCDProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewApplicationResource,
		NewPipedResource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &PipeCDProvider{
			version: version,
		}
	}
}

type APIClient interface {
	api.APIServiceClient
}
