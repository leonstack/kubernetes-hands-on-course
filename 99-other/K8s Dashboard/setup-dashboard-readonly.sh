#!/bin/bash

set -e

echo "=== 创建 K8s Dashboard 只读用户 ==="

# 1. 创建 ServiceAccount
echo "1. 创建 ServiceAccount..."
kubectl create serviceaccount dashboard-readonly-user -n kube-system --dry-run=client -o yaml | kubectl apply -f -

# 2. 创建 ClusterRole
echo "2. 创建只读 ClusterRole..."
cat <<EOF | kubectl apply -f -
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
EOF

# 3. 创建 ClusterRoleBinding
echo "3. 创建 ClusterRoleBinding..."
cat <<EOF | kubectl apply -f -
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
EOF

# 4. 创建 Secret Token（可选，用于长期有效的 Token）
echo "4. 创建 Secret Token..."
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: dashboard-readonly-user-token
  namespace: kube-system
  annotations:
    kubernetes.io/service-account.name: dashboard-readonly-user
type: kubernetes.io/service-account-token
EOF

echo -e "\n=== 设置完成 ==="
echo -e "\n等待 Secret 创建完成..."
sleep 3

echo -e "\n=== 获取长期有效 Token ==="
# 等待 Secret 准备就绪
echo "等待 Secret 准备就绪..."
for i in {1..10}; do
    if kubectl -n kube-system get secret dashboard-readonly-user-token &>/dev/null; then
        TOKEN=$(kubectl -n kube-system get secret dashboard-readonly-user-token -o jsonpath='{.data.token}' 2>/dev/null | base64 -d 2>/dev/null)
        if [ ! -z "$TOKEN" ]; then
            echo -e "\n长期有效 Token："
            echo "$TOKEN"
            break
        fi
    fi
    echo "等待中... ($i/10)"
    sleep 2
done

if [ -z "$TOKEN" ]; then
    echo -e "\n无法自动获取长期有效 Token，请手动运行："
    echo "kubectl -n kube-system get secret dashboard-readonly-user-token -o jsonpath='{.data.token}' | base64 -d"
fi

echo -e "\n=== 使用说明 ==="
echo "1. 复制上面的 Token"
echo "2. 启动 kubectl proxy: kubectl proxy"
echo "3. 访问 Dashboard: http://localhost:8001/api/v1/namespaces/kubernetes-dashboard/services/https:kubernetes-dashboard:/proxy/"
echo "4. 选择 Token 登录方式并粘贴 Token"
echo "5. 享受只读访问权限！"