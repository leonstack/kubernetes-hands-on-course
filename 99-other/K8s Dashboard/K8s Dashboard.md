# 使用 Bearer Token 登录 K8s Dashboard

## 需求

本文档介绍如何通过 RBAC（基于角色的访问控制）机制创建只读权限的 ServiceAccount，并使用 Bearer Token 认证方式安全访问 Kubernetes Dashboard，实现对集群资源的只读监控和查看功能。

## 快速开始

可以使用如下脚本：

```bash
# 一键部署只读用户
./setup-dashboard-readonly.sh

# 清理资源（可选）
./cleanup-dashboard-readonly.sh
```

## 解决方案

### 1. 创建只读用户和角色绑定

#### 1.1 创建 ServiceAccount

```bash
# 创建专用的 ServiceAccount
kubectl create serviceaccount dashboard-readonly-user -n kube-system
```

#### 1.2 创建 ClusterRole（只读权限）

创建 `dashboard-readonly-clusterrole.yaml` 文件：

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: dashboard-readonly
rules:
- apiGroups: [""]
  resources: ["*"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["apps"]
  resources: ["*"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["extensions"]
  resources: ["*"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["networking.k8s.io"]
  resources: ["*"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["storage.k8s.io"]
  resources: ["*"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["batch"]
  resources: ["*"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["autoscaling"]
  resources: ["*"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["policy"]
  resources: ["*"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["rbac.authorization.k8s.io"]
  resources: ["*"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["metrics.k8s.io"]
  resources: ["*"]
  verbs: ["get", "list", "watch"]
```

应用 ClusterRole：

```bash
kubectl apply -f dashboard-readonly-clusterrole.yaml
```

#### 1.3 创建 ClusterRoleBinding

创建 `dashboard-readonly-clusterrolebinding.yaml` 文件：

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: dashboard-readonly
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: dashboard-readonly
subjects:
- kind: ServiceAccount
  name: dashboard-readonly-user
  namespace: kube-system
```

应用 ClusterRoleBinding：

```bash
kubectl apply -f dashboard-readonly-clusterrolebinding.yaml
```

### 2. 获取 Bearer Token

#### 2.1 方法一：使用 kubectl 直接获取（推荐）

> **版本要求**：此方法需要 kubectl 1.24+ 和 Kubernetes 1.24+ 版本支持

```bash
# 获取 ServiceAccount 的 Token
kubectl -n kube-system create token dashboard-readonly-user
```

#### 2.2 方法二：创建长期有效的 Secret Token

创建 `dashboard-readonly-secret.yaml` 文件：

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: dashboard-readonly-user-token
  namespace: kube-system
  annotations:
    kubernetes.io/service-account.name: dashboard-readonly-user
type: kubernetes.io/service-account-token
```

应用 Secret：

```bash
kubectl apply -f dashboard-readonly-secret.yaml
```

获取 Token：

```bash
# 获取 Token
kubectl -n kube-system get secret dashboard-readonly-user-token -o jsonpath='{.data.token}' | base64 -d
```

### 3. 核心原理

#### 3.1 RBAC 权限控制机制

Kubernetes 使用基于角色的访问控制（RBAC）来管理用户权限：

- **ServiceAccount**：为应用程序提供身份标识
- **ClusterRole**：定义集群级别的权限规则
- **ClusterRoleBinding**：将 ServiceAccount 与 ClusterRole 绑定
- **Secret Token**：提供长期有效的认证凭据

#### 3.2 只读权限设计

只读权限通过限制 RBAC 动词实现：

```yaml
verbs: ["get", "list", "watch"]
```

- `get`：获取单个资源
- `list`：列出资源集合
- `watch`：监听资源变化

排除了 `create`、`update`、`patch`、`delete` 等修改操作。

#### 3.3 Token 认证机制

Kubernetes 支持两种 Token 获取方式：

1. **临时 Token**（kubectl 1.24+）：

   ```bash
   kubectl create token dashboard-readonly-user
   ```

   - 默认 24 小时有效期
   - 更安全，自动过期

2. **长期有效 Token**（兼容所有版本）：

   ```bash
   kubectl get secret dashboard-readonly-user-token -o jsonpath='{.data.token}' | base64 -d
   ```

   - 永久有效，直到手动删除
   - 兼容性更好

### 4. 一键部署脚本

使用提供的脚本可以自动完成所有配置：

```bash
# 运行部署脚本
./setup-dashboard-readonly.sh
```

脚本会自动执行以下操作：

1. 创建 ServiceAccount
2. 创建只读 ClusterRole
3. 创建 ClusterRoleBinding
4. 创建长期有效的 Secret Token
5. 等待 Secret 准备就绪并自动获取 Token
6. 显示使用说明

### 5. 清理脚本

当不再需要只读访问时，可以使用清理脚本移除所有相关资源：

```bash
# 运行清理脚本
./cleanup-dashboard-readonly.sh
```

清理脚本会安全地删除：

- ClusterRoleBinding
- ClusterRole
- Secret Token
- ServiceAccount

并验证清理结果，确保所有资源都已正确移除。

### 6. 获取 Token 并登录

#### 6.1 获取 Token

部署脚本运行后会自动显示长期有效的 Token。如果需要手动获取：

```bash
# 获取长期有效 Token
kubectl -n kube-system get secret dashboard-readonly-user-token -o jsonpath='{.data.token}' | base64 -d
```

#### 6.2 访问 Dashboard

##### 方式一：kubectl proxy

1. 启动 kubectl proxy：

   ```bash
   kubectl proxy
   ```

2. 在浏览器中访问：

   ```text
   http://localhost:8001/api/v1/namespaces/kubernetes-dashboard/services/https:kubernetes-dashboard:/proxy/
   ```

**优势**：

- 安全性高，流量通过 kubectl 加密传输
- 无需暴露 Dashboard 端口到外网
- 自动处理 TLS 证书验证

##### 方式二：NodePort Service

如果需要从集群外部直接访问，可以创建 NodePort Service：

```bash
# 创建 NodePort Service
kubectl patch svc kubernetes-dashboard -n kubernetes-dashboard -p '{"spec":{"type":"NodePort"}}'

# 获取访问端口
kubectl get svc kubernetes-dashboard -n kubernetes-dashboard
```

然后通过 `https://<节点IP>:<NodePort>` 访问。

**注意事项**：

- 需要处理 TLS 证书警告
- 暴露端口存在安全风险
- 生产环境建议使用 Ingress + TLS

##### 登录步骤

无论使用哪种访问方式：

1. 选择 "Token" 登录方式
2. 粘贴上面获取的 Token
3. 点击 "Sign in"

### 7. 验证权限

登录后，你应该能够：

- ✅ 查看所有 Namespace 的资源
- ✅ 查看 Pods、Services、Deployments 等
- ✅ 查看日志和事件
- ❌ 无法创建、修改或删除任何资源

### 8. 故障排除

#### 8.1 kubectl create token 命令不支持

**错误信息**：`error: unknown command "token dashboard-readonly-user"`

**原因**：较旧版本的 kubectl（< 1.24）不支持 `create token` 命令。

**解决方案**：

1. 升级 kubectl 到 1.24+ 版本
2. 或者使用长期有效的 Secret Token 方法：

   ```bash
   kubectl -n kube-system get secret dashboard-readonly-user-token -o jsonpath='{.data.token}' | base64 -d
   ```

#### 8.2 Token 无效或过期

```bash
# 重新生成 Token（kubectl 1.24+）
kubectl -n kube-system create token dashboard-readonly-user

# 或获取长期有效 Token
kubectl -n kube-system get secret dashboard-readonly-user-token -o jsonpath='{.data.token}' | base64 -d
```

#### 8.3 Secret Token 不存在或为空

**检查 Secret 状态**：

```bash
kubectl -n kube-system get secret dashboard-readonly-user-token
kubectl -n kube-system describe secret dashboard-readonly-user-token
```

**重新创建 Secret**：

```bash
kubectl delete secret dashboard-readonly-user-token -n kube-system
kubectl apply -f - <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: dashboard-readonly-user-token
  namespace: kube-system
  annotations:
    kubernetes.io/service-account.name: dashboard-readonly-user
type: kubernetes.io/service-account-token
EOF
```

#### 8.4 权限不足

检查 ClusterRoleBinding 是否正确：

```bash
kubectl get clusterrolebinding dashboard-readonly -o yaml
```

#### 8.5 Dashboard 无法访问

确保 Dashboard 已部署并且 kubectl proxy 正在运行：

```bash
# 检查 Dashboard Pod
kubectl get pods -n kubernetes-dashboard

# 启动 proxy
kubectl proxy
```

## 兼容性说明

### kubectl 版本兼容性

| kubectl 版本 | create token 命令 | 推荐方法 |
|-------------|------------------|----------|
| < 1.24 | ❌ 不支持 | 使用 Secret Token |
| >= 1.24 | ✅ 支持 | 使用临时 Token（推荐）|

### 检查 kubectl 版本

```bash
kubectl version --client
```

### 不同版本的获取 Token 方法

**kubectl 1.24+ （推荐）**：

```bash
# 临时 Token（24小时有效）
kubectl -n kube-system create token dashboard-readonly-user
```

**kubectl < 1.24 或兼容性考虑**：

```bash
# 长期有效 Token
kubectl -n kube-system get secret dashboard-readonly-user-token -o jsonpath='{.data.token}' | base64 -d
```

## 安全注意事项

1. **Token 安全**：请妥善保管 Bearer Token，不要在不安全的环境中使用
2. **最小权限原则**：只读权限已经是最小必要权限
3. **Token 轮换**：建议定期轮换 Token，使用临时 Token 更安全
4. **网络安全**：确保 Dashboard 访问通过安全的网络连接
5. **版本兼容性**：根据 kubectl 版本选择合适的 Token 获取方法

## 总结

通过以上步骤，你已经成功创建了一个只读的 ServiceAccount 和相应的 Bearer Token，可以安全地访问 Kubernetes Dashboard 查看集群资源，而不会有误操作的风险。
