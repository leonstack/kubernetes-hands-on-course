#!/bin/bash

set -e

echo "=== 清理 K8s Dashboard 只读用户资源 ==="

# 确认操作
read -p "确定要删除所有 dashboard-readonly 相关资源吗？(y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "操作已取消"
    exit 1
fi

echo "开始清理资源..."

# 1. 删除 ClusterRoleBinding
echo "1. 删除 ClusterRoleBinding..."
kubectl delete clusterrolebinding dashboard-readonly --ignore-not-found=true

# 2. 删除 ClusterRole
echo "2. 删除 ClusterRole..."
kubectl delete clusterrole dashboard-readonly --ignore-not-found=true

# 3. 删除 Secret
echo "3. 删除 Secret..."
kubectl delete secret dashboard-readonly-user-token -n kube-system --ignore-not-found=true

# 4. 删除 ServiceAccount
echo "4. 删除 ServiceAccount..."
kubectl delete serviceaccount dashboard-readonly-user -n kube-system --ignore-not-found=true

echo "\n=== 清理完成 ==="
echo "所有 dashboard-readonly 相关资源已删除。"

# 验证清理结果
echo "\n=== 验证清理结果 ==="
echo "检查剩余资源..."

echo "\nClusterRoleBinding:"
kubectl get clusterrolebinding | grep dashboard-readonly || echo "✅ 无相关 ClusterRoleBinding"

echo "\nClusterRole:"
kubectl get clusterrole | grep dashboard-readonly || echo "✅ 无相关 ClusterRole"

echo "\nSecret:"
kubectl get secret -n kube-system | grep dashboard-readonly || echo "✅ 无相关 Secret"

echo "\nServiceAccount:"
kubectl get serviceaccount -n kube-system | grep dashboard-readonly || echo "✅ 无相关 ServiceAccount"

echo "\n清理验证完成！"