# Kubernetes - Update Deployments

## 步骤 00：介绍
- 我们可以使用两种选项更新 Deployment
  - 设置镜像
  - 编辑 Deployment

## 步骤 01：使用 "设置镜像" 选项将应用程序版本从 V1 更新到 V2
### 更新 Deployment
- **观察：** 请检查 `spec.container.name` yaml 输出中的容器名称并记下它，然后在 `kubectl set image` 命令中替换 <Container-Name>
```
# 从当前 Deployment 获取容器名称
kubectl get deployment my-first-deployment -o yaml

# 更新 Deployment - 现在应该可以工作了
kubectl set image deployment/<Deployment-Name> <Container-Name>=<Container-Image> --record=true
kubectl set image deployment/my-first-deployment kubenginx=grissomsh/kubenginx:2.0.0 --record=true
```
### 验证推出状态（Deployment 状态）
- **观察：** 默认情况下，推出以滚动更新模式进行，因此没有停机时间。
```
# 验证推出状态
kubectl rollout status deployment/my-first-deployment

# 验证 Deployment
kubectl get deploy
```
### 描述 Deployment
- **观察：**
  - 验证事件并了解 Kubernetes 默认对新应用程序发布执行 "滚动更新"。
  - 也就是说，我们的应用程序不会有停机时间。
```
# 描述 Deployment
kubectl describe deployment my-first-deployment
```
### 验证 ReplicaSet
- **观察：** 将为新版本创建新的 ReplicaSet
```
# 验证 ReplicaSet
kubectl get rs
```

### 验证 Pod
- **观察：** 新 ReplicaSet 的 Pod 模板哈希标签应该出现在 Pod 上，让我们知道这些 Pod 属于新的 ReplicaSet。
```
# 列出 Pod
kubectl get po
```

### 验证 Deployment 的推出历史
- **观察：** 我们有推出历史，因此我们可以使用可用的修订历史切换回较旧的修订版本。

```
# 检查 Deployment 的推出历史
kubectl rollout history deployment/<Deployment-Name>
kubectl rollout history deployment/my-first-deployment  
```

### 使用公网 IP 访问应用程序
- 每当我们在浏览器中访问应用程序时，应该看到 `Application Version:V2`
```
# 获取 NodePort
kubectl get svc
观察：记下以 3 开头的端口（例如：80:3xxxx/TCP）。捕获端口 3xxxx 并在下面的应用程序 URL 中使用它。

# 获取工作节点的公网 IP
kubectl get nodes -o wide
观察：如果您的 Kubernetes 集群在 AWS EKS 上设置，请记下 "EXTERNAL-IP"。

# 应用程序 URL
http://<worker-node-public-ip>:<Node-Port>
```


## 步骤 02：使用 "编辑 Deployment" 选项将应用程序从 V2 更新到 V3
### 编辑 Deployment
```
# 编辑 Deployment
kubectl edit deployment/<Deployment-Name> --record=true
kubectl edit deployment/my-first-deployment --record=true
```

```yml
# 从 2.0.0 更改
    spec:
      containers:
      - image: grissomsh/kubenginx:2.0.0

# 更改为 3.0.0
    spec:
      containers:
      - image: grissomsh/kubenginx:3.0.0
```

### 验证推出状态
- **观察：** 推出以滚动更新模式进行，因此没有停机时间。
```
# 验证推出状态
kubectl rollout status deployment/my-first-deployment
```
### 验证 ReplicaSet
- **观察：** 现在我们应该看到 3 个 ReplicaSet，因为我们已将应用程序更新到第 3 个版本 3.0.0
```
# 验证 ReplicaSet 和 Pod
kubectl get rs
kubectl get po
```
### 验证推出历史
```
# 检查 Deployment 的推出历史
kubectl rollout history deployment/<Deployment-Name>
kubectl rollout history deployment/my-first-deployment   
```

### 使用公网 IP 访问应用程序
- 每当我们在浏览器中访问应用程序时，应该看到 `Application Version:V3`
```
# 获取 NodePort
kubectl get svc
观察：记下以 3 开头的端口（例如：80:3xxxx/TCP）。捕获端口 3xxxx 并在下面的应用程序 URL 中使用它。

# 获取工作节点的公网 IP
kubectl get nodes -o wide
观察：如果您的 Kubernetes 集群在 AWS EKS 上设置，请记下 "EXTERNAL-IP"。

# 应用程序 URL
http://<worker-node-public-ip>:<Node-Port>
```
