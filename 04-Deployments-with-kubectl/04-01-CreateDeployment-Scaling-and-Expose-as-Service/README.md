# Kubernetes - Deployment

## 步骤 01：Deployment 介绍
- 什么是 Deployment？
- 使用 Deployment 可以做什么？
- 创建 Deployment
- 扩展 Deployment
- 将 Deployment 暴露为 Service

## 步骤 02：创建 Deployment
- 创建 Deployment 来推出 ReplicaSet
- 验证 Deployment、ReplicaSet 和 Pod
- **Docker 镜像位置：** https://hub.docker.com/repository/docker/grissomsh/kubenginx
```
# 创建 Deployment
kubectl create deployment <Deplyment-Name> --image=<Container-Image>
kubectl create deployment my-first-deployment --image=grissomsh/kubenginx:1.0.0 

# 验证 Deployment
kubectl get deployments
kubectl get deploy 

# 描述 Deployment
kubectl describe deployment <deployment-name>
kubectl describe deployment my-first-deployment

# 验证 ReplicaSet
kubectl get rs

# 验证 Pod
kubectl get po
```
## 步骤 03：扩展 Deployment
- 扩展 Deployment 以增加副本（Pod）数量
```
# 扩展 Deployment
kubectl scale --replicas=20 deployment/<Deployment-Name>
kubectl scale --replicas=20 deployment/my-first-deployment 

# 验证 Deployment
kubectl get deploy

# 验证 ReplicaSet
kubectl get rs

# 验证 Pod
kubectl get po

# 缩减 Deployment
kubectl scale --replicas=10 deployment/my-first-deployment 
```

## 步骤 04：将 Deployment 暴露为 Service
- 使用 Service（NodePort Service）暴露 **Deployment** 以从外部（互联网）访问应用程序
```
# 将 Deployment 暴露为 Service
kubectl expose deployment <Deployment-Name>  --type=NodePort --port=80 --target-port=80 --name=<Service-Name-To-Be-Created>
kubectl expose deployment my-first-deployment --type=NodePort --port=80 --target-port=80 --name=my-first-deployment-service

# 获取 Service 信息
kubectl get svc
观察：记下以 3 开头的端口（例如：80:3xxxx/TCP）。捕获端口 3xxxx 并在下面的应用程序 URL 中使用它。

# 获取工作节点的公网 IP
kubectl get nodes -o wide
观察：如果您的 Kubernetes 集群在 AWS EKS 上设置，请记下 "EXTERNAL-IP"。
```
- **使用公网 IP 访问应用程序**
```
http://<worker-node-public-ip>:<Node-Port>
```