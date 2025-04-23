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
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"go.uber.org/mock/gomock"

	"github.com/pipe-cd/pipecd/pkg/app/server/service/apiservice"
	"github.com/pipe-cd/pipecd/pkg/model"
	"github.com/pipe-cd/terraform-provider-pipecd/internal/provider/mock"
)

func TestAccDataSourcePiped(t *testing.T) {
	t.Parallel()

	const pipedID = "test_piped_id"

	getReq := &apiservice.GetPipedRequest{PipedId: pipedID}
	getResp := &apiservice.GetPipedResponse{
		Piped: &model.Piped{
			Id:        pipedID,
			Name:      "test_name",
			Desc:      "test_desc",
			ProjectId: "test_project",
			Repositories: []*model.ApplicationGitRepository{
				{
					Id:     "test_repo_id",
					Remote: "test_repo_remote",
					Branch: "test_repo_branch",
				},
			},
			PlatformProviders: []*model.Piped_PlatformProvider{
				{
					Name: "test_provider_name",
					Type: "test_provider_type",
				},
			},
		},
	}

	ctrl := gomock.NewController(t)
	client := mock.NewMockAPIClient(ctrl)
	client.EXPECT().GetPiped(gomock.Any(), getReq).Return(getResp, nil).AnyTimes()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactories(client),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourcePiped(pipedID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.pipecd_piped.test", "id", pipedID),
					resource.TestCheckResourceAttr("data.pipecd_piped.test", "name", "test_name"),
					resource.TestCheckResourceAttr("data.pipecd_piped.test", "description", "test_desc"),
					resource.TestCheckResourceAttr("data.pipecd_piped.test", "project_id", "test_project"),
					resource.TestCheckResourceAttr("data.pipecd_piped.test", "repositories.#", "1"),
					resource.TestCheckResourceAttr("data.pipecd_piped.test", "repositories.0.id", "test_repo_id"),
					resource.TestCheckResourceAttr("data.pipecd_piped.test", "repositories.0.remote", "test_repo_remote"),
					resource.TestCheckResourceAttr("data.pipecd_piped.test", "repositories.0.branch", "test_repo_branch"),
					resource.TestCheckResourceAttr("data.pipecd_piped.test", "platform_providers.#", "1"),
					resource.TestCheckResourceAttr("data.pipecd_piped.test", "platform_providers.0.name", "test_provider_name"),
					resource.TestCheckResourceAttr("data.pipecd_piped.test", "platform_providers.0.type", "test_provider_type"),
				),
			},
		},
	})
}

func testAccDataSourcePiped(pipedID string) string {
	return providerConfig + fmt.Sprintf(`
data "pipecd_piped" "test" {
	id = "%s"
}`, pipedID)
}
