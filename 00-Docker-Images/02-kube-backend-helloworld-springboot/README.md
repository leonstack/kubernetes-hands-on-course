# 2. Spring Boot Hello World - Kubernetes 后端服务

## 1. 📋 项目概述

这是一个基于 Spring Boot 的简单 Hello World REST API 服务，专为 Kubernetes 环境设计。该项目演示了如何构建、容器化和部署一个生产就绪的 Spring Boot 应用程序。

### 1.1 🎯 主要特性

- **RESTful API**：提供简单的 Hello World 端点
- **服务器信息**：返回容器主机名信息，便于负载均衡测试
- **容器化**：使用 Docker 进行容器化部署
- **安全优化**：非 root 用户运行，JVM 参数优化
- **健康检查**：内置健康检查端点
- **Kubernetes 就绪**：适配 Kubernetes 环境的配置

## 2. 🏗️ 项目结构

```text
kube-helloworld/
├── src/
│   ├── main/
│   │   ├── java/com/grissomsh/helloworld/
│   │   │   ├── HelloworldApplication.java      # 主应用类
│   │   │   ├── HelloWorldController.java       # REST 控制器
│   │   │   └── serverinfo/
│   │   │       └── ServerInformationService.java # 服务器信息服务
│   │   └── resources/
│   │       └── application.properties          # 应用配置
│   └── test/
│       └── java/com/grissomsh/helloworld/
│           └── HelloworldApplicationTests.java # 测试类
├── Dockerfile                                  # Docker 构建文件
├── pom.xml                                     # Maven 配置
└── README.md                                   # 项目文档
```

## 3. 🚀 快速开始

### 3.1 前置条件

- Java 8 或更高版本
- Maven 3.6+
- Docker（用于容器化）
- Kubernetes 集群（用于部署）

### 3.2 本地开发

#### 3.2.1 克隆项目

```bash
cd /Users/wangtianqing/Project/kubernetes-fundamentals/00-Docker-Images/02-kube-backend-helloworld-springboot/kube-helloworld
```

#### 3.2.2 编译和运行

```bash
# 编译项目
mvn clean compile

# 运行测试
mvn test

# 打包应用
mvn clean package

# 运行应用
java -jar target/hello-world-rest-api.jar

# 或者使用 Maven 插件运行
mvn spring-boot:run
```

#### 3.2.3 测试 API

```bash
# 测试 Hello World 端点
curl http://localhost:8080/hello

# 预期响应
# Hello World V1 LOCAL
```

## 4. 🐳 Docker 容器化

### 4.1 多阶段构建架构

本项目采用多阶段构建 Dockerfile，相比单阶段构建具有显著优势。多阶段构建将应用的编译和运行分离，大幅减少最终镜像大小，提高安全性和部署效率。

#### 4.1.1 阶段1: 构建阶段 (builder)

```dockerfile
FROM maven:3.8.6-openjdk-8-alpine AS builder
```

**职责**：

- 使用官方 Maven 镜像（包含 JDK 8 和 Maven）编译 Java 源码
- 下载和缓存 Maven 依赖
- 执行单元测试（可选）
- 生成可执行的 JAR 文件

**优化特性**：

- **官方镜像**：使用 Maven 官方镜像，避免额外安装 Maven
- **依赖缓存**：先复制 `pom.xml`，利用 Docker 层缓存优化依赖下载
- **离线构建**：使用 `mvn dependency:go-offline` 预下载依赖
- **构建验证**：验证构建产物的存在性

#### 4.1.2 阶段2: 运行阶段 (runtime)

```dockerfile
FROM eclipse-temurin:8-jre-alpine AS runtime
```

**职责**：

- 使用轻量级 JRE 8 环境运行应用（与构建阶段版本一致）
- 配置安全的非 root 用户
- 设置健康检查和监控
- 优化 JVM 参数

### 4.2 完整 Dockerfile 分析

我们的多阶段 Dockerfile 采用了多项最佳实践：

```dockerfile
# =============================================================================
# 多阶段构建 Dockerfile
# 阶段1: 构建阶段 - 使用 Maven 镜像编译源码
# 阶段2: 运行阶段 - 使用轻量级 JRE 镜像运行应用
# =============================================================================

# ===== 构建阶段 =====
# 使用官方 Maven 镜像，包含 JDK 8 和 Maven，避免额外安装
FROM maven:3.8.6-openjdk-8-alpine AS builder

# 设置构建阶段的维护者信息
LABEL stage=builder

# 设置工作目录
WORKDIR /build

# 首先复制 pom.xml 以利用 Docker 缓存层
# 这样当源码改变但依赖不变时，可以重用依赖下载的缓存层
COPY pom.xml .

# 下载依赖（利用缓存层优化）
RUN mvn dependency:go-offline -B

# 复制源码
COPY src ./src

# 编译应用并跳过测试（生产环境建议启用测试）
RUN mvn clean package -DskipTests -B

# 验证构建产物
RUN ls -la /build/target/ && \
    test -f /build/target/hello-world-rest-api.jar

# ===== 运行阶段 =====
# 使用 JRE 8 镜像，与构建阶段的 JDK 版本保持一致
FROM eclipse-temurin:8-jre-alpine AS runtime

# 设置维护者信息和标签
LABEL maintainer="Grissom <wang.tianqing.cn@outlook.com>" \
      description="Spring Boot Hello World Application - Multi-stage Build" \
      version="1.0.0" \
      build-stage="multi-stage" \
      base-image="eclipse-temurin:8-jre-alpine"

# 安装运行时需要的工具（用于健康检查）
RUN apk add --no-cache wget

# 创建非root用户以提高安全性
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 设置工作目录
WORKDIR /app

# 从构建阶段复制编译好的 JAR 文件
COPY --from=builder /build/target/hello-world-rest-api.jar app.jar

# 创建日志目录
RUN mkdir -p /app/logs && \
    chown -R appuser:appgroup /app

# 切换到非root用户
USER appuser

# 暴露端口
EXPOSE 8080

# 添加健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/actuator/health || exit 1

# 优化JVM参数并启动应用
ENTRYPOINT ["java", \
    "-XX:+UseContainerSupport", \
    "-XX:MaxRAMPercentage=75.0", \
    "-XX:+UseG1GC", \
    "-XX:+UseStringDeduplication", \
    "-XX:+PrintGCDetails", \
    "-XX:+PrintGCTimeStamps", \
    "-Xloggc:/app/logs/gc.log", \
    "-Djava.security.egd=file:/dev/./urandom", \
    "-Dspring.profiles.active=${SPRING_PROFILES_ACTIVE:-default}", \
    "-jar", "app.jar"]
```

### 4.3 📊 多阶段构建优势对比

#### 4.3.1 镜像大小对比

| 构建方式 | 镜像大小 | 说明 |
|----------|----------|------|
| 单阶段构建 | ~200MB | 包含完整 JDK + 源码 + Maven 缓存 |
| 多阶段构建 | ~120MB | 仅包含 JRE + 应用 JAR |
| **减少** | **~40%** | **显著减少存储和传输成本** |

#### 4.3.2 安全性提升

| 方面 | 单阶段 | 多阶段 | 改进 |
|------|--------|--------|------|
| 攻击面 | 大（包含编译工具） | 小（仅运行时） | ✅ 减少潜在漏洞 |
| 敏感信息 | 可能包含源码 | 仅包含编译产物 | ✅ 避免源码泄露 |
| 工具链 | 包含 Maven/JDK | 仅包含 JRE | ✅ 减少可利用工具 |

#### 4.3.3 构建效率

| 特性 | 说明 | 优势 |
|------|------|------|
| 层缓存 | 依赖和源码分层 | 🚀 源码变更时重用依赖缓存 |
| 并行构建 | 可并行构建多个阶段 | 🚀 提高 CI/CD 效率 |
| 增量构建 | 智能缓存机制 | 🚀 减少重复构建时间 |

### 4.4 🔒 安全特性

- **非 root 用户**：使用 `appuser` 用户运行应用
- **最小权限**：只授予必要的文件权限
- **安全基础镜像**：使用 Eclipse Temurin 官方镜像
- **JVM 安全**：配置安全的随机数生成器

### 4.5 ⚡ 性能优化

- **容器感知**：`-XX:+UseContainerSupport` 让 JVM 感知容器环境
- **内存管理**：`-XX:MaxRAMPercentage=75.0` 限制内存使用
- **垃圾收集**：使用 G1GC 和字符串去重优化
- **启动优化**：配置快速启动参数

### 4.6 构建和运行 Docker 镜像

#### 4.6.1 基本构建（多阶段）

```bash
# 构建多阶段镜像（无需预先编译）
docker build -t kube-helloworld:multi-stage .

# 查看镜像大小对比
docker images | grep kube-helloworld

# 运行容器
docker run -d -p 8080:8080 --name hello-app kube-helloworld:multi-stage

# 测试应用
curl http://localhost:8080/hello

# 查看容器日志
docker logs hello-app

# 停止和清理
docker stop hello-app
docker rm hello-app
```

#### 4.6.2 构建特定阶段（调试用）

```bash
# 只构建到 builder 阶段（用于调试构建问题）
docker build --target builder -t kube-helloworld:builder .

# 进入 builder 阶段容器查看构建产物
docker run -it kube-helloworld:builder /bin/sh
ls -la /build/target/
```

#### 4.6.3 使用构建参数

```bash
# 启用测试的构建
docker build --build-arg SKIP_TESTS=false -t kube-helloworld:with-tests .

# 指定 Maven 配置
docker build --build-arg MAVEN_OPTS="-Xmx1024m" -t kube-helloworld:optimized .

# 使用 BuildKit 进行并行构建
DOCKER_BUILDKIT=1 docker build -t kube-helloworld:buildkit .
```

#### 高级运行配置

```bash
# 指定 Spring Profile
docker run -d -p 8080:8080 \
  -e SPRING_PROFILES_ACTIVE=production \
  --name hello-app-prod \
  kube-helloworld:multi-stage

# 挂载日志目录
docker run -d -p 8080:8080 \
  -v $(pwd)/logs:/app/logs \
  --name hello-app-with-logs \
  kube-helloworld:multi-stage

# 查看 GC 日志
docker exec hello-app-with-logs tail -f /app/logs/gc.log
```

### 使用 Maven Docker 插件

项目配置了 Spotify 的 dockerfile-maven-plugin：

```bash
# 使用 Maven 构建 Docker 镜像
mvn clean package dockerfile:build

# 推送到仓库（需要先配置仓库）
# mvn dockerfile:push
```

## 🎛️ API 文档

### 端点列表

| 方法 | 路径 | 描述 | 响应示例 |
|------|------|------|----------|
| GET | `/hello` | 返回 Hello World 消息和服务器信息 | `Hello World V1 abc12` |

### 响应格式

```json
{
  "message": "Hello World V1 {server_id}",
  "server_id": "容器主机名的后5位字符"
}
```

## ☸️ Kubernetes 部署

### 基本部署

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: hello-world-deployment
spec:
  replicas: 3
  selector:
    matchLabels:
      app: hello-world
  template:
    metadata:
      labels:
        app: hello-world
    spec:
      containers:
      - name: hello-world
        image: kube-helloworld:1.0.0
        ports:
        - containerPort: 8080
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /hello
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /hello
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: hello-world-service
spec:
  selector:
    app: hello-world
  ports:
  - port: 80
    targetPort: 8080
  type: LoadBalancer
```

### 部署命令

```bash
# 应用部署配置
kubectl apply -f k8s-deployment.yaml

# 查看部署状态
kubectl get deployments
kubectl get pods
kubectl get services

# 测试服务
kubectl port-forward service/hello-world-service 8080:80
curl http://localhost:8080/hello

# 扩缩容
kubectl scale deployment hello-world-deployment --replicas=5

# 查看日志
kubectl logs -l app=hello-world
```

## 🔧 配置说明

### Maven 配置 (pom.xml)

- **Spring Boot 版本**：2.2.7.RELEASE
- **Java 版本**：1.8
- **构建输出**：hello-world-rest-api.jar
- **测试跳过**：`maven.test.skip=true`（生产环境建议启用测试）

### 应用配置 (application.properties)

当前为空配置文件，可以根据需要添加：

```properties
# 服务器配置
server.port=8080
server.servlet.context-path=/

# 日志配置
logging.level.com.stacksimplify=INFO
logging.pattern.console=%d{yyyy-MM-dd HH:mm:ss} - %msg%n

# 健康检查配置
management.endpoints.web.exposure.include=health,info
management.endpoint.health.show-details=always

# 应用信息
info.app.name=Hello World API
info.app.version=1.0.0
info.app.description=Spring Boot Hello World for Kubernetes
```

## 🧪 测试

### 单元测试

```bash
# 运行所有测试
mvn test

# 运行特定测试
mvn test -Dtest=HelloworldApplicationTests

# 生成测试报告
mvn surefire-report:report
```

### 集成测试

```bash
# 启动应用进行集成测试
mvn spring-boot:run &
APP_PID=$!

# 等待应用启动
sleep 10

# 测试 API
curl -f http://localhost:8080/hello || echo "API test failed"

# 停止应用
kill $APP_PID
```

### 负载测试

```bash
# 使用 ab 进行简单负载测试
ab -n 1000 -c 10 http://localhost:8080/hello

# 使用 curl 测试多个实例
for i in {1..10}; do
  curl http://localhost:8080/hello
  echo
done
```

## 📊 监控和日志

### 健康检查

```bash
# Docker 健康检查
docker inspect --format='{{.State.Health.Status}}' hello-app

# Kubernetes 健康检查
kubectl describe pod <pod-name>
```

### 日志收集

```bash
# Docker 日志
docker logs -f hello-app

# Kubernetes 日志
kubectl logs -f deployment/hello-world-deployment

# 聚合日志
kubectl logs -l app=hello-world --tail=100
```

### 性能监控

```bash
# 容器资源使用
docker stats hello-app

# Kubernetes 资源使用
kubectl top pods -l app=hello-world
kubectl top nodes
```

## 🔍 故障排除

### 常见问题

#### 1. 多阶段构建失败

```bash
# 检查 builder 阶段
docker build --target builder -t debug-builder .
docker run -it debug-builder /bin/sh

# 在容器内检查
ls -la /build/
mvn dependency:tree

# 查看构建过程
docker build --progress=plain --no-cache -t kube-helloworld:debug .
```

#### 2. 依赖下载失败

```bash
# 使用国内 Maven 镜像
docker build --build-arg MAVEN_MIRROR=https://maven.aliyun.com/repository/public .

# 检查网络连接
docker run --rm maven:3.8.6-openjdk-8-alpine ping -c 3 repo1.maven.org
```

#### 3. 内存不足问题

```bash
# 增加构建内存
docker build --memory=2g -t kube-helloworld:large-mem .

# 检查系统资源
docker system df
docker system prune
```

#### 4. 应用启动失败

```bash
# 检查 Java 版本
java -version

# 检查 JAR 文件
ls -la target/

# 查看详细启动日志
java -jar target/hello-world-rest-api.jar --debug
```

#### 5. Docker 构建失败

```bash
# 检查 Dockerfile
docker build --no-cache -t kube-helloworld:debug .

# 逐步构建调试
docker build --target <stage> -t debug-image .

# 检查各阶段镜像
docker images --filter "label=stage=builder"

# 比较镜像层
docker history kube-helloworld:multi-stage
```

#### 6. Kubernetes 部署问题

```bash
# 检查部署状态
kubectl describe deployment hello-world-deployment

# 检查 Pod 状态
kubectl describe pod <pod-name>

# 查看事件
kubectl get events --sort-by=.metadata.creationTimestamp
```

### 调试技巧

#### Docker 调试

```bash
# 进入运行中的容器
docker exec -it hello-app /bin/sh

# 检查容器内文件
docker exec hello-app ls -la /app/
docker exec hello-app cat /app/logs/gc.log

# 监控容器资源
docker stats hello-app

# 检查健康状态
docker inspect --format='{{.State.Health.Status}}' hello-app
```

#### Kubernetes 调试

```bash
# 进入 Pod
kubectl exec -it <pod-name> -- /bin/sh

# 检查网络连接
kubectl run debug --image=busybox --rm -it --restart=Never -- /bin/sh
# 在 debug pod 中测试连接
wget -qO- http://hello-world-service/hello

# 查看详细日志
kubectl logs -f <pod-name> --previous
kubectl logs -l app=hello-world --tail=100
```

#### 性能监控

```bash
# 记录构建时间
time docker build -t kube-helloworld:timed .

# 分析构建步骤耗时
docker build --progress=plain -t kube-helloworld:analyzed . 2>&1 | grep "#[0-9]"

# 查看 GC 日志
docker exec hello-app tail -f /app/logs/gc.log

# Kubernetes 资源使用
kubectl top pods -l app=hello-world
kubectl top nodes
```

## 🚀 最佳实践

### 多阶段构建优化

#### 构建效率

```dockerfile
# 利用构建缓存
COPY pom.xml .
RUN mvn dependency:go-offline

# 分层复制源码
COPY src ./src
RUN mvn package -DskipTests
```

#### 镜像安全

```dockerfile
# 使用非 root 用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 设置文件权限
CHOWN appuser:appgroup /app
USER appuser

# 移除不必要的包
RUN apk del .build-deps
```

#### 镜像优化

```bash
# 使用 .dockerignore
echo "target/" >> .dockerignore
echo "*.log" >> .dockerignore
echo ".git" >> .dockerignore

# 压缩镜像层
docker build --squash -t kube-helloworld:compressed .

# 使用 distroless 镜像（生产环境）
FROM gcr.io/distroless/java:8
```

### 开发最佳实践

1. **版本管理**：使用语义化版本控制
2. **配置外部化**：使用 ConfigMap 和 Secret
3. **健康检查**：实现 liveness 和 readiness 探针
4. **资源限制**：设置合理的 CPU 和内存限制
5. **日志结构化**：使用结构化日志格式
6. **代码质量**：遵循 Spring Boot 最佳实践
7. **测试覆盖**：编写单元测试和集成测试

### 安全最佳实践

#### 容器安全

```bash
# 扫描镜像漏洞
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock \
  aquasec/trivy image kube-helloworld:multi-stage

# 使用安全基础镜像
FROM eclipse-temurin:8-jre-alpine
# 或者
FROM gcr.io/distroless/java:8
```

#### 应用安全

1. **非 root 用户**：始终使用非特权用户运行
2. **镜像扫描**：定期扫描镜像漏洞
3. **最小权限**：只授予必要的权限
4. **密钥管理**：使用 Kubernetes Secret 管理敏感信息
5. **配置安全**：不在代码中硬编码敏感信息
6. **依赖管理**：定期更新依赖版本

### 性能优化

#### JVM 调优

```dockerfile
# 生产环境 JVM 参数
ENV JAVA_OPTS="-Xms512m -Xmx1024m -XX:+UseG1GC -XX:+PrintGCDetails -Xloggc:/app/logs/gc.log"
```

#### 应用优化

1. **JVM 参数**：合理配置 JVM 参数
2. **连接池**：使用连接池
3. **缓存策略**：实施缓存策略
4. **性能监控**：监控应用性能
5. **协议优化**：启用 HTTP/2

### 运维最佳实践

#### Docker 部署

```bash
# 使用健康检查
docker run -d \
  --health-cmd="curl -f http://localhost:8080/actuator/health || exit 1" \
  --health-interval=30s \
  --health-timeout=10s \
  --health-retries=3 \
  kube-helloworld:multi-stage
```

#### Kubernetes 部署

1. **监控告警**：设置关键指标监控
2. **备份策略**：制定数据备份计划
3. **滚动更新**：使用滚动更新策略
4. **资源配额**：设置命名空间资源配额
5. **多阶段构建**：使用多阶段构建减小镜像大小
6. **配置管理**：使用 ConfigMap 和 Secret
7. **安全策略**：实施 Pod 安全策略

### CI/CD 集成

#### GitHub Actions 示例

```yaml
name: Build and Deploy
on:
  push:
    branches: [ main ]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Build multi-stage image
      run: |
        docker build -t ${{ github.repository }}:${{ github.sha }} .
        docker build -t ${{ github.repository }}:latest .
    
    - name: Security scan
      run: |
        docker run --rm -v /var/run/docker.sock:/var/run/docker.sock \
          aquasec/trivy image ${{ github.repository }}:${{ github.sha }}
    
    - name: Push to registry
      run: |
        echo ${{ secrets.DOCKER_PASSWORD }} | docker login -u ${{ secrets.DOCKER_USERNAME }} --password-stdin
        docker push ${{ github.repository }}:${{ github.sha }}
        docker push ${{ github.repository }}:latest
```

## 📚 参考资源

- [Spring Boot 官方文档](https://spring.io/projects/spring-boot)
- [Docker 最佳实践](https://docs.docker.com/develop/dev-best-practices/)
- [Kubernetes 部署指南](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/)
- [Eclipse Temurin 镜像](https://hub.docker.com/_/eclipse-temurin)
- [Maven Docker 插件](https://github.com/spotify/dockerfile-maven)

## 📄 许可证

本项目仅用于学习和演示目的。

---

**维护者**: Grissom <wang.tianqing.cn@outlook.com>  
**版本**: 1.0.0  
**最后更新**: 2024年
