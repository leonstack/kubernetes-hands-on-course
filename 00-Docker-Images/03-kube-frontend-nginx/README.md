# Kubernetes 前端 Nginx 代理镜像构建指南

## 📋 项目概述

本项目创建一个优化的 Nginx 前端代理 Docker 镜像，用于 Kubernetes 环境中的反向代理服务。该镜像遵循安全最佳实践，包含健康检查、非 root 用户运行等特性。

## 🔧 步骤01: 前置准备

### 必要条件

- Docker 环境已安装
- Docker Hub 账户（用于镜像推送）
- 基本的 Kubernetes 和 Nginx 知识

### 账户设置

- 创建你的 Docker Hub 账户：<https://hub.docker.com/>
- **重要提示**: 将下面命令中的 `grissomsh` 替换为你的 Docker Hub 账户ID

## 🐳 步骤02: Dockerfile 分析

我们使用了优化的 Dockerfile，具有以下特性：

```dockerfile
# 使用官方nginx镜像的特定版本，提高安全性和稳定性
FROM nginx:1.25-alpine

# 设置维护者信息
LABEL maintainer="your-email@example.com" \
      description="Frontend Nginx proxy for Kubernetes demo" \
      version="1.0"

# 创建非root用户提高安全性
RUN addgroup -g 1001 -S nginx-user && \
    adduser -S -D -H -u 1001 -h /var/cache/nginx -s /sbin/nologin -G nginx-user -g nginx-user nginx-user

# 复制nginx配置文件
COPY default.conf /etc/nginx/conf.d/default.conf

# 创建必要的目录并设置权限
RUN mkdir -p /var/cache/nginx/client_temp \
             /var/cache/nginx/proxy_temp \
             /var/cache/nginx/fastcgi_temp \
             /var/cache/nginx/uwsgi_temp \
             /var/cache/nginx/scgi_temp && \
    chown -R nginx-user:nginx-user /var/cache/nginx && \
    chown -R nginx-user:nginx-user /var/log/nginx && \
    chown -R nginx-user:nginx-user /etc/nginx/conf.d && \
    touch /var/run/nginx.pid && \
    chown nginx-user:nginx-user /var/run/nginx.pid

# 暴露端口
EXPOSE 80

# 切换到非root用户
USER nginx-user

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost/ || exit 1

# 启动nginx
CMD ["nginx", "-g", "daemon off;"]
```

### 🔒 安全特性

- **Alpine Linux 基础镜像**：更小的攻击面
- **非 root 用户运行**：提高容器安全性
- **特定版本标签**：避免意外的镜像更新
- **最小权限原则**：只授予必要的文件权限

## ⚙️ 步骤03: Nginx 配置文件

优化后的 `default.conf` 配置文件：

```nginx
server {
    listen       80;
    server_name  localhost;
    
    location / {
        # 代理到后端Kubernetes服务
        proxy_pass http://my-backend-service:8080;
        
        # 设置代理头信息
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
    
    # 健康检查端点
    location /health {
        return 200 "healthy";
        add_header Content-Type text/plain;
    }
    
    # 错误页面配置
    error_page   500 502 503 504  /50x.html;
    location = /50x.html {
        root   /usr/share/nginx/html;
    }
}
```

### 🚀 配置特性

- **反向代理**：将请求转发到后端服务
- **代理头设置**：保留客户端真实信息
- **健康检查端点**：`/health` 用于 Kubernetes 探针
- **错误页面处理**：标准的 50x 错误处理

### 📝 配置说明

- 将 `my-backend-service:8080` 替换为你的实际后端服务名称和端口
- 健康检查端点可用于 Kubernetes liveness 和 readiness 探针

## 🔨 步骤04: 构建 Docker 镜像

### 基本构建命令

```bash
# 构建 Docker 镜像（使用示例账户）
docker build -t grissomsh/kube-frontend-nginx:1.0.0 .

# 替换为你的 Docker Hub 账户ID
docker build -t <your-docker-hub-id>/kube-frontend-nginx:1.0.0 .

# 同时创建 latest 标签
docker build -t <your-docker-hub-id>/kube-frontend-nginx:1.0.0 \
             -t <your-docker-hub-id>/kube-frontend-nginx:latest .
```

### 🧪 本地测试

```bash
# 运行容器进行本地测试
docker run -d --name nginx-test \
  -p 8080:80 \
  <your-docker-hub-id>/kube-frontend-nginx:1.0.0

# 测试健康检查端点
curl http://localhost:8080/health

# 查看容器日志
docker logs nginx-test

# 清理测试容器
docker stop nginx-test && docker rm nginx-test
```

## 📤 步骤05: 推送到 Docker Hub

### 登录和推送

```bash
# 登录 Docker Hub
docker login

# 推送镜像到 Docker Hub
docker push <your-docker-hub-id>/kube-frontend-nginx:1.0.0
docker push <your-docker-hub-id>/kube-frontend-nginx:latest
```

### ✅ 验证推送

- 登录 Docker Hub 验证镜像：<https://hub.docker.com/repositories>
- 检查镜像标签和大小
- 确认镜像描述和文档

## 🚀 步骤06: Kubernetes 部署

### 示例 Deployment 配置

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: frontend-nginx
spec:
  replicas: 2
  selector:
    matchLabels:
      app: frontend-nginx
  template:
    metadata:
      labels:
        app: frontend-nginx
    spec:
      containers:
      - name: nginx
        image: <your-docker-hub-id>/kube-frontend-nginx:1.0.0
        ports:
        - containerPort: 80
        livenessProbe:
          httpGet:
            path: /health
            port: 80
          initialDelaySeconds: 10
          periodSeconds: 30
        readinessProbe:
          httpGet:
            path: /health
            port: 80
          initialDelaySeconds: 5
          periodSeconds: 10
        resources:
          requests:
            memory: "64Mi"
            cpu: "50m"
          limits:
            memory: "128Mi"
            cpu: "100m"
```

### Service 配置

```yaml
apiVersion: v1
kind: Service
metadata:
  name: frontend-nginx-service
spec:
  selector:
    app: frontend-nginx
  ports:
  - port: 80
    targetPort: 80
  type: LoadBalancer
```

## 🔧 最佳实践

### 安全建议

- ✅ 使用非 root 用户运行容器
- ✅ 定期更新基础镜像
- ✅ 扫描镜像漏洞
- ✅ 使用特定版本标签而非 `latest`

### 性能优化

- ✅ 使用 Alpine Linux 减小镜像大小
- ✅ 合并 RUN 指令减少层数
- ✅ 配置适当的资源限制
- ✅ 启用 Nginx 缓存（如需要）

### 监控和日志

- ✅ 配置健康检查端点
- ✅ 设置适当的探针
- ✅ 收集和分析 Nginx 访问日志
- ✅ 监控容器资源使用情况

## 🐛 故障排除

### 常见问题

**问题 1**: 容器启动失败

```bash
# 检查容器日志
docker logs <container-id>

# 检查权限问题
docker exec -it <container-id> ls -la /var/cache/nginx
```

**问题 2**: 代理连接失败

```bash
# 检查后端服务连通性
kubectl exec -it <pod-name> -- nslookup my-backend-service

# 检查 Nginx 配置
kubectl exec -it <pod-name> -- nginx -t
```

**问题 3**: 健康检查失败

```bash
# 手动测试健康检查
kubectl exec -it <pod-name> -- curl http://localhost/health

# 检查端口监听
kubectl exec -it <pod-name> -- netstat -tlnp
```

## 📚 参考资源

- [Nginx 官方文档](https://nginx.org/en/docs/)
- [Docker 最佳实践](https://docs.docker.com/develop/dev-best-practices/)
- [Kubernetes 部署指南](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/)
- [容器安全最佳实践](https://kubernetes.io/docs/concepts/security/)

## 📝 更新日志

- **v1.0.0**: 初始版本，包含基本代理功能和安全优化
- 使用 Alpine Linux 基础镜像
- 添加非 root 用户支持
- 集成健康检查端点
- 优化权限设置
