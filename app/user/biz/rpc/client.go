// Copyright 2024 CloudWeGo Authors
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

package rpc

import (
	"sync"

	"github.com/cloudwego/biz-demo/gomall/app/checkout/conf"
	checkoututils "github.com/cloudwego/biz-demo/gomall/app/checkout/utils"
	"github.com/cloudwego/biz-demo/gomall/common/clientsuite"
	"github.com/cloudwego/biz-demo/gomall/rpc_gen/kitex_gen/auth/authservice"
	"github.com/cloudwego/kitex/client"
)

var (
	AuthClient   authservice.Client
	once         sync.Once
	err          error
	registryAddr string
	serviceName  string
	commonSuite  client.Option
)

func InitClient() {
	once.Do(func() {
		registryAddr = conf.GetConf().Registry.RegistryAddress[0]
		serviceName = conf.GetConf().Kitex.Service
		commonSuite = client.WithSuite(clientsuite.CommonGrpcClientSuite{
			CurrentServiceName: serviceName,
			RegistryAddr:       registryAddr,
		})
		initAuthClient()

	})
}

func initAuthClient() {
	AuthClient, err = authservice.NewClient("auth", commonSuite)
	checkoututils.MustHandleError(err)
}
