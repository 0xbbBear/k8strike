# k8strike
k8strike 是一款面向 Kubernetes 集群的集成化安全评估与渗透测试工具，采用 Go 语言开发，编译为单一静态二进制文件，可在受限容器环境中零依赖运行。

工具包含四个核心模块：
- **evaluate** — 容器安全评估，自动收集系统信息、检测危险 Capability、分析挂载与网络配置
- **run** — 漏洞利用执行，集成容器逃逸、凭证窃取、持久化驻留、远程控制四大类 exploit
- **spider** — K8s 服务发现，基于 DNS 协议无需 API Server 认证即可枚举集群内部服务
- **tool** — 内置渗透工具，提供 Netcat、端口扫描、kubectl 封装、etcd 客户端等常用组件

**项目目录结构**

```
k8strike/
├── cmd/k8strike/main.go                 # 程序入口
├── pkg/
│   ├── cli/                             # CLI 命令层 (Cobra)
│   │   ├── root.go                      #   根命令与横幅
│   │   ├── evaluate_cmd.go              #   evaluate 子命令
│   │   ├── run_cmd.go                   #   run 子命令
│   │   ├── spider_cmd.go                #   spider 子命令
│   │   └── tool_cmd.go                  #   tool 子命令
│   ├── evaluate/                        # 容器安全评估模块
│   │   ├── engine.go                    #   评估引擎
│   │   ├── registry.go                  #   检查项注册表
│   │   ├── categories.go                #   检查类别定义
│   │   └── ... (15 类安全检查)           #   系统信息/能力/挂载/网络等
│   ├── exploit/                         # 漏洞利用模块
│   │   ├── base/base.go                 #   Exploit 基类
│   │   ├── escaping/                    #   容器逃逸 (14 个 exploit)
│   │   ├── credential_access/           #   凭证访问 (5 个 exploit)
│   │   ├── persistence/                 #   持久化攻击 (5 个 exploit)
│   │   ├── remote_control/              #   远程控制 (2 个 exploit)
│   │   └── discovery/                   #   服务发现 (4 个 exploit)
│   ├── plugin/                          # 插件接口定义
│   │   └── interface.go                 #   ExploitInterface 接口
│   ├── spider/                          # K8s 服务发现模块
│   │   ├── core/                        #   命令调度
│   │   ├── dns/                         #   DNS 解析层
│   │   ├── scanner/                     #   扫描引擎 (AXFR/PTR/Wildcard)
│   │   ├── multi/                       #   多线程扫描
│   │   ├── metrics/                     #   Prometheus 指标解析
│   │   ├── nfs/                         #   NFS 共享发现
│   │   └── admission/                   #   Admission Webhook 列举
│   ├── tool/                            # 内置渗透工具
│   │   ├── netcat/                      #   Netcat (TCP/UDP/Shell)
│   │   ├── probe/                       #   端口扫描器
│   │   ├── kubectl/                     #   kubectl API 封装
│   │   ├── etcdctl/                     #   etcd 客户端
│   │   └── dockerd_api/                 #   Docker Socket 请求
│   └── util/                            # 公共工具函数
├── figures/                             # 论文插图 (Mermaid)
├── labs/                                # 实验环境配置
└── README.md
```

**项目声明**
- 项目名称：k8strike
- 项目作者：Li Yuzhi
- 作者单位：暨南大学网络空间安全学院
- 开发语言：Golang
- 核心技术：Kubernetes安全、容器安全、渗透测试、服务发现
