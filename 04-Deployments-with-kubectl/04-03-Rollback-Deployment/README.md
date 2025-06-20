# Kubernetes Deployment 回滚操作

## 1. 项目概述

### 1.1 学习目标

- 掌握 Deployment 回滚的两种方法
- 理解版本历史管理和追踪机制
- 学会验证回滚结果和应用程序状态
- 掌握滚动重启操作
- 了解生产环境回滚最佳实践

### 1.2 回滚方法介绍

Kubernetes Deployment 支持两种回滚方式：

**1. 回滚到上一个版本**：

- 使用 `kubectl rollout undo` 命令
- 自动回滚到前一个稳定版本
- 适用于快速回滚场景

**2. 回滚到指定版本**：

- 使用 `kubectl rollout undo --to-revision=N` 命令
- 可以回滚到任意历史版本
- 适用于需要回滚到特定版本的场景

### 1.3 前置条件

- 已完成 04-01 和 04-02 教程
- 存在 `my-first-deployment` Deployment
- 具有多个版本历史记录

## 2. 回滚到上一个版本

### 2.1 查看 Deployment 的推出历史

在执行回滚操作之前，首先需要了解当前 Deployment 的版本历史：

```bash
# 查看 Deployment 推出历史
kubectl rollout history deployment/<Deployment-Name>
kubectl rollout history deployment/my-first-deployment
```

**输出示例：**

```text
deployment.apps/my-first-deployment 
REVISION  CHANGE-CAUSE
1         <none>
2         kubectl set image deployment/my-first-deployment kubenginx=grissomsh/kubenginx:2.0.0 --record=true
3         kubectl set image deployment/my-first-deployment kubenginx=grissomsh/kubenginx:3.0.0 --record=true
```

**重要概念：**

- **REVISION**：版本号，每次更新都会递增
- **CHANGE-CAUSE**：变更原因，记录了导致此版本的命令
- 使用 `--record` 参数可以记录变更原因

### 2.2 验证各个版本的详细信息

查看每个版本的具体配置变化，重点关注 "Annotations" 和 "Image" 标签：

```bash
# 查看指定版本的详细信息
kubectl rollout history deployment/my-first-deployment --revision=1
kubectl rollout history deployment/my-first-deployment --revision=2
kubectl rollout history deployment/my-first-deployment --revision=3
```

**观察要点：**

- 镜像版本变化
- 配置参数差异
- 注释信息
- 创建时间戳

### 2.3 执行回滚到上一个版本

执行回滚操作，系统会自动回滚到前一个稳定版本：

```bash
# 回滚 Deployment 到上一个版本
kubectl rollout undo deployment/my-first-deployment

# 查看回滚后的历史记录
kubectl rollout history deployment/my-first-deployment
```

**重要观察：**

- 如果当前是版本 3，回滚后会回到版本 2 的配置
- 但版本号会增加为版本 4（新的版本号）
- 这样可以保持版本历史的连续性

### 2.4 监控回滚过程

```bash
# 监控回滚状态
kubectl rollout status deployment/my-first-deployment

# 实时查看 Pod 变化
kubectl get pods -l app=my-first-deployment --watch
```

### 2.5 验证回滚结果

全面验证 Deployment、Pod 和 ReplicaSet 的状态：

```bash
# 查看 Deployment 状态
kubectl get deploy my-first-deployment -o wide

# 查看 ReplicaSet 状态
kubectl get rs -l app=my-first-deployment

# 查看 Pod 状态
kubectl get po -l app=my-first-deployment --show-labels

# 查看 Deployment 详细信息
kubectl describe deploy my-first-deployment
```

**验证要点：**

- Deployment 的镜像版本是否正确
- 新旧 ReplicaSet 的副本数分布
- Pod 的运行状态和镜像版本
- 事件日志中的回滚记录

### 2.6 测试应用程序访问

验证回滚后的应用程序是否正常工作，应该看到 `Application Version:V2`：

#### 2.6.1 获取访问信息

```bash
# 获取 Service 信息和 NodePort
kubectl get svc my-first-deployment-service

# 获取节点信息
kubectl get nodes -o wide
```

**观察要点：**

- 记录以 3 开头的端口号（例如：80:3xxxx/TCP）
- 记录节点的 EXTERNAL-IP（如果在 AWS EKS 上）

#### 2.6.2 构建访问 URL

```bash
# 应用程序访问地址
http://<worker-node-public-ip>:<Node-Port>

# 示例
http://192.168.1.100:32080
```

#### 2.6.3 使用端口转发进行本地测试

```bash
# 启动端口转发
kubectl port-forward svc/my-first-deployment-service 8080:80

# 在另一个终端测试
curl http://localhost:8080

# 或在浏览器中访问
http://localhost:8080
```

**预期结果：**

- 页面显示 `Application Version:V2`
- 确认回滚成功

## 3. 回滚到指定版本

### 3.1 查看当前版本历史

在回滚到特定版本之前，先确认可用的版本列表：

```bash
# 查看 Deployment 推出历史
kubectl rollout history deployment/my-first-deployment
```

**示例输出：**

```text
deployment.apps/my-first-deployment 
REVISION  CHANGE-CAUSE
1         <none>
2         kubectl set image deployment/my-first-deployment kubenginx=grissomsh/kubenginx:2.0.0 --record=true
3         kubectl set image deployment/my-first-deployment kubenginx=grissomsh/kubenginx:3.0.0 --record=true
4         kubectl rollout undo deployment/my-first-deployment
```

### 3.2 选择目标版本

根据需要选择要回滚的目标版本，查看其详细信息：

```bash
# 查看特定版本的详细信息
kubectl rollout history deployment/my-first-deployment --revision=3
```

### 3.3 执行指定版本回滚

回滚到指定的版本号：

```bash
# 回滚到指定版本（例如版本 3）
kubectl rollout undo deployment/my-first-deployment --to-revision=3

# 监控回滚过程
kubectl rollout status deployment/my-first-deployment
```

### 3.4 验证回滚后的历史记录

```bash
# 查看回滚后的历史记录
kubectl rollout history deployment/my-first-deployment
```

**重要观察：**

- 回滚到版本 3 后，配置会恢复到版本 3 的状态
- 但会生成新的版本号（例如版本 5）
- 原版本 3 在历史中仍然保留

### 3.5 测试指定版本回滚结果

验证回滚到指定版本后的应用程序，应该看到 `Application Version:V3`：

#### 3.5.1 验证镜像版本

```bash
# 验证当前镜像版本
kubectl get deployment my-first-deployment -o jsonpath='{.spec.template.spec.containers[0].image}'
echo

# 验证 Pod 镜像版本
kubectl get pods -l app=my-first-deployment -o jsonpath='{.items[*].spec.containers[0].image}'
echo
```

#### 3.5.2 应用程序功能测试

```bash
# 使用端口转发测试
kubectl port-forward svc/my-first-deployment-service 8080:80 &

# 测试应用程序响应
curl http://localhost:8080

# 停止端口转发
kill %1
```

**预期结果：**

- 镜像版本为 `grissomsh/kubenginx:3.0.0`
- 页面显示 `Application Version:V3`
- 确认指定版本回滚成功

## 4. 滚动重启应用程序

### 4.1 滚动重启介绍

滚动重启会以滚动方式终止现有 Pod 并创建新的 Pod，适用于以下场景：

- 应用配置更新（ConfigMap、Secret 变更）
- 强制重新拉取镜像
- 解决应用程序内存泄漏问题
- 应用环境变量更新

### 4.2 执行滚动重启

```bash
# 执行滚动重启
kubectl rollout restart deployment/my-first-deployment

# 监控重启过程
kubectl rollout status deployment/my-first-deployment

# 实时观察 Pod 变化
kubectl get pods -l app=my-first-deployment --watch
```

### 4.3 验证重启结果

```bash
# 查看 Pod 列表（注意 AGE 列）
kubectl get pods -l app=my-first-deployment

# 查看 Deployment 状态
kubectl get deployment my-first-deployment

# 查看重启历史
kubectl rollout history deployment/my-first-deployment
```

**观察要点：**

- 所有 Pod 的 AGE 都应该是最近创建的
- Deployment 配置没有变化，只是 Pod 被重新创建
- 版本历史中会增加一个新的记录

## 5. 回滚策略和最佳实践

### 5.1 生产环境回滚策略

#### 5.1.1 回滚前检查清单

```bash
# 1. 检查当前应用状态
kubectl get deployment my-first-deployment -o wide
kubectl get pods -l app=my-first-deployment

# 2. 检查应用日志
kubectl logs -l app=my-first-deployment --tail=50

# 3. 检查事件
kubectl get events --sort-by=.metadata.creationTimestamp

# 4. 检查资源使用情况
kubectl top pods -l app=my-first-deployment
```

#### 5.1.2 回滚决策矩阵

| 问题类型 | 回滚方式 | 优先级 | 备注 |
|---------|---------|--------|------|
| 应用启动失败 | 回滚到上一版本 | 高 | 快速恢复服务 |
| 性能问题 | 回滚到指定版本 | 中 | 需要确认稳定版本 |
| 功能缺陷 | 回滚到上一版本 | 中 | 评估影响范围 |
| 安全漏洞 | 立即回滚 | 极高 | 优先保障安全 |

### 5.2 版本管理最佳实践

#### 5.2.1 版本标记策略

```bash
# 使用有意义的标签
kubectl annotate deployment my-first-deployment deployment.kubernetes.io/revision-description="修复登录问题"

# 记录变更原因
kubectl set image deployment/my-first-deployment kubenginx=grissomsh/kubenginx:2.1.0 --record=true

# 添加版本标签
kubectl label deployment my-first-deployment version=v2.1.0
```

#### 5.2.2 历史版本保留策略

```bash
# 设置历史版本保留数量（默认10个）
kubectl patch deployment my-first-deployment -p '{"spec":{"revisionHistoryLimit":5}}'

# 查看当前设置
kubectl get deployment my-first-deployment -o jsonpath='{.spec.revisionHistoryLimit}'
```

### 5.3 监控和告警

#### 5.3.1 回滚监控指标

```bash
# 监控回滚过程中的关键指标
watch "kubectl get deployment my-first-deployment; echo '---'; kubectl get pods -l app=my-first-deployment"

# 检查回滚是否成功
kubectl rollout status deployment/my-first-deployment --timeout=300s
```

#### 5.3.2 健康检查

```bash
# 检查 Pod 健康状态
kubectl get pods -l app=my-first-deployment -o custom-columns=NAME:.metadata.name,STATUS:.status.phase,READY:.status.containerStatuses[0].ready,RESTARTS:.status.containerStatuses[0].restartCount

# 检查服务可用性
kubectl get endpoints my-first-deployment-service
```

## 6. 故障排除

### 6.1 常见回滚问题

#### 6.1.1 回滚失败

**问题症状：**

```bash
# 回滚命令执行后状态异常
kubectl rollout status deployment/my-first-deployment
# 输出：Waiting for deployment "my-first-deployment" rollout to finish...
```

**排查步骤：**

```bash
# 1. 检查 Deployment 事件
kubectl describe deployment my-first-deployment

# 2. 检查 ReplicaSet 状态
kubectl get rs -l app=my-first-deployment

# 3. 检查 Pod 状态和日志
kubectl get pods -l app=my-first-deployment
kubectl logs -l app=my-first-deployment --previous

# 4. 检查资源限制
kubectl describe nodes
```

#### 6.1.2 版本历史丢失

**问题症状：**

```bash
kubectl rollout history deployment/my-first-deployment
# 输出：No rollout history found
```

**解决方案：**

```bash
# 检查 revisionHistoryLimit 设置
kubectl get deployment my-first-deployment -o jsonpath='{.spec.revisionHistoryLimit}'

# 重新设置历史保留数量
kubectl patch deployment my-first-deployment -p '{"spec":{"revisionHistoryLimit":10}}'
```

### 6.2 调试命令集合

```bash
# 全面状态检查
echo "=== Deployment 状态 ==="
kubectl get deployment my-first-deployment -o wide

echo "=== ReplicaSet 状态 ==="
kubectl get rs -l app=my-first-deployment

echo "=== Pod 状态 ==="
kubectl get pods -l app=my-first-deployment -o wide

echo "=== 事件日志 ==="
kubectl get events --field-selector involvedObject.name=my-first-deployment --sort-by=.metadata.creationTimestamp

echo "=== 版本历史 ==="
kubectl rollout history deployment/my-first-deployment

echo "=== 当前镜像版本 ==="
kubectl get deployment my-first-deployment -o jsonpath='{.spec.template.spec.containers[0].image}'
echo
```

## 7. 清理资源

### 7.1 完整清理

如果需要完全清理演示环境：

```bash
# 删除 Deployment（会自动删除 ReplicaSet 和 Pod）
kubectl delete deployment my-first-deployment

# 删除 Service
kubectl delete service my-first-deployment-service

# 验证清理结果
kubectl get all -l app=my-first-deployment
```

### 7.2 保留资源清理

如果只需要重置到初始状态：

```bash
# 回滚到第一个版本
kubectl rollout undo deployment/my-first-deployment --to-revision=1

# 清理多余的历史版本（可选）
kubectl patch deployment my-first-deployment -p '{"spec":{"revisionHistoryLimit":3}}'
```

## 8. 总结

### 8.1 学习要点回顾

通过本教程，您已经掌握了：

✅ **回滚操作**

- 回滚到上一个版本的方法
- 回滚到指定版本的技巧
- 滚动重启的应用场景

✅ **版本管理**

- 版本历史的查看和分析
- 版本变更的追踪和记录
- 历史版本的保留策略

✅ **验证和测试**

- 回滚结果的全面验证
- 应用程序功能的测试方法
- 监控和健康检查

✅ **故障排除**

- 常见回滚问题的诊断
- 调试命令的使用
- 问题解决的最佳实践

### 8.2 下一步学习

建议继续学习以下内容：

- **04-04-Pause-and-Resume-Deployment**：学习暂停和恢复部署
- **高级部署策略**：蓝绿部署、金丝雀发布
- **GitOps 工作流**：自动化部署和回滚
- **监控和告警**：生产环境监控体系

### 8.3 生产环境建议

在生产环境中使用回滚功能时，请注意：

1. **制定回滚计划**：明确回滚触发条件和执行流程
2. **保留足够历史**：设置合适的 `revisionHistoryLimit`
3. **监控回滚过程**：确保回滚操作的成功执行
4. **测试回滚功能**：定期验证回滚机制的有效性
5. **文档化变更**：记录每次部署和回滚的原因

## 9. 参考资料

- [Kubernetes Deployments 官方文档](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/)
- [kubectl rollout 命令参考](https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#rollout)
- [Deployment 故障排除指南](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#troubleshooting)
- [生产环境最佳实践](https://kubernetes.io/docs/concepts/configuration/overview/)
