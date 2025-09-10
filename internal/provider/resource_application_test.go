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
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"go.uber.org/mock/gomock"

	"github.com/pipe-cd/pipecd/pkg/app/server/service/apiservice"
	"github.com/pipe-cd/pipecd/pkg/model"
	"github.com/pipe-cd/terraform-provider-pipecd/internal/provider/mock"
)

func TestAccResourceApplication(t *testing.T) {
	t.Parallel()

	const appID = "test_application_id"

	appGit := &model.ApplicationGitPath{
		Repo: &model.ApplicationGitRepository{
			Id: "repo_id",
		},
		Path:           "path/to/config",
		ConfigFilename: "testapp.pipecd.yaml",
	}

	app := &model.Application{
		Id:               appID,
		Name:             "test_application",
		PipedId:          "test_piped_id",
		GitPath:          appGit,
		Kind:             model.ApplicationKind_CLOUDRUN,
		PlatformProvider: "test_provider",
		Description:      "test description",
	}

	addReq := &apiservice.AddApplicationRequest{
		Name:             app.Name,
		PipedId:          app.PipedId,
		GitPath:          app.GitPath,
		Kind:             app.Kind,
		PlatformProvider: app.PlatformProvider,
		Description:      app.Description,
	}
	addResp := &apiservice.AddApplicationResponse{ApplicationId: appID}

	getReq := &apiservice.GetApplicationRequest{ApplicationId: appID}
	getResp := &apiservice.GetApplicationResponse{Application: app}

	updateReq := &apiservice.UpdateApplicationRequest{ApplicationId: appID}
	updateResp := &apiservice.UpdateApplicationResponse{ApplicationId: appID}

	deleteReq := &apiservice.DeleteApplicationRequest{ApplicationId: appID}
	deleteResp := &apiservice.DeleteApplicationResponse{ApplicationId: appID}

	ctrl := gomock.NewController(t)
	client := mock.NewMockAPIClient(ctrl)
	client.EXPECT().AddApplication(gomock.Any(), addReq).Return(addResp, nil).AnyTimes()
	client.EXPECT().GetApplication(gomock.Any(), getReq).Return(getResp, nil).AnyTimes()
	client.EXPECT().UpdateApplication(gomock.Any(), updateReq).Return(updateResp, nil).AnyTimes()
	client.EXPECT().DeleteApplication(gomock.Any(), deleteReq).Return(deleteResp, nil).AnyTimes()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactories(client),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceApplication(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pipecd_application.test", "name", "test_application"),
					resource.TestCheckResourceAttr("pipecd_application.test", "piped_id", "test_piped_id"),
					resource.TestCheckResourceAttr("pipecd_application.test", "platform_provider", "test_provider"),
					resource.TestCheckResourceAttr("pipecd_application.test", "description", "test description"),
					resource.TestCheckResourceAttr("pipecd_application.test", "git.repository_id", "repo_id"),
					resource.TestCheckResourceAttr("pipecd_application.test", "git.path", "path/to/config"),
					resource.TestCheckResourceAttr("pipecd_application.test", "git.filename", "testapp.pipecd.yaml"),
				),
			},
		},
	})
}

func TestAccResourceApplicationWithBoth(t *testing.T) {
	t.Parallel()

	const appID = "test_application_id"

	appGit := &model.ApplicationGitPath{
		Repo: &model.ApplicationGitRepository{
			Id: "repo_id",
		},
		Path:           "path/to/config",
		ConfigFilename: "testapp.pipecd.yaml",
	}

	app := &model.Application{
		Id:               appID,
		Name:             "test_application",
		PipedId:          "test_piped_id",
		GitPath:          appGit,
		Kind:             model.ApplicationKind_CLOUDRUN,
		PlatformProvider: "test_provider",
		Description:      "test description",
		DeployTargetsByPlugin: map[string]*model.DeployTargets{
			"test_plugin": {
				DeployTargets: []string{"target1", "target2"},
			},
			"test_plugin_2": {
				DeployTargets: []string{"target3", "target4"},
			},
		},
	}

	addReq := &apiservice.AddApplicationRequest{
		Name:             app.Name,
		PipedId:          app.PipedId,
		GitPath:          app.GitPath,
		Kind:             app.Kind,
		PlatformProvider: app.PlatformProvider,
		Description:      app.Description,
	}
	addResp := &apiservice.AddApplicationResponse{ApplicationId: appID}

	getReq := &apiservice.GetApplicationRequest{ApplicationId: appID}
	getResp := &apiservice.GetApplicationResponse{Application: app}

	updateReq := &apiservice.UpdateApplicationRequest{ApplicationId: appID}
	updateResp := &apiservice.UpdateApplicationResponse{ApplicationId: appID}

	updatePluginsReq := &apiservice.UpdateApplicationDeployTargetsRequest{
		ApplicationId: appID,
		DeployTargetsByPlugin: map[string]*model.DeployTargets{
			"test_plugin": {
				DeployTargets: []string{"target1", "target2"},
			},
			"test_plugin_2": {
				DeployTargets: []string{"target3", "target4"},
			},
		},
	}

	deleteReq := &apiservice.DeleteApplicationRequest{ApplicationId: appID}
	deleteResp := &apiservice.DeleteApplicationResponse{ApplicationId: appID}

	ctrl := gomock.NewController(t)
	client := mock.NewMockAPIClient(ctrl)
	client.EXPECT().AddApplication(gomock.Any(), addReq).Return(addResp, nil).AnyTimes()
	client.EXPECT().GetApplication(gomock.Any(), getReq).Return(getResp, nil).AnyTimes()
	client.EXPECT().UpdateApplication(gomock.Any(), updateReq).Return(updateResp, nil).AnyTimes()
	client.EXPECT().UpdateApplicationDeployTargets(gomock.Any(), updatePluginsReq).Return(nil, nil).AnyTimes()
	client.EXPECT().DeleteApplication(gomock.Any(), deleteReq).Return(deleteResp, nil).AnyTimes()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactories(client),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceApplicationWithPlugins(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pipecd_application.test", "name", "test_application"),
					resource.TestCheckResourceAttr("pipecd_application.test", "piped_id", "test_piped_id"),
					resource.TestCheckResourceAttr("pipecd_application.test", "platform_provider", "test_provider"),
					resource.TestCheckResourceAttr("pipecd_application.test", "description", "test description"),
					resource.TestCheckResourceAttr("pipecd_application.test", "git.repository_id", "repo_id"),
					resource.TestCheckResourceAttr("pipecd_application.test", "git.path", "path/to/config"),
					resource.TestCheckResourceAttr("pipecd_application.test", "git.filename", "testapp.pipecd.yaml"),
					resource.TestCheckResourceAttr("pipecd_application.test", "plugins.0.name", "test_plugin"),
					resource.TestCheckResourceAttr("pipecd_application.test", "plugins.0.deploy_targets.#", "2"),
					resource.TestCheckResourceAttr("pipecd_application.test", "plugins.0.deploy_targets.0", "target1"),
					resource.TestCheckResourceAttr("pipecd_application.test", "plugins.0.deploy_targets.1", "target2"),
					resource.TestCheckResourceAttr("pipecd_application.test", "plugins.1.name", "test_plugin_2"),
					resource.TestCheckResourceAttr("pipecd_application.test", "plugins.1.deploy_targets.#", "2"),
					resource.TestCheckResourceAttr("pipecd_application.test", "plugins.1.deploy_targets.0", "target3"),
					resource.TestCheckResourceAttr("pipecd_application.test", "plugins.1.deploy_targets.1", "target4"),
				),
			},
		},
	})
}

func testAccResourceApplication() string {
	return providerConfig + `
resource "pipecd_application" "test" {
	name = "test_application"
	piped_id = "test_piped_id"
	kind = "CLOUDRUN"
	platform_provider = "test_provider"
	description = "test description"
	git = {
		repository_id = "repo_id"
		path = "path/to/config"
		filename = "testapp.pipecd.yaml"
	}
}`
}

func testAccResourceApplicationWithPlugins() string {
	return providerConfig + `
resource "pipecd_application" "test" {
	name = "test_application"
	piped_id = "test_piped_id"
	kind = "CLOUDRUN"
	platform_provider = "test_provider"
	description = "test description"
	git = {
		repository_id = "repo_id"
		path = "path/to/config"
		filename = "testapp.pipecd.yaml"
	}
	plugins = [
		{
			name = "test_plugin"
			deploy_targets = ["target1", "target2"]
		},
		{
			name = "test_plugin_2"
			deploy_targets = ["target3", "target4"]
		}
	]
}`
}
