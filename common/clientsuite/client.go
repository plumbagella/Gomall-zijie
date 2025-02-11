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

package clientsuite // 定义当前包名为 clientsuite

import (
	// kitex/client 提供了创建和管理 RPC 客户端的基础功能
	"github.com/cloudwego/kitex/client"
	// kitex/pkg/rpcinfo 用于封装与 RPC 相关的基本信息，例如服务名称
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	// kitex/pkg/transmeta 用来实现跨网络传输的元数据管理，这里用于 HTTP2 协议下的元数据传输
	"github.com/cloudwego/kitex/pkg/transmeta"
	// kitex/transport 定义了底层传输协议相关的内容，例如 TCP、HTTP2，GRPC 等
	"github.com/cloudwego/kitex/transport"
	// kitex-contrib/obs-opentelemetry/tracing 用于添加 OpenTelemetry 的追踪功能，收集调用链信息
	"github.com/kitex-contrib/obs-opentelemetry/tracing"
	// kitex-contrib/registry-consul 提供对 Consul 服务注册中心的支持，便于服务发现
	consul "github.com/kitex-contrib/registry-consul"
)

// CommonGrpcClientSuite 是一个结构体，用于封装创建 GRPC 客户端时的基础配置。
// 这个结构体持有两个关键参数：
//   - CurrentServiceName: 当前服务的名称，用于区分不同的客户端服务调用。
//   - RegistryAddr: Consul 注册中心的地址，用于服务发现，让客户端自动获取调用目标的地址。
type CommonGrpcClientSuite struct {
	CurrentServiceName string // 定义当前客户端调用发起方的服务名称
	RegistryAddr       string // 存储 Consul 注册中心的网络地址（IP:端口格式）
}

// Options 方法返回一个 client.Option 类型的切片，包含了创建客户端所需要的所有参数配置。
// 这些配置主要用于：
//  1. 指定使用 Consul 进行服务发现（Resolver）。
//  2. 设置如何传输元数据（MetaHandler）。
//  3. 配置底层传输协议为 gRPC。
//  4. 添加客户端的基本信息，以及 OpenTelemetry 追踪功能的中间件。
func (s CommonGrpcClientSuite) Options() []client.Option {
	// --------------------------------------------------------------------
	// 1. 使用 Consul 注册中心地址创建一个 Consul Resolver
	//    Consul Resolver：是一个实现了 kitex 客户端取服务地址的组件，它会从 Consul 注册中心中
	//    获取目标服务的地址列表，从而实现服务的自动发现。
	r, err := consul.NewConsulResolver(s.RegistryAddr)
	// 如果在创建 Consul Resolver 过程中有错误，则直接使程序崩溃（panic）。
	// 在实际生产环境中，建议替换为更友好的错误处理机制。
	if err != nil {
		panic(err)
	}

	// --------------------------------------------------------------------
	// 2. 构建一个 client.Option 类型的切片（数组），用来存放所有的客户端配置选项。
	//    这些选项将会在创建客户端对象时传递进去，以确保客户端能正确工作。
	opts := []client.Option{
		// 使用 WithResolver 选项将上面创建的 Consul Resolver 添加到客户端配置中，
		// 这样客户端在发起调用时，会自动从 Consul 注册中心获取目标服务地址。
		client.WithResolver(r),
		// WithMetaHandler 选项用来设置元数据（metadata）处理器，这里指定使用 transmeta.ClientHTTP2Handler，
		// 用于在 HTTP2 协议下传输元数据信息（如请求头、上下文信息等）。
		client.WithMetaHandler(transmeta.ClientHTTP2Handler),
		// WithTransportProtocol 选项指定客户端将使用 gRPC 作为底层传输协议，
		// 使得客户端通信遵循 gRPC 协议标准。
		client.WithTransportProtocol(transport.GRPC),
	}

	// --------------------------------------------------------------------
	// 3. 追加其他客户端配置选项：
	//    使用 append 函数向已有的选项切片中增加其他配置，确保客户端具备必要的基本信息和追踪功能。
	opts = append(opts,
		// WithClientBasicInfo 选项设置客户端调用时的基本信息（EndpointBasicInfo），
		// 其中 ServiceName 字段被设置为当前客户端的服务名称，便于服务端区分调用者身份。
		client.WithClientBasicInfo(&rpcinfo.EndpointBasicInfo{
			ServiceName: s.CurrentServiceName,
		}),
		// WithSuite 选项添加一个中间件套件，这里使用 tracing.NewClientSuite()，
		// 该套件集成了 OpenTelemetry 追踪工具，可以记录客户端的调用链，便于监控和调试。
		client.WithSuite(tracing.NewClientSuite()),
	)

	// --------------------------------------------------------------------
	// 4. 返回配置好的客户端选项集合，后续在创建客户端时会将这些选项传入，
	//    使得客户端能够自动实现服务发现、元数据传输、追踪功能以及使用指定的传输协议工作。
	return opts
}
