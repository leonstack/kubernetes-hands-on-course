# Kubernetes - 使用 YAML 管理 POD
## 步骤 01：Kubernetes YAML 顶级对象
- 讨论 Kubernetes YAML 顶级对象
- **01-kube-base-definition.yml**
```yml
apiVersion:
kind:
metadata:
  
spec:
```
-  **Pod API 对象参考：**  https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#pod-v1-core

## 步骤 02：使用 YAML 创建简单的 Pod 定义
- 我们将创建一个非常基本的 Pod 定义
- **02-pod-definition.yml**
```yml
apiVersion: v1 # 字符串
kind: Pod  # 字符串
metadata: # 字典
  name: myapp-pod
  labels: # 字典 
    app: myapp         
spec:
  containers: # 列表
    - name: myapp
      image: grissomsh/kubenginx:1.0.0
      ports:
        - containerPort: 80
```
- **创建 Pod**
```
# 创建 Pod
kubectl create -f 02-pod-definition.yml
[或]
kubectl apply -f 02-pod-definition.yml

# 列出 Pod
kubectl get pods
```

## 步骤 03：创建 NodePort Service
- **03-pod-nodeport-service.yml**
```yml
apiVersion: v1
kind: Service
metadata:
  name: myapp-pod-nodeport-service  # Service 的名称
spec:
  type: NodePort
  selector:
  # 在匹配此标签选择器的 Pod 之间负载均衡流量
    app: myapp
  # 接受发送到端口 80 的流量    
  ports: 
    - name: http
      port: 80    # Service 端口
      targetPort: 80 # 容器端口
      nodePort: 31231 # NodePort
```
- **为 Pod 创建 NodePort Service**
```
# 创建 Service
kubectl apply -f 03-pod-nodeport-service.yml

# 列出 Service
kubectl get svc

# 获取公网 IP
kubectl get nodes -o wide

# 访问应用程序
http://<WorkerNode-Public-IP>:<NodePort>
http://<WorkerNode-Public-IP>:31231
```

## API 对象参考
-  **Pod**: https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#pod-v1-core
- **Service**: https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#service-v1-core

## 更新的 API 对象参考
-  **Pod**: https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/
-  **Service**: https://kubernetes.io/docs/reference/kubernetes-api/service-resources/service-v1/
- **Kubernetes API 参考：** https://kubernetes.io/docs/reference/kubernetes-api/

