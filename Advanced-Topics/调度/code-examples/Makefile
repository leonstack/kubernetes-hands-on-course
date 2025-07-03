# Makefile
# Kubernetes 调度器高级实践工具构建管理

# 变量定义
REGISTRY ?= localhost:5000
TAG ?= latest
GO_VERSION ?= 1.21
KUBECTL_VERSION ?= v1.28.0

# 工具列表
TOOLS := tenant-resource-manager scheduler-audit-analyzer scheduler-visualizer heatmap-generator performance-analyzer

# 颜色定义
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[1;33m
BLUE := \033[0;34m
NC := \033[0m

# 默认目标
.DEFAULT_GOAL := help

# 帮助信息
.PHONY: help
help: ## 显示帮助信息
	@echo "$(BLUE)Kubernetes 调度器高级实践工具构建系统$(NC)"
	@echo ""
	@echo "$(YELLOW)可用目标:$(NC)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "$(YELLOW)变量:$(NC)"
	@echo "  $(GREEN)REGISTRY$(NC)        镜像仓库地址 (当前: $(REGISTRY))"
	@echo "  $(GREEN)TAG$(NC)             镜像标签 (当前: $(TAG))"
	@echo "  $(GREEN)GO_VERSION$(NC)      Go 版本 (当前: $(GO_VERSION))"
	@echo ""
	@echo "$(YELLOW)示例:$(NC)"
	@echo "  make build-all                    # 构建所有工具"
	@echo "  make build TOOL=performance-analyzer  # 构建特定工具"
	@echo "  make deploy-all REGISTRY=docker.io/myorg TAG=v1.0.0  # 构建并部署"

# 环境检查
.PHONY: check-env
check-env: ## 检查构建环境
	@echo "$(BLUE)检查构建环境...$(NC)"
	@command -v go >/dev/null 2>&1 || { echo "$(RED)错误: Go 未安装$(NC)"; exit 1; }
	@command -v docker >/dev/null 2>&1 || { echo "$(RED)错误: Docker 未安装$(NC)"; exit 1; }
	@echo "$(GREEN)✓ Go 版本: $$(go version)$(NC)"
	@echo "$(GREEN)✓ Docker 版本: $$(docker --version)$(NC)"
	@if command -v kubectl >/dev/null 2>&1; then \
		echo "$(GREEN)✓ kubectl 版本: $$(kubectl version --client --short 2>/dev/null || kubectl version --client)$(NC)"; \
	else \
		echo "$(YELLOW)⚠ kubectl 未安装 (部署功能不可用)$(NC)"; \
	fi

# 依赖管理
.PHONY: deps
deps: ## 下载 Go 依赖
	@echo "$(BLUE)下载 Go 依赖...$(NC)"
	go mod download
	go mod tidy
	@echo "$(GREEN)✓ 依赖下载完成$(NC)"

# 代码检查
.PHONY: lint
lint: ## 运行代码检查
	@echo "$(BLUE)运行代码检查...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "$(YELLOW)golangci-lint 未安装，使用 go vet$(NC)"; \
		go vet ./...; \
	fi
	@echo "$(GREEN)✓ 代码检查完成$(NC)"

# 格式化代码
.PHONY: fmt
fmt: ## 格式化代码
	@echo "$(BLUE)格式化代码...$(NC)"
	go fmt ./...
	@echo "$(GREEN)✓ 代码格式化完成$(NC)"

# 运行测试
.PHONY: test
test: ## 运行测试
	@echo "$(BLUE)运行测试...$(NC)"
	go test -v -race -coverprofile=coverage.out ./...
	@echo "$(GREEN)✓ 测试完成$(NC)"

# 测试覆盖率
.PHONY: coverage
coverage: test ## 生成测试覆盖率报告
	@echo "$(BLUE)生成覆盖率报告...$(NC)"
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✓ 覆盖率报告生成: coverage.html$(NC)"

# 清理
.PHONY: clean
clean: ## 清理构建产物
	@echo "$(BLUE)清理构建产物...$(NC)"
	go clean -cache -modcache -testcache
	docker builder prune -f
	docker image prune -f
	rm -f coverage.out coverage.html
	@echo "$(GREEN)✓ 清理完成$(NC)"

# 构建单个工具
.PHONY: build
build: check-env ## 构建单个工具 (使用 TOOL=工具名)
	@if [ -z "$(TOOL)" ]; then \
		echo "$(RED)错误: 请指定 TOOL 变量$(NC)"; \
		echo "可用工具: $(TOOLS)"; \
		exit 1; \
	fi
	@echo "$(BLUE)构建 $(TOOL)...$(NC)"
	./build.sh -r $(REGISTRY) -t $(TAG) $(TOOL)
	@echo "$(GREEN)✓ $(TOOL) 构建完成$(NC)"

# 构建所有工具
.PHONY: build-all
build-all: check-env ## 构建所有工具
	@echo "$(BLUE)构建所有工具...$(NC)"
	./build.sh -r $(REGISTRY) -t $(TAG) --all
	@echo "$(GREEN)✓ 所有工具构建完成$(NC)"

# 推送单个工具
.PHONY: push
push: ## 推送单个工具镜像 (使用 TOOL=工具名)
	@if [ -z "$(TOOL)" ]; then \
		echo "$(RED)错误: 请指定 TOOL 变量$(NC)"; \
		exit 1; \
	fi
	@echo "$(BLUE)推送 $(TOOL)...$(NC)"
	./build.sh -r $(REGISTRY) -t $(TAG) --push $(TOOL)
	@echo "$(GREEN)✓ $(TOOL) 推送完成$(NC)"

# 推送所有工具
.PHONY: push-all
push-all: ## 推送所有工具镜像
	@echo "$(BLUE)推送所有工具...$(NC)"
	./build.sh -r $(REGISTRY) -t $(TAG) --push --all
	@echo "$(GREEN)✓ 所有工具推送完成$(NC)"

# 部署单个工具
.PHONY: deploy
deploy: ## 部署单个工具 (使用 TOOL=工具名)
	@if [ -z "$(TOOL)" ]; then \
		echo "$(RED)错误: 请指定 TOOL 变量$(NC)"; \
		exit 1; \
	fi
	@echo "$(BLUE)部署 $(TOOL)...$(NC)"
	./build.sh -r $(REGISTRY) -t $(TAG) --deploy $(TOOL)
	@echo "$(GREEN)✓ $(TOOL) 部署完成$(NC)"

# 部署所有工具
.PHONY: deploy-all
deploy-all: ## 部署所有工具
	@echo "$(BLUE)部署所有工具...$(NC)"
	./build.sh -r $(REGISTRY) -t $(TAG) --deploy --all
	@echo "$(GREEN)✓ 所有工具部署完成$(NC)"

# 卸载工具
.PHONY: undeploy
undeploy: ## 卸载工具 (使用 TOOL=工具名)
	@if [ -z "$(TOOL)" ]; then \
		echo "$(RED)错误: 请指定 TOOL 变量$(NC)"; \
		exit 1; \
	fi
	@echo "$(BLUE)卸载 $(TOOL)...$(NC)"
	@if [ -f "$(TOOL)-deployment.yaml" ]; then \
		kubectl delete -f $(TOOL)-deployment.yaml --ignore-not-found=true; \
		echo "$(GREEN)✓ $(TOOL) 卸载完成$(NC)"; \
	else \
		echo "$(YELLOW)⚠ 部署文件 $(TOOL)-deployment.yaml 不存在$(NC)"; \
	fi

# 卸载所有工具
.PHONY: undeploy-all
undeploy-all: ## 卸载所有工具
	@echo "$(BLUE)卸载所有工具...$(NC)"
	@for tool in $(TOOLS); do \
		if [ -f "$$tool-deployment.yaml" ]; then \
			echo "卸载 $$tool..."; \
			kubectl delete -f $$tool-deployment.yaml --ignore-not-found=true; \
		fi; \
	done
	@echo "$(GREEN)✓ 所有工具卸载完成$(NC)"

# 查看状态
.PHONY: status
status: ## 查看部署状态
	@echo "$(BLUE)查看部署状态...$(NC)"
	@echo "$(YELLOW)Deployments:$(NC)"
	@kubectl get deployments -n kube-system -l component=scheduler-monitoring 2>/dev/null || echo "无部署"
	@echo ""
	@echo "$(YELLOW)Services:$(NC)"
	@kubectl get services -n kube-system -l component=scheduler-monitoring 2>/dev/null || echo "无服务"
	@echo ""
	@echo "$(YELLOW)Pods:$(NC)"
	@kubectl get pods -n kube-system -l component=scheduler-monitoring 2>/dev/null || echo "无 Pod"

# 查看日志
.PHONY: logs
logs: ## 查看工具日志 (使用 TOOL=工具名)
	@if [ -z "$(TOOL)" ]; then \
		echo "$(RED)错误: 请指定 TOOL 变量$(NC)"; \
		exit 1; \
	fi
	@echo "$(BLUE)查看 $(TOOL) 日志...$(NC)"
	kubectl logs -n kube-system -l app=$(TOOL) --tail=100 -f

# 端口转发
.PHONY: port-forward
port-forward: ## 端口转发访问服务 (使用 TOOL=工具名 PORT=本地端口)
	@if [ -z "$(TOOL)" ]; then \
		echo "$(RED)错误: 请指定 TOOL 变量$(NC)"; \
		exit 1; \
	fi
	@PORT=$${PORT:-8080}; \
	echo "$(BLUE)端口转发 $(TOOL) 到 localhost:$$PORT...$(NC)"; \
	echo "$(GREEN)访问地址: http://localhost:$$PORT$(NC)"; \
	kubectl port-forward -n kube-system svc/$(TOOL) $$PORT:80

# 开发环境设置
.PHONY: dev-setup
dev-setup: ## 设置开发环境
	@echo "$(BLUE)设置开发环境...$(NC)"
	@echo "安装开发工具..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "安装 golangci-lint..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.54.2; \
	fi
	@echo "$(GREEN)✓ 开发环境设置完成$(NC)"

# 生成文档
.PHONY: docs
docs: ## 生成文档
	@echo "$(BLUE)生成文档...$(NC)"
	@if command -v godoc >/dev/null 2>&1; then \
		echo "启动 godoc 服务器..."; \
		echo "$(GREEN)访问地址: http://localhost:6060/pkg/k8s-scheduler-tools/$(NC)"; \
		godoc -http=:6060; \
	else \
		echo "$(YELLOW)godoc 未安装，使用 go doc$(NC)"; \
		go doc -all; \
	fi

# 安全扫描
.PHONY: security
security: ## 运行安全扫描
	@echo "$(BLUE)运行安全扫描...$(NC)"
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "$(YELLOW)gosec 未安装，跳过安全扫描$(NC)"; \
		echo "安装命令: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi

# 性能测试
.PHONY: bench
bench: ## 运行性能测试
	@echo "$(BLUE)运行性能测试...$(NC)"
	go test -bench=. -benchmem ./...
	@echo "$(GREEN)✓ 性能测试完成$(NC)"

# 完整的 CI 流程
.PHONY: ci
ci: deps fmt lint test security ## 运行完整的 CI 流程
	@echo "$(GREEN)✓ CI 流程完成$(NC)"

# 发布流程
.PHONY: release
release: ci build-all push-all ## 运行发布流程
	@echo "$(GREEN)✓ 发布流程完成$(NC)"
	@echo "$(BLUE)镜像已推送到: $(REGISTRY)$(NC)"
	@echo "$(BLUE)标签: $(TAG)$(NC)"

# 快速开始
.PHONY: quickstart
quickstart: ## 快速开始 (构建并部署所有工具)
	@echo "$(BLUE)快速开始...$(NC)"
	make check-env
	make deps
	make build-all
	make deploy-all
	make status
	@echo "$(GREEN)✓ 快速开始完成!$(NC)"
	@echo ""
	@echo "$(YELLOW)访问服务:$(NC)"
	@echo "  make port-forward TOOL=performance-analyzer PORT=8080"
	@echo "  make port-forward TOOL=heatmap-generator PORT=8081"
	@echo "  make port-forward TOOL=scheduler-visualizer PORT=8082"

# 工具特定的快捷方式
.PHONY: performance-analyzer heatmap-generator scheduler-visualizer tenant-resource-manager scheduler-audit-analyzer
performance-analyzer: ## 构建并部署性能分析器
	make build deploy TOOL=performance-analyzer

heatmap-generator: ## 构建并部署热力图生成器
	make build deploy TOOL=heatmap-generator

scheduler-visualizer: ## 构建并部署调度可视化工具
	make build deploy TOOL=scheduler-visualizer

tenant-resource-manager: ## 构建租户资源管理器
	make build TOOL=tenant-resource-manager

scheduler-audit-analyzer: ## 构建审计分析器
	make build TOOL=scheduler-audit-analyzer

# 调试目标
.PHONY: debug
debug: ## 显示调试信息
	@echo "$(BLUE)调试信息:$(NC)"
	@echo "  Registry: $(REGISTRY)"
	@echo "  Tag: $(TAG)"
	@echo "  Go Version: $(GO_VERSION)"
	@echo "  Tools: $(TOOLS)"
	@echo "  Current Directory: $$(pwd)"
	@echo "  Go Env:"
	@go env | grep -E '(GOOS|GOARCH|GOVERSION|GOPATH|GOROOT)'

# 验证部署
.PHONY: verify
verify: ## 验证部署是否正常工作
	@echo "$(BLUE)验证部署...$(NC)"
	@for tool in $(TOOLS); do \
		if kubectl get deployment $$tool -n kube-system >/dev/null 2>&1; then \
			echo "$(GREEN)✓ $$tool 部署存在$(NC)"; \
			if kubectl rollout status deployment/$$tool -n kube-system --timeout=60s >/dev/null 2>&1; then \
				echo "$(GREEN)✓ $$tool 运行正常$(NC)"; \
			else \
				echo "$(RED)✗ $$tool 运行异常$(NC)"; \
			fi; \
		else \
			echo "$(YELLOW)- $$tool 未部署$(NC)"; \
		fi; \
	done

.PHONY: $(TOOLS)