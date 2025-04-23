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

const appID = "test_application_id"

func TestAccDataSourceApplication(t *testing.T) {
	t.Parallel()

	getReq := &apiservice.GetApplicationRequest{ApplicationId: appID}
	getResp := &apiservice.GetApplicationResponse{
		Application: &model.Application{
			Id:               appID,
			Name:             "test_name",
			ProjectId:        "test_project",
			PipedId:          "test_piped_id",
			Kind:             model.ApplicationKind_CLOUDRUN,
			PlatformProvider: "test_provider",
			Description:      "test_desc",
			GitPath: &model.ApplicationGitPath{
				Repo: &model.ApplicationGitRepository{
					Id:     "test_repo_id",
					Remote: "test_repo_remote",
					Branch: "test_repo_branch",
				},
				Path:           "test_git_path",
				ConfigFilename: "test_git_config_filename",
				Url:            "test_git_url",
			},
		},
	}

	ctrl := gomock.NewController(t)
	client := mock.NewMockAPIClient(ctrl)
	client.EXPECT().GetApplication(gomock.Any(), getReq).Return(getResp, nil).AnyTimes()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactories(client),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceApplication(appID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.pipecd_application.test", "id", appID),
					resource.TestCheckResourceAttr("data.pipecd_application.test", "name", "test_name"),
					resource.TestCheckResourceAttr("data.pipecd_application.test", "project_id", "test_project"),
					resource.TestCheckResourceAttr("data.pipecd_application.test", "piped_id", "test_piped_id"),
					resource.TestCheckResourceAttr("data.pipecd_application.test", "kind", "CLOUDRUN"),
					resource.TestCheckResourceAttr("data.pipecd_application.test", "platform_provider", "test_provider"),
					resource.TestCheckResourceAttr("data.pipecd_application.test", "description", "test_desc"),
					resource.TestCheckResourceAttr("data.pipecd_application.test", "git.repository_id", "test_repo_id"),
					resource.TestCheckResourceAttr("data.pipecd_application.test", "git.remote", "test_repo_remote"),
					resource.TestCheckResourceAttr("data.pipecd_application.test", "git.branch", "test_repo_branch"),
					resource.TestCheckResourceAttr("data.pipecd_application.test", "git.path", "test_git_path"),
					resource.TestCheckResourceAttr("data.pipecd_application.test", "git.filename", "test_git_config_filename"),
				),
			},
		},
	})
}

func TestAccDataSourceApplicationWithBoth(t *testing.T) {
	t.Parallel()

	getReq := &apiservice.GetApplicationRequest{ApplicationId: appID}
	getResp := &apiservice.GetApplicationResponse{
		Application: &model.Application{
			Id:               appID,
			Name:             "test_name",
			ProjectId:        "test_project",
			PipedId:          "test_piped_id",
			Kind:             model.ApplicationKind_CLOUDRUN,
			PlatformProvider: "test_provider",
			DeployTargetsByPlugin: map[string]*model.DeployTargets{
				"test_plugin": {
					DeployTargets: []string{"test_target"},
				},
				"test_plugin_2": {
					DeployTargets: []string{"test_target_2"},
				},
			},
			Description: "test_desc",
			GitPath: &model.ApplicationGitPath{
				Repo: &model.ApplicationGitRepository{
					Id:     "test_repo_id",
					Remote: "test_repo_remote",
					Branch: "test_repo_branch",
				},
				Path:           "test_git_path",
				ConfigFilename: "test_git_config_filename",
				Url:            "test_git_url",
			},
		},
	}

	ctrl := gomock.NewController(t)
	client := mock.NewMockAPIClient(ctrl)
	client.EXPECT().GetApplication(gomock.Any(), getReq).Return(getResp, nil).AnyTimes()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactories(client),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceApplication(appID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.pipecd_application.test", "id", appID),
					resource.TestCheckResourceAttr("data.pipecd_application.test", "name", "test_name"),
					resource.TestCheckResourceAttr("data.pipecd_application.test", "project_id", "test_project"),
					resource.TestCheckResourceAttr("data.pipecd_application.test", "piped_id", "test_piped_id"),
					resource.TestCheckResourceAttr("data.pipecd_application.test", "kind", "CLOUDRUN"),
					resource.TestCheckResourceAttr("data.pipecd_application.test", "platform_provider", "test_provider"),
					resource.TestCheckResourceAttr("data.pipecd_application.test", "plugins.#", "2"),
					resource.TestCheckResourceAttr("data.pipecd_application.test", "plugins.0.name", "test_plugin"),
					resource.TestCheckResourceAttr("data.pipecd_application.test", "plugins.0.deploy_targets.#", "1"),
					resource.TestCheckResourceAttr("data.pipecd_application.test", "plugins.0.deploy_targets.0", "test_target"),
					resource.TestCheckResourceAttr("data.pipecd_application.test", "plugins.1.name", "test_plugin_2"),
					resource.TestCheckResourceAttr("data.pipecd_application.test", "plugins.1.deploy_targets.#", "1"),
					resource.TestCheckResourceAttr("data.pipecd_application.test", "plugins.1.deploy_targets.0", "test_target_2"),
					resource.TestCheckResourceAttr("data.pipecd_application.test", "description", "test_desc"),
					resource.TestCheckResourceAttr("data.pipecd_application.test", "git.repository_id", "test_repo_id"),
					resource.TestCheckResourceAttr("data.pipecd_application.test", "git.remote", "test_repo_remote"),
					resource.TestCheckResourceAttr("data.pipecd_application.test", "git.branch", "test_repo_branch"),
					resource.TestCheckResourceAttr("data.pipecd_application.test", "git.path", "test_git_path"),
					resource.TestCheckResourceAttr("data.pipecd_application.test", "git.filename", "test_git_config_filename"),
				),
			},
		},
	})
}

func testAccDataSourceApplication(appID string) string {
	return providerConfig + fmt.Sprintf(`
data "pipecd_application" "test" {
	id = "%s"
}`, appID)
}
