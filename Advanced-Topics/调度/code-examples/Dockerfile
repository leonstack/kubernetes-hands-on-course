# Dockerfile
# 多阶段构建 Kubernetes 调度器高级实践工具

# 构建阶段
FROM golang:1.23-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的包
RUN apk add --no-cache git ca-certificates tzdata

# 复制 go mod 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建参数，用于指定要构建的工具
ARG TOOL_NAME=performance-analyzer

# 构建二进制文件
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o /app/bin/${TOOL_NAME} \
    ./cmd/${TOOL_NAME}/main.go

# 运行阶段
FROM scratch

# 从构建阶段复制必要文件
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

# 构建参数
ARG TOOL_NAME=performance-analyzer

# 复制二进制文件
COPY --from=builder /app/bin/${TOOL_NAME} /app

# 创建非 root 用户
USER 65534:65534

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["./app", "--health-check"]

# 设置入口点
ENTRYPOINT ["/app"]

# 默认命令
CMD []

# 标签
LABEL maintainer="Kubernetes Scheduler Team" \
      version="1.0.0" \
      description="Kubernetes Scheduler Advanced Practice Tools" \
      org.opencontainers.image.title="k8s-scheduler-tools" \
      org.opencontainers.image.description="Enterprise-grade Kubernetes scheduler monitoring and analysis tools" \
      org.opencontainers.image.version="1.0.0" \
      org.opencontainers.image.created="$(date -u +'%Y-%m-%dT%H:%M:%SZ')" \
      org.opencontainers.image.source="https://github.com/your-org/k8s-scheduler-tools" \
      org.opencontainers.image.licenses="MIT"