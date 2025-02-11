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

package mtl

import (
	"net"
	"net/http"

	"github.com/cloudwego/kitex/pkg/registry"
	"github.com/cloudwego/kitex/server"
	consul "github.com/kitex-contrib/registry-consul"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var Registry *prometheus.Registry

// InitMetric 初始化 Prometheus 指标收集，并通过 HTTP 端口暴露指标数据。
// 参数说明：
//
//	serviceName: 当前应用或服务的名称，用于给监控指标打标签，区分数据来源。
//	metricsPort: 指定 Prometheus 指标暴露的端口（例如 ":8080"）。
//	registryAddr: Consul 服务注册中心的地址，用于服务发现和注册，便于系统中其他服务发现此服务。
func InitMetric(serviceName string, metricsPort string, registryAddr string) {
	// --------------------------------------------------------------------
	// 1. 创建一个全新的 Prometheus 指标注册器 (Registry)
	//    这个 Registry 是一个容器，用于收集和保存程序中注册的所有监控指标。
	Registry = prometheus.NewRegistry()

	// --------------------------------------------------------------------
	// 2. 注册 Go 运行时监控指标
	//    这些指标由 NewGoCollector() 提供，包含了诸如垃圾回收、goroutines 数量等信息，
	//    有助于监控 Go 程序运行时的基础表现。
	Registry.MustRegister(collectors.NewGoCollector())

	// --------------------------------------------------------------------
	// 3. 注册进程相关的监控指标
	//    NewProcessCollector() 会收集操作系统级别的指标，例如进程启动时间、打开的文件描述符数量等，
	//    方便了解当前进程的资源使用情况。
	Registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	// --------------------------------------------------------------------
	// 4. 初始化 Consul 客户端，用于服务注册
	//    通过 Consul 注册服务，可以使其他系统或组件发现这个监控服务。
	//    registryAddr 参数指定了 Consul 服务地址。
	r, _ := consul.NewConsulRegister(registryAddr)

	// --------------------------------------------------------------------
	// 5. 将传入的 metricsPort 弄成一个 TCP 地址对象
	//    这个步骤将字符串形式的端口（例如 ":8080"）转换成 net.TCPAddr 类型，
	//    以便后续在服务注册时能使用正确的地址格式。
	addr, _ := net.ResolveTCPAddr("tcp", metricsPort)

	// --------------------------------------------------------------------
	// 6. 构建一个服务注册信息结构体（registry.Info）
	//    在这里，设置 Prometheus 服务的相关信息：
	//      - ServiceName: 注册时用的服务名称，这里设为 "prometheus"。
	//      - Addr: 监听指标的地址，通过上一步转换后得到。
	//      - Weight: 服务权重（默认为 1），有助于负载均衡的决策。
	//      - Tags: 这里附加一个标签 "service"，其值为传入的 serviceName，便于区分不同应用的监控数据。
	registryInfo := &registry.Info{
		ServiceName: "prometheus",                              // 服务名称，标明这是 Prometheus 服务
		Addr:        addr,                                      // 监听地址，前面通过 net.ResolveTCPAddr 解析得到
		Weight:      1,                                         // 默认权重，通常用于负载均衡策略
		Tags:        map[string]string{"service": serviceName}, // 附加标签标明当前应用的名称
	}

	// --------------------------------------------------------------------
	// 7. 将构建好的服务信息注册到 Consul 中
	//    这样其他服务可以通过 Consul 发现并访问 Prometheus 的指标接口。
	_ = r.Register(registryInfo)

	// --------------------------------------------------------------------
	// 8. 注册一个服务关闭钩子
	//    当应用关闭时，此回调函数会自动执行，目的是将 Prometheus 服务从 Consul 中注销，避免留下无效的注册信息。
	server.RegisterShutdownHook(func() {
		r.Deregister(registryInfo) //nolint:errcheck 忽略错误检查，仅作演示
	})

	// --------------------------------------------------------------------
	// 9. 将 Prometheus 的 HTTP 请求处理器挂载到 "/metrics" 路径
	//    通过 promhttp.HandlerFor 函数，将 Registry 中已注册的所有指标暴露出来，
	//    这样外部 Prometheus 服务器在访问该路径时，就可以获取到最新的监控数据。
	http.Handle("/metrics", promhttp.HandlerFor(Registry, promhttp.HandlerOpts{}))

	// --------------------------------------------------------------------
	// 10. 启动一个 HTTP 服务器监听指定的端口，用于提供 "/metrics" 接口
	//      此处使用 go 关键字启动一个 goroutine，不会阻塞主线程，使得 HTTP 服务在后台运行
	go http.ListenAndServe(metricsPort, nil) //nolint:errcheck 忽略错误检查，实际使用中应进行错误处理
}
