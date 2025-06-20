# 1. Kubernetes - Update Deployments

## 1.1 项目概述

本教程将详细介绍如何更新 Kubernetes Deployment，包括两种主要的更新方法和完整的验证流程。通过本教程，您将学会如何安全地进行应用程序版本升级，并掌握滚动更新的核心概念。

### 1.1.1 学习目标

- 掌握 Deployment 更新的两种方法
- 理解滚动更新机制和零停机部署
- 学会监控和验证更新过程
- 了解版本历史管理和回滚准备

## 1.2 Deployment 更新介绍

### 1.2.1 更新方法概述

Kubernetes 提供了多种更新 Deployment 的方法：

1. **kubectl set image** - 快速镜像更新
   - 适用于简单的镜像版本升级
   - 命令行操作，快速便捷
   - 支持记录更新历史

2. **kubectl edit** - 交互式编辑
   - 适用于复杂的配置修改
   - 可以同时修改多个字段
   - 实时验证 YAML 语法

3. **kubectl apply** - 声明式更新（推荐生产环境）
   - 基于 YAML 文件的版本控制
   - 支持 GitOps 工作流
   - 便于审计和回滚

### 1.2.2 滚动更新策略

- **零停机时间**：逐步替换旧版本 Pod
- **可配置策略**：控制更新速度和并发数
- **自动回滚**：检测到问题时自动恢复
- **健康检查**：确保新版本正常运行

## 2. 方法一：使用 kubectl set image 更新 (V1 → V2)

### 2.1 准备工作

在开始更新之前，确保您已经有一个运行中的 Deployment。如果没有，请先参考前面的教程创建一个。

```bash
# 检查当前 Deployment 状态
kubectl get deployments
kubectl get pods -l app=my-first-deployment

# 检查当前镜像版本
kubectl get deployment my-first-deployment -o jsonpath='{.spec.template.spec.containers[0].image}'
```

### 2.2 获取容器信息

**重要提示：** 在使用 `kubectl set image` 命令之前，需要确认容器名称。

```bash
# 方法1：查看完整的 YAML 配置
kubectl get deployment my-first-deployment -o yaml

# 方法2：直接获取容器名称
kubectl get deployment my-first-deployment -o jsonpath='{.spec.template.spec.containers[0].name}'

# 方法3：查看 Deployment 描述信息
kubectl describe deployment my-first-deployment
```

### 2.3 执行镜像更新

```bash
# 基本语法
kubectl set image deployment/<Deployment-Name> <Container-Name>=<New-Image> --record=true

# 实际更新命令（将版本从 1.0.0 更新到 2.0.0）
kubectl set image deployment/my-first-deployment kubenginx=grissomsh/kubenginx:2.0.0 --record=true

# 验证命令是否成功执行
echo "更新命令执行完成，开始验证..."
```

**参数说明：**

- `--record=true`：记录此次更新到历史中，便于后续回滚
- `kubenginx`：容器名称（需要根据实际情况调整）
- `grissomsh/kubenginx:2.0.0`：新的镜像版本

### 2.4 监控更新过程

**重要概念：** Kubernetes 默认使用滚动更新策略，确保应用程序零停机时间。

```bash
# 实时监控更新状态
kubectl rollout status deployment/my-first-deployment

# 查看 Deployment 当前状态
kubectl get deployments
kubectl get deployment my-first-deployment -o wide

# 实时观察 Pod 变化（可选，按 Ctrl+C 停止）
kubectl get pods -l app=my-first-deployment -w
```

### 2.5 验证更新结果

```bash
# 检查新的镜像版本
kubectl get deployment my-first-deployment -o jsonpath='{.spec.template.spec.containers[0].image}'

# 验证 Deployment 详细信息
kubectl describe deployment my-first-deployment

# 查看更新事件
kubectl get events --sort-by=.metadata.creationTimestamp | grep my-first-deployment | tail -10
```

**观察要点：**

- 滚动更新过程中的事件序列
- 新旧 ReplicaSet 的替换过程
- Pod 的创建和终止时间
- 更新策略的执行效果

### 2.6 验证 ReplicaSet 变化

**核心概念：** 每次更新都会创建新的 ReplicaSet，旧的 ReplicaSet 会被保留用于回滚。

```bash
# 查看所有 ReplicaSet
kubectl get replicasets
kubectl get rs -l app=my-first-deployment

# 查看 ReplicaSet 详细信息
kubectl get rs -l app=my-first-deployment -o wide

# 查看 ReplicaSet 的所有者引用
kubectl get rs -l app=my-first-deployment -o yaml | grep -A 5 ownerReferences
```

**观察要点：**

- 新 ReplicaSet 的副本数应该等于期望的副本数
- 旧 ReplicaSet 的副本数应该为 0
- ReplicaSet 名称包含 Pod 模板哈希

### 2.7 验证 Pod 状态

```bash
# 查看 Pod 列表和标签
kubectl get pods -l app=my-first-deployment --show-labels

# 查看 Pod 详细信息
kubectl get pods -l app=my-first-deployment -o wide

# 检查 Pod 的镜像版本
kubectl get pods -l app=my-first-deployment -o jsonpath='{.items[*].spec.containers[0].image}'

# 查看 Pod 的所有者引用
kubectl get pods -l app=my-first-deployment -o yaml | grep -A 5 ownerReferences
```

**验证要点：**

- Pod 模板哈希标签匹配新的 ReplicaSet
- 所有 Pod 都运行新版本的镜像
- Pod 状态为 Running 且 Ready

### 2.8 查看更新历史

```bash
# 查看 Deployment 的推出历史
kubectl rollout history deployment/my-first-deployment

# 查看特定修订版本的详细信息
kubectl rollout history deployment/my-first-deployment --revision=1
kubectl rollout history deployment/my-first-deployment --revision=2

# 比较不同版本的差异
echo "=== 版本 1 详情 ==="
kubectl rollout history deployment/my-first-deployment --revision=1
echo "=== 版本 2 详情 ==="
kubectl rollout history deployment/my-first-deployment --revision=2
```

**历史记录说明：**

- `REVISION`：修订版本号，递增
- `CHANGE-CAUSE`：更新原因（使用 --record 时记录）
- 历史记录用于回滚操作

### 2.9 测试应用程序访问

**验证目标：** 确认新版本应用程序正常运行，显示 `Application Version:V2`。

```bash
# 检查 Service 状态
kubectl get services
kubectl get svc my-first-deployment-service -o wide

# 获取 NodePort 信息
NODE_PORT=$(kubectl get svc my-first-deployment-service -o jsonpath='{.spec.ports[0].nodePort}')
echo "NodePort: $NODE_PORT"

# 获取节点 IP 地址
NODE_IP=$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="ExternalIP")].address}')
if [ -z "$NODE_IP" ]; then
    NODE_IP=$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}')
fi
echo "Node IP: $NODE_IP"

# 构建访问 URL
echo "应用程序访问地址: http://$NODE_IP:$NODE_PORT"
```

### 2.10 连接测试

```bash
# 使用 curl 测试（如果可用）
if command -v curl &> /dev/null; then
    echo "测试应用程序响应..."
    curl -s http://$NODE_IP:$NODE_PORT | grep -i version || echo "版本信息未找到"
else
    echo "curl 命令不可用，请在浏览器中访问: http://$NODE_IP:$NODE_PORT"
fi

# 使用端口转发进行本地测试
echo "启动端口转发进行本地测试..."
kubectl port-forward svc/my-first-deployment-service 8080:80 &
PORT_FORWARD_PID=$!
sleep 3

# 测试本地连接
curl -s http://localhost:8080 | head -10 || echo "本地连接测试失败"

# 停止端口转发
kill $PORT_FORWARD_PID 2>/dev/null
```

**预期结果：**

- 应用程序正常响应
- 页面显示 `Application Version:V2`
- 所有 Pod 都能正常处理请求

## 3. 方法二：使用 kubectl edit 更新 (V2 → V3)

### 3.1 交互式编辑介绍

`kubectl edit` 命令提供了一种交互式的方式来修改 Kubernetes 资源。它会打开默认编辑器（通常是 vi/vim），允许您直接编辑资源的 YAML 配置。

**优势：**

- 可以同时修改多个字段
- 实时 YAML 语法验证
- 支持复杂的配置更改
- 适合学习和调试

### 3.2 准备编辑环境

```bash
# 设置首选编辑器（可选）
export EDITOR=nano  # 或者 vim, code 等

# 查看当前配置
kubectl get deployment my-first-deployment -o yaml | grep -A 5 -B 5 image

# 备份当前配置（推荐）
kubectl get deployment my-first-deployment -o yaml > deployment-backup-v2.yaml
echo "配置已备份到 deployment-backup-v2.yaml"
```

### 3.3 执行交互式编辑

```bash
# 启动交互式编辑
kubectl edit deployment/my-first-deployment --record=true
```

**编辑指南：**

在打开的编辑器中，找到以下部分：

```yaml
# 查找这个部分（大约在第 35-40 行）
spec:
  template:
    spec:
      containers:
      - image: grissomsh/kubenginx:2.0.0  # 将此行修改
        name: kubenginx
```

**修改步骤：**

1. 找到 `image: grissomsh/kubenginx:2.0.0` 行
2. 将 `2.0.0` 修改为 `3.0.0`
3. 保存并退出编辑器
   - vim: 按 `Esc`，输入 `:wq`，按 `Enter`
   - nano: 按 `Ctrl+X`，按 `Y`，按 `Enter`

```yaml
# 修改后的内容
spec:
  template:
    spec:
      containers:
      - image: grissomsh/kubenginx:3.0.0  # 已修改
        name: kubenginx
```

### 3.4 验证编辑结果

```bash
# 监控更新过程
kubectl rollout status deployment/my-first-deployment

# 验证镜像版本已更新
kubectl get deployment my-first-deployment -o jsonpath='{.spec.template.spec.containers[0].image}'

# 查看更新事件
kubectl get events --sort-by=.metadata.creationTimestamp | grep my-first-deployment | tail -5
```

### 3.5 验证 ReplicaSet 状态

**重要观察：** 现在应该有 3 个 ReplicaSet（V1、V2、V3），只有最新的处于活跃状态。

```bash
# 查看所有 ReplicaSet
kubectl get replicasets -l app=my-first-deployment

# 详细查看 ReplicaSet 状态
kubectl get rs -l app=my-first-deployment -o wide

# 验证 Pod 分布
kubectl get pods -l app=my-first-deployment --show-labels

# 检查 Pod 镜像版本
kubectl get pods -l app=my-first-deployment -o jsonpath='{.items[*].spec.containers[0].image}'
```

### 3.6 更新历史验证

```bash
# 查看完整的更新历史
kubectl rollout history deployment/my-first-deployment

# 查看各个版本的详细信息
echo "=== 版本 1 (初始版本) ==="
kubectl rollout history deployment/my-first-deployment --revision=1

echo "=== 版本 2 (set image 更新) ==="
kubectl rollout history deployment/my-first-deployment --revision=2

echo "=== 版本 3 (edit 更新) ==="
kubectl rollout history deployment/my-first-deployment --revision=3
```

### 3.7 应用程序功能测试

**验证目标：** 确认应用程序显示 `Application Version:V3`。

```bash
# 获取访问信息
NODE_PORT=$(kubectl get svc my-first-deployment-service -o jsonpath='{.spec.ports[0].nodePort}')
NODE_IP=$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}')

echo "应用程序访问地址: http://$NODE_IP:$NODE_PORT"

# 功能测试
echo "测试应用程序版本..."
kubectl port-forward svc/my-first-deployment-service 8080:80 &
PORT_FORWARD_PID=$!
sleep 3

# 验证版本信息
if command -v curl &> /dev/null; then
    VERSION_INFO=$(curl -s http://localhost:8080 | grep -i "version" || echo "未找到版本信息")
    echo "版本信息: $VERSION_INFO"
else
    echo "请在浏览器中访问 http://localhost:8080 验证版本为 V3"
fi

# 清理端口转发
 kill $PORT_FORWARD_PID 2>/dev/null
 ```

## 4. 更新策略配置

### 4.1 滚动更新参数

```bash
# 查看当前更新策略
kubectl get deployment my-first-deployment -o yaml | grep -A 10 strategy

# 配置更新策略（可选）
kubectl patch deployment my-first-deployment -p '{
  "spec": {
    "strategy": {
      "type": "RollingUpdate",
      "rollingUpdate": {
        "maxUnavailable": "25%",
        "maxSurge": "25%"
      }
    }
  }
}'
```

**参数说明：**

- `maxUnavailable`: 更新过程中不可用的 Pod 最大数量
- `maxSurge`: 更新过程中可以创建的额外 Pod 最大数量
- `type`: 更新类型（RollingUpdate 或 Recreate）

### 4.2 健康检查配置

```bash
# 查看当前健康检查配置
kubectl get deployment my-first-deployment -o yaml | grep -A 15 -B 5 probe

# 添加就绪性探针（示例）
kubectl patch deployment my-first-deployment -p '{
  "spec": {
    "template": {
      "spec": {
        "containers": [{
          "name": "kubenginx",
          "readinessProbe": {
            "httpGet": {
              "path": "/",
              "port": 80
            },
            "initialDelaySeconds": 5,
            "periodSeconds": 10
          }
        }]
      }
    }
  }
}'
```

## 5. 最佳实践

### 5.1 生产环境更新建议

1. **使用声明式配置**

   ```bash
   # 推荐：使用 YAML 文件管理
   kubectl apply -f deployment.yaml
   
   # 而不是命令式更新
   kubectl set image deployment/my-app container=image:tag
   ```

2. **版本控制和标签管理**

   ```bash
   # 使用语义化版本标签
   kubectl set image deployment/my-app container=myapp:v1.2.3
   
   # 添加有意义的标签
   kubectl label deployment my-app version=v1.2.3 release=stable
   ```

3. **渐进式部署**

   ```bash
   # 先更新少量副本进行测试
   kubectl scale deployment my-app --replicas=2
   kubectl set image deployment/my-app container=myapp:v1.2.3
   
   # 验证无误后扩展到全部副本
   kubectl scale deployment my-app --replicas=10
   ```

### 5.2 监控和验证

```bash
# 设置更新超时
kubectl rollout status deployment/my-first-deployment --timeout=300s

# 监控资源使用
kubectl top pods -l app=my-first-deployment

# 检查日志
kubectl logs -l app=my-first-deployment --tail=50

# 验证服务可用性
kubectl get endpoints my-first-deployment-service
```

## 6. 故障排除

### 6.1 常见问题

**问题 1：更新卡住不动**：

```bash
# 检查 Pod 状态
kubectl get pods -l app=my-first-deployment

# 查看 Pod 详细信息
kubectl describe pod <pod-name>

# 检查镜像拉取状态
kubectl get events --sort-by=.metadata.creationTimestamp | grep -i pull
```

**问题 2：新版本启动失败**：

```bash
# 查看 Pod 日志
kubectl logs <pod-name> --previous

# 检查容器状态
kubectl describe pod <pod-name>

# 验证镜像是否存在
docker pull grissomsh/kubenginx:3.0.0
```

**问题 3：服务不可用**：

```bash
# 检查 Service 端点
kubectl get endpoints

# 验证标签选择器
kubectl get pods --show-labels
kubectl get svc my-first-deployment-service -o yaml | grep selector

# 测试 Pod 直接访问
kubectl port-forward pod/<pod-name> 8080:80
```

### 6.2 回滚操作

```bash
# 快速回滚到上一个版本
kubectl rollout undo deployment/my-first-deployment

# 回滚到特定版本
kubectl rollout undo deployment/my-first-deployment --to-revision=2

# 验证回滚结果
kubectl rollout status deployment/my-first-deployment
kubectl get deployment my-first-deployment -o jsonpath='{.spec.template.spec.containers[0].image}'
```

### 6.3 调试命令集合

```bash
# 完整的状态检查脚本
echo "=== Deployment 状态 ==="
kubectl get deployment my-first-deployment -o wide

echo "=== ReplicaSet 状态 ==="
kubectl get rs -l app=my-first-deployment

echo "=== Pod 状态 ==="
kubectl get pods -l app=my-first-deployment -o wide

echo "=== Service 状态 ==="
kubectl get svc my-first-deployment-service

echo "=== 最近事件 ==="
kubectl get events --sort-by=.metadata.creationTimestamp | tail -10

echo "=== 更新历史 ==="
kubectl rollout history deployment/my-first-deployment
```

## 7. 清理资源

### 7.1 完整清理

```bash
# 删除 Deployment（会自动删除 ReplicaSet 和 Pod）
kubectl delete deployment my-first-deployment

# 删除 Service
kubectl delete service my-first-deployment-service

# 验证清理结果
kubectl get all -l app=my-first-deployment
```

### 7.2 保留资源清理

```bash
# 仅缩减到 0 副本（保留配置）
kubectl scale deployment my-first-deployment --replicas=0

# 验证 Pod 已停止
kubectl get pods -l app=my-first-deployment
```

## 8. 总结

### 8.1 学习要点回顾

通过本教程，您已经掌握了：

✅ **两种更新方法**

- `kubectl set image`：快速镜像更新
- `kubectl edit`：交互式配置修改

✅ **滚动更新机制**

- 零停机时间部署
- 渐进式 Pod 替换
- 自动健康检查

✅ **验证和监控**

- 更新状态监控
- ReplicaSet 管理
- 版本历史追踪

✅ **故障排除**

- 常见问题诊断
- 回滚操作
- 调试技巧

### 8.2 下一步学习

- **Deployment 回滚**：学习版本回退和历史管理
- **暂停和恢复**：掌握部署过程控制
- **蓝绿部署**：了解高级部署策略
- **金丝雀发布**：学习渐进式发布技术
- **Helm 包管理**：使用 Helm 管理复杂应用

### 8.3 生产环境建议

1. **使用 GitOps 工作流**
2. **实施自动化测试**
3. **配置监控和告警**
4. **建立回滚预案**
5. **定期备份配置**

## 9. 参考资料

- [Kubernetes Deployments 官方文档](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/)
- [滚动更新策略](https://kubernetes.io/docs/tutorials/kubernetes-basics/update/update-intro/)
- [kubectl 命令参考](https://kubernetes.io/docs/reference/kubectl/)
- [Deployment 故障排除](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#troubleshooting)
