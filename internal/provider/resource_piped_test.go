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

	"github.com/golang/mock/gomock"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/pipe-cd/pipecd/pkg/app/server/service/apiservice"
	"github.com/pipe-cd/pipecd/pkg/model"
	"github.com/pipe-cd/terraform-provider-pipecd/internal/provider/mock"
)

func TestAccResourcePiped(t *testing.T) {
	t.Parallel()

	const pipedID = "test_piped_id"
	const pipedAPIKey = "test_piped_api_key"

	piped := &model.Piped{
		Id:   pipedID,
		Name: "test_piped",
		Desc: "test description",
	}

	registerReq := &apiservice.RegisterPipedRequest{
		Name: piped.Name,
		Desc: piped.Desc,
	}
	registerResp := &apiservice.RegisterPipedResponse{Id: pipedID, Key: pipedAPIKey}

	updateReq := &apiservice.UpdatePipedRequest{
		PipedId: pipedID,
		Name:    piped.Name,
		Desc:    piped.Desc,
	}
	updateResp := &apiservice.UpdatePipedResponse{}

	getReq := &apiservice.GetPipedRequest{PipedId: pipedID}
	getResp := &apiservice.GetPipedResponse{Piped: piped}

	disableReq := &apiservice.DisablePipedRequest{PipedId: pipedID}
	disableResp := &apiservice.DisablePipedResponse{}

	ctrl := gomock.NewController(t)
	client := mock.NewMockAPIClient(ctrl)
	client.EXPECT().RegisterPiped(gomock.Any(), registerReq).Return(registerResp, nil).AnyTimes()
	client.EXPECT().UpdatePiped(gomock.Any(), updateReq).Return(updateResp, nil).AnyTimes()
	client.EXPECT().GetPiped(gomock.Any(), getReq).Return(getResp, nil).AnyTimes()
	client.EXPECT().DisablePiped(gomock.Any(), disableReq).Return(disableResp, nil).AnyTimes()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactories(client),
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePiped(registerReq.Name, registerReq.Desc),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pipecd_piped.test", "id", pipedID),
					resource.TestCheckResourceAttr("pipecd_piped.test", "name", registerReq.Name),
					resource.TestCheckResourceAttr("pipecd_piped.test", "description", registerReq.Desc),
					resource.TestCheckResourceAttr("pipecd_piped.test", "api_key", pipedAPIKey),
				),
			},
		},
	})
}

func testAccResourcePiped(name, desc string) string {
	return providerConfig + fmt.Sprintf(`
resource "pipecd_piped" "test" {
	name = "%s"
	description = "%s"
}`, name, desc)
}
