# Kubernetes 调度器企业级高级实践工具集

本项目提供了一套完整的 Kubernetes 调度器企业级高级实践工具，包括多租户资源管理、安全审计分析、调度决策可视化、集群资源热力图和性能趋势分析等功能。

> 📖 **完整文档**: 详细的使用指南、API文档、架构说明和最佳实践请参考 [docs/README.md](docs/README.md)

## 项目结构

```bash
code-examples/
├── cmd/                          # 主程序入口
│   ├── heatmap-generator/        # 集群资源热力图生成器
│   │   └── main.go
│   ├── performance-analyzer/     # 调度性能趋势分析器
│   │   ├── main.go
│   │   └── main_entry.go
│   ├── scheduler-analyzer/       # 调度器分析器
│   │   └── main.go
│   ├── scheduler-audit-analyzer/ # 调度器安全审计分析器
│   │   └── main.go
│   ├── scheduler-visualizer/     # 调度决策可视化工具
│   │   └── main.go
│   └── tenant-resource-manager/  # 多租户资源管理器
│       └── main.go
├── pkg/                          # 可复用的包
│   ├── automation/               # 自动化相关包
│   │   └── scheduler-automation.go
│   ├── recovery/                 # 恢复相关包
│   │   ├── health-recovery.go
│   │   └── scheduler-recovery.go
│   ├── scheduler/                # 调度器核心包
│   │   ├── batch-scheduler.go
│   │   ├── custom-preemption.go
│   │   ├── dynamic-resource-quota.go
│   │   ├── edge-scheduler.go
│   │   ├── health-checker.go
│   │   ├── node-resource-optimizer.go
│   │   ├── performance-tuning.go
│   │   ├── recovery-manager.go
│   │   ├── scheduler-analyzer.go
│   │   ├── scheduler-metrics.go
│   │   ├── scheduler-recovery.go
│   │   ├── scheduler-selector.go
│   │   ├── scheduler-troubleshooter.go
│   │   └── workload-classifier.go
│   └── troubleshooter/           # 故障排除包
│       └── scheduler-troubleshooter.go
├── internal/                     # 内部包
│   ├── config/                   # 配置管理
│   ├── metrics/                  # 指标收集
│   └── utils/                    # 工具函数
├── deployments/                  # 部署相关文件
│   ├── docker/                   # Docker 相关文件
│   └── kubernetes/               # Kubernetes 部署文件
│       ├── heatmap-generator-deployment.yaml
│       ├── performance-analyzer-deployment.yaml
│       ├── rbac.yaml
│       ├── scheduler-analyzer-deployment.yaml
│       ├── scheduler-audit-analyzer-deployment.yaml
│       ├── scheduler-visualizer-deployment.yaml
│       └── tenant-resource-manager-deployment.yaml
├── configs/                      # 配置文件
│   ├── config-example.yaml       # 配置示例
│   ├── high-performance-scheduler-config.yaml  # 高性能调度器配置
│   ├── monitoring-config.yaml    # 监控配置
│   ├── preemption-config.yaml    # 抢占配置
│   ├── scheduler-health-config.yaml  # 调度器健康检查配置
│   ├── scheduler-recovery-config.yaml  # 调度器恢复配置
│   ├── monitoring/               # 监控配置
│   │   ├── scheduler-alerts.yaml
│   │   └── scheduler-monitoring.yaml
│   └── scheduler/                # 调度器配置
│       ├── batch-scheduler-config.yaml
│       ├── edge-node-labels.yaml
│       ├── edge-scheduler-config.yaml
│       ├── edge-workload-example.yaml
│       ├── priority-resource-allocation.yaml
│       ├── scheduler-config.yaml
│       ├── scheduler-ha-deployment.yaml
│       ├── scheduler-memory-optimization.yaml
│       └── workload-scheduling-policies.yaml
├── scripts/                      # 脚本文件
│   └── build.sh                  # 构建脚本
├── docs/                         # 文档
│   ├── PROJECT_SUMMARY.md
│   └── README.md
├── bin/                          # 编译输出目录
│   ├── heatmap-generator
│   ├── performance-analyzer
│   ├── scheduler-analyzer
│   ├── scheduler-audit-analyzer
│   ├── scheduler-visualizer
│   └── tenant-resource-manager
├── build-local.sh                # 本地构建脚本
├── build.sh                      # 构建脚本
├── Dockerfile                    # Docker 构建文件
├── go.mod                        # Go 模块文件 (Go 1.23.0)
├── go.sum                        # Go 依赖校验文件
└── Makefile                      # 构建文件
```

## 快速开始

### 环境要求

- Kubernetes 1.28+
- Go 1.23+
- Docker
- kubectl

### 构建项目

```bash
# 检查环境
make check-env

# 下载依赖
make deps

# 构建所有工具
make build-all

# 或者单独构建某个工具
make build TOOL=performance-analyzer
make build TOOL=heatmap-generator
make build TOOL=scheduler-analyzer
make build TOOL=scheduler-visualizer
make build TOOL=tenant-resource-manager
make build TOOL=scheduler-audit-analyzer

# 直接使用 Go 命令构建
go build -o bin/heatmap-generator ./cmd/heatmap-generator
go build -o bin/performance-analyzer ./cmd/performance-analyzer
go build -o bin/scheduler-analyzer ./cmd/scheduler-analyzer
go build -o bin/scheduler-visualizer ./cmd/scheduler-visualizer
go build -o bin/tenant-resource-manager ./cmd/tenant-resource-manager
go build -o bin/scheduler-audit-analyzer ./cmd/scheduler-audit-analyzer

# 运行测试
make test

# 代码格式化
make fmt

# 代码检查
make lint
```

> **注意**: 如果在Docker构建过程中遇到网络连接问题（如TLS握手超时），构建脚本已配置使用 `--network=host` 参数来解决网络连接问题。

### 部署到 Kubernetes

```bash
# 部署所有工具
make deploy-all

# 或者单独部署某个工具
make deploy TOOL=performance-analyzer
make deploy TOOL=heatmap-generator

# 使用自定义镜像仓库和标签
make deploy-all REGISTRY=docker.io/myorg TAG=v1.0.0

# 直接使用 kubectl 部署
kubectl apply -f deployments/kubernetes/

# 应用配置
kubectl apply -f configs/scheduler/
kubectl apply -f configs/monitoring/

# 查看部署状态
make status

# 卸载工具
make undeploy TOOL=performance-analyzer
make undeploy-all
```

### 单个工具操作

```bash
# 构建特定工具
make build TOOL=heatmap-generator

# 部署特定工具
make deploy TOOL=heatmap-generator

# 卸载特定工具
make undeploy TOOL=heatmap-generator

# 查看工具状态
make status

# 查看帮助
make help
```

> 📋 **更多命令**: 完整的Makefile使用指南、环境变量配置和高级选项请参考 [完整文档](docs/README.md#3-快速开始)。

## 工具概览

本项目包含以下6个核心工具：

| 工具 | 功能 | 类型 | 默认端口 | 访问方式 |
|------|------|------|----------|----------|
| **调度器分析器** | 调度器行为分析和性能监控 | Web | 8080 | `http://localhost:8080` |
| **多租户资源管理器** | 多租户资源配额管理和策略控制 | Web | 8080 | `http://localhost:8080` |
| **调度器安全审计分析器** | 调度安全事件分析和合规检查 | Web | 8080 | `http://localhost:8080` |
| **调度决策可视化工具** | 调度决策流程可视化展示 | Web | 8080 | `http://localhost:8080` |
| **集群资源热力图生成器** | 集群资源使用热力图 | Web | 8082 | `http://localhost:8082` |
| **调度性能趋势分析器** | 性能趋势分析和异常检测 | Web | 8081 | `http://localhost:8081` |

> 💡 **提示**: CLI工具适合命令行操作和自动化，Web工具提供可视化界面。详细功能说明请参考 [完整文档](docs/README.md#1-工具概览)。
