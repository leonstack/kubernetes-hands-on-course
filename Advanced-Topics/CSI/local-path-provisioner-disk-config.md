# Local Path Provisioner 外部磁盘配置与存储挂载机制详解

## 1. 概述

本文档详细介绍如何配置 Local Path Provisioner 使用外部磁盘设备，实现 Kubernetes 集群中的本地持久化存储。通过将外部设备（如 `/dev/sdb`）格式化并挂载到指定目录，然后配置 Local Path Provisioner 使用该目录作为存储后端。

### 1.1 架构概览

```text
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   物理设备       │    │   主机文件系统     │    │   Kubelet       │    │   容器应用        │
│   /dev/sdb      │───▶│   /mnt/disk1     │───▶│   卷管理         │───▶│   /data         │
└─────────────────┘    └──────────────────┘    └─────────────────┘    └─────────────────┘
                              │                         ▲
                              ▼                         │
                       ┌──────────────────┐             │
                       │ Local Path       │─────────────┘
                       │ Provisioner      │ 动态卷创建
                       │ (CSI Driver)     │
                       └──────────────────┘
                              │
                              ▼
                       ┌──────────────────┐
                       │ Kubernetes       │
                       │ PVC/PV/SC        │
                       └──────────────────┘
```

### 1.2 配置流程

1. **设备准备**：格式化外部设备并挂载到主机目录
2. **Provisioner 配置**：修改 Local Path Provisioner 的 ConfigMap
3. **存储类配置**：创建或使用现有的 StorageClass
4. **应用部署**：创建 PVC 和 Pod 使用持久化存储

---

## 2. 外部存储设备准备与系统集成

### 2.1 设备格式化

首先需要对外部设备进行格式化。以 `/dev/sdb` 为例：

```bash
# 检查设备状态
lsblk /dev/sdb

# 创建文件系统（使用 ext4）
sudo mkfs.ext4 /dev/sdb

# 或者使用 xfs 文件系统（推荐用于大容量存储）
sudo mkfs.xfs /dev/sdb
```

**注意事项**：

- 格式化操作会清除设备上的所有数据，请确保数据已备份
- 对于生产环境，建议使用 xfs 文件系统以获得更好的性能
- 可以根据需要创建分区表，但对于专用存储设备，直接格式化整个设备通常更简单

**安全和性能考虑**：

- **设备权限**：确保只有授权用户可以访问存储设备
- **文件系统选择**：ext4 适合通用场景，xfs 适合大文件和高并发场景
- **块大小优化**：可通过 `-b` 参数指定块大小以优化性能
- **预留空间**：建议为文件系统预留 5-10% 空间以避免性能下降

### 2.2 创建挂载点

```bash
# 创建挂载目录
sudo mkdir -p /mnt/disk1

# 设置适当的权限
sudo chmod 755 /mnt/disk1
```

### 2.3 挂载设备

```bash
# 临时挂载（重启后失效）
sudo mount /dev/sdb /mnt/disk1

# 验证挂载状态
df -h /mnt/disk1
mount | grep /mnt/disk1
```

### 2.4 配置永久挂载

为确保系统重启后自动挂载，需要修改 `/etc/fstab` 文件：

```bash
# 获取设备 UUID（推荐使用 UUID 而非设备名）
sudo blkid /dev/sdb

# 编辑 fstab 文件
sudo vim /etc/fstab

# 添加以下行（替换 YOUR_UUID 为实际的 UUID）
UUID=YOUR_UUID /mnt/disk1 ext4 defaults,noatime 0 2

# 测试 fstab 配置
sudo mount -a

# 验证挂载
df -h /mnt/disk1
```

**fstab 参数说明**：

- `defaults`：使用默认挂载选项
- `noatime`：不更新访问时间，提高性能
- `0`：不进行 dump 备份
- `2`：文件系统检查优先级（根文件系统为 1，其他为 2）

### 2.5 设置存储目录权限

```bash
# 为 Local Path Provisioner 创建专用目录
sudo mkdir -p /mnt/disk1/local-path-provisioner

# 设置权限，确保 kubelet 可以访问
sudo chmod 777 /mnt/disk1/local-path-provisioner

# 验证权限设置
ls -la /mnt/disk1/
```

---

## 3. Local Path Provisioner 配置

### 3.1 查看当前配置

```bash
# 查看 Local Path Provisioner 的当前配置
kubectl get configmap local-path-config -n local-path-storage -o yaml
```

### 3.2 修改 ConfigMap 配置

创建新的配置文件，指定使用 `/mnt/disk1/local-path-provisioner` 作为存储路径：

```yaml
# local-path-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: local-path-config
  namespace: local-path-storage
data:
  config.json: |-
    {
      "nodePathMap": [
        {
          "node": "DEFAULT_PATH_FOR_NON_LISTED_NODES",
          "paths": ["/mnt/disk1/local-path-provisioner"]
        }
      ]
    }
  setup: |-
    #!/bin/sh
    set -eu
    mkdir -m 0777 -p "$VOL_DIR"
  teardown: |-
    #!/bin/sh
    set -eu
    rm -rf "$VOL_DIR"
  helperPod.yaml: |-
    apiVersion: v1
    kind: Pod
    metadata:
      name: helper-pod
    spec:
      priorityClassName: system-node-critical
      tolerations:
        - key: node.kubernetes.io/disk-pressure
          operator: Exists
          effect: NoSchedule
      containers:
      - name: helper-pod
        image: busybox
        imagePullPolicy: IfNotPresent
```

### 3.3 应用配置

```bash
# 应用新的配置
kubectl apply -f local-path-config.yaml

# 重启 Local Path Provisioner 以加载新配置
kubectl rollout restart deployment local-path-provisioner -n local-path-storage

# 验证 Provisioner 状态
kubectl get pods -n local-path-storage
kubectl logs -n local-path-storage -l app=local-path-provisioner
```

### 3.4 针对特定节点的配置

如果需要为特定节点配置不同的存储路径：

```yaml
# 特定节点配置示例
data:
  config.json: |-
    {
      "nodePathMap": [
        {
          "node": "worker-node-1",
          "paths": ["/mnt/disk1/local-path-provisioner"]
        },
        {
          "node": "worker-node-2", 
          "paths": ["/mnt/disk2/local-path-provisioner"]
        },
        {
          "node": "DEFAULT_PATH_FOR_NON_LISTED_NODES",
          "paths": ["/opt/local-path-provisioner"]
        }
      ]
    }
```

---

## 4. StorageClass 配置

### 4.1 查看现有 StorageClass

```bash
# 查看现有的 StorageClass
kubectl get storageclass

# 查看 Local Path 的 StorageClass 详情
kubectl get storageclass local-path -o yaml
```

### 4.2 创建自定义 StorageClass（可选）

如果需要创建专用的 StorageClass：

```yaml
# custom-local-storage.yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: local-disk-storage
  annotations:
    storageclass.kubernetes.io/is-default-class: "false"
provisioner: rancher.io/local-path
parameters:
  # 可以添加自定义参数
reclaimPolicy: Delete
allowVolumeExpansion: false
volumeBindingMode: WaitForFirstConsumer
```

```bash
# 应用自定义 StorageClass
kubectl apply -f custom-local-storage.yaml

# 验证创建
kubectl get storageclass local-disk-storage
```

---

## 5. 持久化存储应用配置与部署验证

### 5.1 创建 PersistentVolumeClaim

```yaml
# test-pvc.yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: test-local-pvc
  namespace: default
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: local-path  # 或使用自定义的 local-disk-storage
  resources:
    requests:
      storage: 10Gi
```

### 5.2 创建测试 Pod

```yaml
# test-pod.yaml
apiVersion: v1
kind: Pod
metadata:
  name: test-local-storage-pod
  namespace: default
spec:
  containers:
  - name: test-container
    image: nginx:latest
    ports:
    - containerPort: 80
    volumeMounts:
    - name: local-storage
      mountPath: /data
      mountPropagation: HostToContainer
    command:
    - /bin/sh
    - -c
    - |
      echo "Local storage test started at $(date)" > /data/test.log
      echo "Container hostname: $(hostname)" >> /data/test.log
      echo "Storage mount info:" >> /data/test.log
      df -h /data >> /data/test.log
      nginx -g "daemon off;"
  volumes:
  - name: local-storage
    persistentVolumeClaim:
      claimName: test-local-pvc
  restartPolicy: Always
```

### 5.3 部署和验证

```bash
# 创建 PVC
kubectl apply -f test-pvc.yaml

# 检查 PVC 状态
kubectl get pvc test-local-pvc

# 创建 Pod
kubectl apply -f test-pod.yaml

# 检查 Pod 状态
kubectl get pod test-local-storage-pod

# 查看 Pod 详情
kubectl describe pod test-local-storage-pod

# 验证存储挂载
kubectl exec test-local-storage-pod -- df -h /data
kubectl exec test-local-storage-pod -- cat /data/test.log
```

---

## 6. 存储挂载机制深度解析

### 6.1 容器内挂载点信息

在第5章创建的 `test-local-storage-pod` 容器中，可以查看挂载信息：

```bash
# 查看容器内的挂载点信息
kubectl exec test-local-storage-pod -- mount | grep /data

# 查看容器内存储使用情况
kubectl exec test-local-storage-pod -- df -h /data

# 查看挂载点的详细信息
kubectl exec test-local-storage-pod -- findmnt /data
```

**预期输出示例**：

```text
/dev/sdb on /data type ext4 (rw,relatime,noatime)
Filesystem      Size  Used Avail Use% Mounted on
/dev/sdb         50G  1.2G   46G   3% /data
```

### 6.2 主机端路径映射关系

#### 6.2.1 完整的路径映射链

Local Persistent Volume 的完整挂载链路包含四个关键层次：

```text
物理设备 → 主机挂载点 → Kubelet 管理目录 → 容器挂载点
   ↓           ↓              ↓              ↓
/dev/sdb → /mnt/disk1 → /var/lib/kubelet/pods/<pod-uid>/volumes/kubernetes.io~local-volume/test-local-pvc → /data
   │           │              │              │
   │           │              │              └─ 容器内访问路径
   │           │              └─ Kubelet 卷管理路径
   │           └─ 主机文件系统挂载点
   └─ 物理存储设备
```

**各层作用说明：**

| 层次 | 挂载类型 | 作用描述 | 管理组件 |
|------|----------|----------|----------|
| 第1层 | 文件系统挂载 | 将物理设备格式化并挂载到主机文件系统 | 系统管理员/自动化脚本 |
| 第2层 | Bind Mount | 将主机目录绑定到 Kubelet 管理路径 | Local Path Provisioner |
| 第3层 | Bind Mount + Namespace | 将 Kubelet 路径映射到容器 mount namespace | Kubelet |
| 第4层 | 容器内访问 | 应用程序通过容器内路径访问存储 | 容器运行时 |

#### 6.2.2 四层挂载参数分析

每一层的挂载都有具体的 Linux 挂载参数：

**第1层：物理设备 → 主机挂载点** (`/dev/sdb → /mnt/disk1`)

```bash
mount | grep /dev/sdb
# 输出: /dev/sdb on /mnt/disk1 type ext4 (rw,relatime,noatime)
```

- `rw`: 读写模式
- `relatime`: 相对时间更新，提高性能
- `noatime`: 不更新访问时间，进一步提高性能

**第2层：主机挂载点 → Local Path Provisioner 目录** (目录创建)

```bash
# Local Path Provisioner 在主机挂载点下创建 PVC 专用目录
ls -la /mnt/disk1/local-path-provisioner/
# 输出: drwxrwxrwx 3 root root 4096 ... pvc-<uuid>/
```

- 继承第1层的所有挂载参数
- Local Path Provisioner 负责目录的生命周期管理

**第3层：Local Path Provisioner 目录 → Kubelet 管理目录** (绑定挂载)

```bash
cat /proc/mounts | grep kubelet | grep test-local-pvc
# 输出: /mnt/disk1/local-path-provisioner/pvc-xxx on /var/lib/kubelet/pods/.../volumes/kubernetes.io~local-volume/test-local-pvc type none (rw,relatime,bind)
```

- `bind`: 绑定挂载，将 Local Path Provisioner 目录绑定到 Kubelet 管理路径
- `rw`: 读写权限
- `relatime`: 继承原挂载点的时间属性

**第4层：Kubelet 目录 → 容器挂载点** (绑定挂载 + 挂载传播)

Kubelet 将管理目录通过绑定挂载映射到容器的挂载命名空间：

```bash
# 主机端 Kubelet 管理目录
/var/lib/kubelet/pods/<pod-uid>/volumes/kubernetes.io~local-volume/test-local-pvc
# ↓ bind mount 到容器内
# 容器内挂载点: /data

# 在容器内查看挂载信息
kubectl exec test-local-storage-pod -- cat /proc/mounts | grep /data
# 输出: /dev/sdb on /data type ext4 (rw,relatime,noatime,rshared)
```

关键挂载参数说明：

- `bind`: Kubelet 通过 bind mount 将主机目录绑定到容器的 mount namespace
- `rw`: 读写权限传递到容器内
- `rslave`: 从属挂载传播（对应 Pod 配置中的 `mountPropagation: HostToContainer`）
  - 主机侧的挂载变化会传播到容器内
  - 支持动态挂载场景（如热插拔存储设备）
  - 实现了主机到容器的单向挂载事件传播
- 容器内的 `/data` 实际指向主机的 Kubelet 管理目录

**挂载传播机制详解**：

- **None (rprivate)**：完全隔离，主机和容器的挂载操作互不影响
- **HostToContainer (rslave)**：主机挂载变化传播到容器，容器挂载不影响主机
- **Bidirectional (rshared)**：双向传播，主机和容器的挂载变化相互影响

### 6.3 挂载关系验证

#### 6.3.1 验证数据一致性

```bash
# 在容器内创建测试文件
kubectl exec test-local-storage-pod -- sh -c "echo 'Container test' > /data/container-test.txt"

# 在主机上查找并验证文件
sudo find /mnt/disk1/local-path-provisioner -name "container-test.txt" -exec cat {} \;

# 在主机上创建文件
PVC_DIR=$(sudo find /mnt/disk1/local-path-provisioner -type d -name "pvc-*" | head -1)
sudo sh -c "echo 'Host test' > $PVC_DIR/host-test.txt"

# 在容器内验证文件
kubectl exec test-local-storage-pod -- cat /data/host-test.txt
```

#### 6.3.2 挂载传播验证

由于设置了 `mountPropagation: HostToContainer`，可以验证挂载传播行为：

```bash
# 查看挂载传播参数
kubectl exec test-local-storage-pod -- cat /proc/mounts | grep /data

# 验证传播模式对比
echo "传播模式说明："
echo "- None (rprivate): 完全隔离，无传播"
echo "- HostToContainer (rslave): 主机→容器单向传播"
echo "- Bidirectional (rshared): 双向传播"
```

### 6.4 数据流向分析

**写入操作流程**：

1. 应用程序在容器内向 `/data/file.txt` 写入数据
2. 容器运行时将写入操作传递给主机内核
3. 主机内核将数据写入 `/mnt/disk1/local-path-provisioner/pvc-xxx/file.txt`
4. 文件系统将数据最终写入物理设备 `/dev/sdb`

**读取操作流程**：

1. 应用程序从容器内的 `/data/file.txt` 读取数据
2. 容器运行时从主机文件系统读取数据
3. 主机内核从 `/mnt/disk1/local-path-provisioner/pvc-xxx/file.txt` 读取
4. 文件系统从物理设备 `/dev/sdb` 读取实际数据

### 6.5 关键挂载参数说明

#### 6.5.1 性能相关参数

| 挂载选项 | 性能影响 | 适用场景 | 说明 |
|---------|---------|---------|------|
| `noatime` | 🟢 高性能 | 读密集型应用 | 不记录访问时间，减少写入 |
| `relatime` | 🟡 平衡性能 | 通用场景 | 相对时间更新，兼顾性能和功能 |
| `bind` | 🟢 零拷贝 | 目录映射 | 最高效的目录共享方式 |
| `rshared` | 🟡 轻微开销 | 需要挂载传播 | 维护传播关系有少量开销 |

#### 6.5.2 监控挂载状态

```bash
# 监控物理设备状态
lsblk /dev/sdb

# 监控主机挂载点使用情况
df -h /mnt/disk1

# 监控容器内存储状态
kubectl exec test-local-storage-pod -- df -h /data

# 检查 PVC 目录
ls -la /mnt/disk1/local-path-provisioner/pvc-*/
```

---

## 参考资料

- [Local Path Provisioner 官方文档](https://github.com/rancher/local-path-provisioner)
- [Kubernetes 持久化卷文档](https://kubernetes.io/docs/concepts/storage/persistent-volumes/)
- [Linux 文件系统挂载指南](https://man7.org/linux/man-pages/man8/mount.8.html)
