# Kubernetes Deployment 暂停和恢复操作

## 项目概述

本教程将指导您学习 Kubernetes Deployment 的暂停和恢复功能。当需要对 Deployment 进行多项更改时，可以先暂停部署，完成所有更改后再恢复，这样可以避免每次更改都触发新的滚动更新，提高操作效率并减少资源消耗。

## 学习目标

通过本教程，您将掌握：

✅ **暂停和恢复概念**

- 理解暂停和恢复 Deployment 的应用场景
- 掌握暂停状态下的配置更改方法
- 学习批量更改的最佳实践

✅ **操作技能**

- 使用 `kubectl rollout pause` 暂停部署
- 在暂停状态下进行多项配置更改
- 使用 `kubectl rollout resume` 恢复部署
- 监控和验证暂停恢复过程

✅ **实际应用**

- 应用版本从 V3 升级到 V4
- 同时设置容器资源限制
- 验证批量更改的效果

## 前置条件

在开始本教程之前，请确保：

1. **Kubernetes 集群**：已配置并运行的 Kubernetes 集群
2. **kubectl 工具**：已安装并配置 kubectl 命令行工具
3. **现有 Deployment**：已存在名为 `my-first-deployment` 的 Deployment
4. **应用版本**：当前应用版本为 V3（grissomsh/kubenginx:3.0.0）
5. **基础知识**：了解 Deployment、ReplicaSet 和 Pod 的基本概念

## 应用场景

**为什么需要暂停和恢复 Deployment？**

在生产环境中，经常需要对应用进行多项配置更改：

- 🔄 更新应用镜像版本
- 📊 调整资源限制和请求
- 🔧 修改环境变量
- 📝 更新标签和注解

如果逐一进行这些更改，每次都会触发新的滚动更新，导致：

- ⚠️ 多次不必要的 Pod 重启
- 📈 资源消耗增加
- ⏱️ 部署时间延长
- 🔍 版本历史混乱

通过暂停和恢复功能，可以：

- ✅ 批量完成所有更改
- ✅ 只触发一次滚动更新
- ✅ 提高部署效率
- ✅ 保持版本历史清晰  

## 1. 暂停和恢复 Deployment 操作

### 1.1 检查当前 Deployment 和应用状态

在开始暂停和恢复操作之前，我们需要了解当前的部署状态，这将帮助我们验证操作的效果。

#### 1.1.1 查看部署历史

```bash
# 检查 Deployment 的推出历史
kubectl rollout history deployment/my-first-deployment
```

**示例输出：**

```text
deployment.apps/my-first-deployment 
REVISION  CHANGE-CAUSE
1         kubectl create --filename=deployment.yaml --record=true
2         kubectl set image deployment/my-first-deployment kubenginx=grissomsh/kubenginx:2.0.0 --record=true
3         kubectl set image deployment/my-first-deployment kubenginx=grissomsh/kubenginx:3.0.0 --record=true
```

**观察要点：**

- 📝 记录当前的最新版本号（例如：版本 3）
- 📋 注意版本变更的历史记录
- 🔍 确认当前使用的镜像版本

#### 1.1.2 查看 ReplicaSet 状态

```bash
# 获取 ReplicaSet 列表
kubectl get rs
```

**示例输出：**

```text
NAME                               DESIRED   CURRENT   READY   AGE
my-first-deployment-7d9c6c8b4f     3         3         3       10m
my-first-deployment-6b8d4c7a5e     0         0         0       20m
my-first-deployment-5a7b3c6d9f     0         0         0       30m
```

**观察要点：**

- 📊 记录当前活跃的 ReplicaSet 数量
- 🔢 注意 DESIRED、CURRENT、READY 的数值
- 📅 观察各个 ReplicaSet 的创建时间

#### 1.1.3 查看 Pod 状态

```bash
# 查看 Pod 详细状态
kubectl get pods -l app=my-first-deployment -o wide
```

**示例输出：**

```text
NAME                                   READY   STATUS    RESTARTS   AGE   IP           NODE
my-first-deployment-7d9c6c8b4f-abc12   1/1     Running   0          10m   10.244.1.5   worker-1
my-first-deployment-7d9c6c8b4f-def34   1/1     Running   0          10m   10.244.2.3   worker-2
my-first-deployment-7d9c6c8b4f-ghi56   1/1     Running   0          10m   10.244.1.6   worker-1
```

#### 1.1.4 访问应用程序

```bash
# 获取 Service 信息
kubectl get service my-first-deployment-service

# 获取节点信息
kubectl get nodes -o wide
```

**访问应用：**

```bash
# 通过 NodePort 访问（替换为实际的节点IP和端口）
http://<worker-node-ip>:<Node-Port>

# 或使用端口转发进行本地测试
kubectl port-forward service/my-first-deployment-service 8080:80
# 然后访问：http://localhost:8080
```

**观察要点：**

- 🌐 记录当前应用程序的版本（应该显示 V3）
- ✅ 确认应用程序正常响应
- 📝 记录访问地址和端口信息

### 1.2 暂停 Deployment 并进行多项更改

现在我们将演示暂停 Deployment 的核心功能：在暂停状态下进行多项配置更改，而不触发滚动更新。

#### 1.2.1 暂停 Deployment

```bash
# 暂停 Deployment 的滚动更新
kubectl rollout pause deployment/my-first-deployment
```

**预期输出：**

```text
deployment.apps/my-first-deployment paused
```

**重要说明：**

- 🛑 暂停后，任何对 Deployment 的更改都不会立即触发滚动更新
- 📝 所有更改会被记录，但等待恢复时才会生效
- ✅ 现有的 Pod 继续正常运行，不受影响

#### 1.2.2 第一项更改：更新应用版本

```bash
# 更新应用镜像版本从 V3 到 V4
kubectl set image deployment/my-first-deployment kubenginx=grissomsh/kubenginx:4.0.0 --record=true
```

**预期输出：**

```text
deployment.apps/my-first-deployment image updated
```

#### 1.2.3 验证暂停状态下的行为

```bash
# 检查推出历史（应该没有新版本）
kubectl rollout history deployment/my-first-deployment
```

**观察要点：**

- 📊 版本数量应该与之前记录的相同
- 🔢 最新版本号应该没有变化
- ⏸️ 确认暂停状态阻止了新的滚动更新

```bash
# 检查 ReplicaSet 状态（应该没有新的 ReplicaSet）
kubectl get rs
```

**观察要点：**

- 📈 ReplicaSet 数量应该与之前相同
- 🚫 没有创建新的 ReplicaSet
- ✅ 当前活跃的 ReplicaSet 保持不变

```bash
# 检查 Deployment 状态
kubectl get deployment my-first-deployment -o wide
```

**示例输出：**

```text
NAME                  READY   UP-TO-DATE   AVAILABLE   AGE   CONTAINERS   IMAGES                      SELECTOR
my-first-deployment   3/3     3            3           45m   kubenginx    grissomsh/kubenginx:3.0.0   app=my-first-deployment
```

**注意：** IMAGES 列仍显示旧版本（3.0.0），因为更改尚未应用。

#### 1.2.4 第二项更改：设置资源限制

```bash
# 为容器设置 CPU 和内存限制
kubectl set resources deployment/my-first-deployment -c=kubenginx --limits=cpu=20m,memory=30Mi
```

**预期输出：**

```text
deployment.apps/my-first-deployment resource requirements updated
```

#### 1.2.5 验证多项更改的累积效果

```bash
# 查看 Deployment 的详细配置
kubectl describe deployment my-first-deployment
```

**关键观察点：**

- 🔍 在 Pod Template 部分应该看到新的镜像版本（4.0.0）
- 📊 应该看到新设置的资源限制
- ⏸️ 但这些更改尚未应用到实际的 Pod

```bash
# 确认 Pod 仍在使用旧配置
kubectl get pods -l app=my-first-deployment -o jsonpath='{.items[0].spec.containers[0].image}'
echo
kubectl get pods -l app=my-first-deployment -o jsonpath='{.items[0].spec.containers[0].resources}'
echo
```

**预期结果：**

- 镜像仍为：`grissomsh/kubenginx:3.0.0`
- 资源限制：可能为空或显示旧的设置

### 1.3 恢复 Deployment

现在我们将恢复 Deployment，这将触发一次滚动更新，应用之前在暂停状态下进行的所有更改。

#### 1.3.1 执行恢复操作

```bash
# 恢复 Deployment 的滚动更新
kubectl rollout resume deployment/my-first-deployment
```

**预期输出：**

```text
deployment.apps/my-first-deployment resumed
```

**重要说明：**

- 🚀 恢复后会立即开始滚动更新
- 📦 所有在暂停期间的更改会一次性应用
- 🔄 只会创建一个新的版本，而不是多个版本

#### 1.3.2 监控滚动更新过程

```bash
# 实时监控滚动更新状态
kubectl rollout status deployment/my-first-deployment
```

**预期输出：**

```text
Waiting for deployment "my-first-deployment" rollout to finish: 1 out of 3 new replicas have been updated...
Waiting for deployment "my-first-deployment" rollout to finish: 1 out of 3 new replicas have been updated...
Waiting for deployment "my-first-deployment" rollout to finish: 2 out of 3 new replicas have been updated...
Waiting for deployment "my-first-deployment" rollout to finish: 2 out of 3 new replicas have been updated...
Waiting for deployment "my-first-deployment" rollout to finish: 1 old replicas are pending termination...
Waiting for deployment "my-first-deployment" rollout to finish: 1 old replicas are pending termination...
deployment "my-first-deployment" successfully rolled out
```

#### 1.3.3 验证版本历史

```bash
# 检查推出历史（应该看到新版本）
kubectl rollout history deployment/my-first-deployment
```

**示例输出：**

```text
deployment.apps/my-first-deployment 
REVISION  CHANGE-CAUSE
1         kubectl create --filename=deployment.yaml --record=true
2         kubectl set image deployment/my-first-deployment kubenginx=grissomsh/kubenginx:2.0.0 --record=true
3         kubectl set image deployment/my-first-deployment kubenginx=grissomsh/kubenginx:3.0.0 --record=true
4         kubectl set image deployment/my-first-deployment kubenginx=grissomsh/kubenginx:4.0.0 --record=true
```

**观察要点：**

- ✅ 应该看到新创建的版本（版本 4）
- 📝 CHANGE-CAUSE 显示最后一次记录的更改
- 🔢 版本号比之前增加了 1

#### 1.3.4 验证 ReplicaSet 状态

```bash
# 检查 ReplicaSet 列表
kubectl get rs
```

**示例输出：**

```text
NAME                               DESIRED   CURRENT   READY   AGE
my-first-deployment-8e7f5d6c9b     3         3         3       2m
my-first-deployment-7d9c6c8b4f     0         0         0       15m
my-first-deployment-6b8d4c7a5e     0         0         0       25m
my-first-deployment-5a7b3c6d9f     0         0         0       35m
```

**观察要点：**

- 🆕 应该看到一个新的 ReplicaSet（最新的那个）
- 📊 新 ReplicaSet 的 DESIRED=3, CURRENT=3, READY=3
- 📉 旧 ReplicaSet 的副本数都变为 0
- ⏰ 新 ReplicaSet 的 AGE 应该很短

#### 1.3.5 验证 Pod 状态

```bash
# 检查 Pod 状态和详细信息
kubectl get pods -l app=my-first-deployment -o wide
```

**示例输出：**

```text
NAME                                   READY   STATUS    RESTARTS   AGE   IP           NODE
my-first-deployment-8e7f5d6c9b-xyz12   1/1     Running   0          2m    10.244.1.7   worker-1
my-first-deployment-8e7f5d6c9b-abc34   1/1     Running   0          2m    10.244.2.4   worker-2
my-first-deployment-8e7f5d6c9b-def56   1/1     Running   0          2m    10.244.1.8   worker-1
```

**观察要点：**

- 🆕 所有 Pod 都是新创建的（AGE 很短）
- ✅ 所有 Pod 状态都是 Running
- 🔄 Pod 名称包含新的 ReplicaSet 哈希值

### 1.4 验证应用程序更新

恢复部署后，我们需要验证所有更改是否正确应用，包括镜像版本和资源限制。

#### 1.4.1 验证镜像版本更新

```bash
# 检查 Pod 中的镜像版本
kubectl get pods -l app=my-first-deployment -o jsonpath='{.items[0].spec.containers[0].image}'
echo
```

**预期输出：**

```text
grissomsh/kubenginx:4.0.0
```

#### 1.4.2 验证资源限制设置

```bash
# 检查容器的资源限制
kubectl get pods -l app=my-first-deployment -o jsonpath='{.items[0].spec.containers[0].resources}'
echo
```

**预期输出：**

```text
{"limits":{"cpu":"20m","memory":"30Mi"}}
```

#### 1.4.3 访问应用程序

```bash
# 获取 Service 访问信息
kubectl get service my-first-deployment-service

# 通过 NodePort 访问应用程序
# 替换为实际的节点IP和端口
echo "访问地址：http://<node-ip>:<node-port>"

# 或使用端口转发进行本地测试
kubectl port-forward service/my-first-deployment-service 8080:80 &
echo "本地访问：http://localhost:8080"

# 使用 curl 测试（如果可用）
if command -v curl &> /dev/null; then
    echo "测试应用程序响应："
    curl -s http://localhost:8080 | grep -i version || echo "请手动访问验证版本"
fi
```

**观察要点：**

- 🌐 应用程序应该显示 **V4 版本**
- ✅ 确认应用程序正常响应
- 🔄 验证所有更改都已生效

#### 1.4.4 完整状态验证

```bash
# 显示完整的部署状态
echo "=== Deployment 状态 ==="
kubectl get deployment my-first-deployment -o wide

echo "\n=== Pod 详细信息 ==="
kubectl describe pods -l app=my-first-deployment | grep -E "Image:|Limits:"

echo "\n=== 版本历史 ==="
kubectl rollout history deployment/my-first-deployment
```

## 2. 清理资源

### 2.1 完整清理

如果需要完全清理演示环境：

```bash
# 删除 Deployment（会自动删除 ReplicaSet 和 Pod）
kubectl delete deployment my-first-deployment

# 删除 Service
kubectl delete service my-first-deployment-service

# 验证清理结果
kubectl get all -l app=my-first-deployment
```

**预期输出：**

```text
No resources found in default namespace.
```

### 2.2 保留资源清理

如果只需要重置到初始状态：

```bash
# 回滚到第一个版本
kubectl rollout undo deployment/my-first-deployment --to-revision=1

# 等待回滚完成
kubectl rollout status deployment/my-first-deployment

# 验证回滚结果
 kubectl get deployment my-first-deployment -o wide
 ```

## 3. 最佳实践和高级用法

### 3.1 暂停和恢复的最佳实践

#### 3.1.1 何时使用暂停和恢复

**适用场景：**

- 🔧 需要同时进行多项配置更改
- 📊 批量更新镜像版本和资源配置
- 🔄 避免频繁的滚动更新
- 🎯 在维护窗口期间进行计划性更新

**不适用场景：**

- 🚨 紧急安全补丁（需要立即应用）
- 🔥 生产环境的热修复
- 📈 单一配置更改

#### 3.1.2 操作前检查清单

```bash
# 1. 检查当前部署状态
kubectl get deployment my-first-deployment -o wide

# 2. 确认应用程序正常运行
kubectl get pods -l app=my-first-deployment

# 3. 备份当前配置（可选）
kubectl get deployment my-first-deployment -o yaml > deployment-backup.yaml

# 4. 检查集群资源
kubectl top nodes
kubectl top pods -l app=my-first-deployment
```

### 3.2 监控和验证策略

#### 3.2.1 暂停状态监控

```bash
# 检查部署是否处于暂停状态
kubectl get deployment my-first-deployment -o jsonpath='{.spec.paused}'
echo

# 查看暂停状态的详细信息
kubectl describe deployment my-first-deployment | grep -A 5 -B 5 "Paused"
```

#### 3.2.2 更改验证脚本

```bash
#!/bin/bash
# 验证暂停期间的更改

echo "=== 检查暂停状态 ==="
PAUSED=$(kubectl get deployment my-first-deployment -o jsonpath='{.spec.paused}')
if [ "$PAUSED" = "true" ]; then
    echo "✅ Deployment 已暂停"
else
    echo "❌ Deployment 未暂停"
fi

echo "\n=== 检查待应用的更改 ==="
kubectl describe deployment my-first-deployment | grep -A 10 "Pod Template"

echo "\n=== 检查当前运行的 Pod ==="
kubectl get pods -l app=my-first-deployment -o custom-columns=NAME:.metadata.name,IMAGE:.spec.containers[0].image,STATUS:.status.phase
```

### 3.3 高级配置示例

#### 3.3.1 复杂的批量更改

```bash
# 暂停部署
kubectl rollout pause deployment/my-first-deployment

# 1. 更新镜像版本
kubectl set image deployment/my-first-deployment kubenginx=grissomsh/kubenginx:4.0.0 --record=true

# 2. 设置资源限制和请求
kubectl set resources deployment/my-first-deployment -c=kubenginx --limits=cpu=50m,memory=64Mi --requests=cpu=10m,memory=32Mi

# 3. 添加环境变量
kubectl set env deployment/my-first-deployment APP_ENV=production

# 4. 更新标签
kubectl label deployment my-first-deployment version=v4.0.0 --overwrite

# 5. 添加注解
kubectl annotate deployment my-first-deployment deployment.kubernetes.io/change-cause="Batch update to v4.0.0 with resource limits"

# 恢复部署
kubectl rollout resume deployment/my-first-deployment
```

## 4. 故障排除

### 4.1 常见问题和解决方案

#### 4.1.1 暂停状态下无法应用更改

**问题症状：**

```bash
kubectl set image deployment/my-first-deployment kubenginx=new-image:tag
# 更改命令成功，但 Pod 没有更新
```

**解决方案：**

```bash
# 检查部署是否处于暂停状态
kubectl get deployment my-first-deployment -o jsonpath='{.spec.paused}'

# 如果返回 true，需要恢复部署
kubectl rollout resume deployment/my-first-deployment
```

#### 4.1.2 恢复后滚动更新失败

**问题症状：**

```bash
kubectl rollout status deployment/my-first-deployment
# 输出：Waiting for deployment "my-first-deployment" rollout to finish...
```

**排查步骤：**

```bash
# 1. 检查 Deployment 事件
kubectl describe deployment my-first-deployment

# 2. 检查 Pod 状态
kubectl get pods -l app=my-first-deployment

# 3. 查看 Pod 日志
kubectl logs -l app=my-first-deployment --previous

# 4. 检查资源限制是否合理
kubectl describe nodes
```

#### 4.1.3 忘记恢复部署

**问题症状：**

- 更改已应用但 Pod 没有更新
- 部署一直处于暂停状态

**解决方案：**

```bash
# 检查所有暂停的部署
kubectl get deployments -o jsonpath='{range .items[?(@.spec.paused==true)]}{.metadata.name}{"\n"}{end}'

# 恢复特定部署
kubectl rollout resume deployment/my-first-deployment
```

### 4.2 调试命令集合

```bash
# 完整状态检查脚本
echo "=== Deployment 状态 ==="
kubectl get deployment my-first-deployment -o wide

echo "\n=== 暂停状态 ==="
kubectl get deployment my-first-deployment -o jsonpath='{.spec.paused}'
echo

echo "\n=== ReplicaSet 状态 ==="
kubectl get rs -l app=my-first-deployment

echo "\n=== Pod 状态 ==="
kubectl get pods -l app=my-first-deployment -o wide

echo "\n=== 版本历史 ==="
kubectl rollout history deployment/my-first-deployment

echo "\n=== 最近事件 ==="
kubectl get events --field-selector involvedObject.name=my-first-deployment --sort-by=.metadata.creationTimestamp | tail -5
```

## 5. 总结

### 5.1 学习要点回顾

通过本教程，您已经掌握了：

✅ **暂停和恢复概念**

- 理解暂停和恢复的应用场景和优势
- 掌握批量更改的操作流程
- 学会监控和验证暂停恢复过程

✅ **实际操作技能**

- 使用 `kubectl rollout pause` 暂停部署
- 在暂停状态下进行多项配置更改
- 使用 `kubectl rollout resume` 恢复部署
- 验证更改的应用效果

✅ **最佳实践**

- 了解何时使用暂停和恢复功能
- 掌握操作前的检查清单
- 学会故障排除和问题诊断

### 5.2 关键优势总结

**效率提升：**

- 🚀 减少滚动更新次数
- ⏱️ 缩短部署时间
- 📊 降低资源消耗

**操作安全：**

- 🔒 批量验证更改
- 📝 保持版本历史清晰
- 🎯 精确控制更新时机

**生产环境友好：**

- 🕐 支持维护窗口操作
- 📈 减少服务中断
- 🔍 便于问题追踪

### 5.3 下一步学习

建议继续学习以下内容：

- **05-Services-with-kubectl**：学习服务发现和负载均衡
- **高级部署策略**：蓝绿部署、金丝雀发布
- **自动化部署**：CI/CD 流水线集成
- **监控和告警**：生产环境监控体系

### 5.4 生产环境建议

在生产环境中使用暂停和恢复功能时，请注意：

1. **制定操作计划**：明确更改内容和验证步骤
2. **设置监控告警**：确保及时发现问题
3. **准备回滚方案**：制定应急处理流程
4. **文档化操作**：记录每次暂停和恢复的原因
5. **团队协作**：确保团队成员了解暂停状态

## 6. 参考资料

- [Kubernetes Deployments 官方文档](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/)
- [kubectl rollout 命令参考](https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#rollout)
- [Deployment 滚动更新策略](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#rolling-update-deployment)
- [生产环境最佳实践](https://kubernetes.io/docs/concepts/configuration/overview/)
