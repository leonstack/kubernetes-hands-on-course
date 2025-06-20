# 1. 🚀 Kubernetes 学习 - Nginx 示例镜像

这是一个用于 Kubernetes 学习的示例应用程序，包含四个不同版本的 nginx 容器镜像。

## 1. 📁 项目结构

```text
01-kubenginx/
├── V1-Release/          # 版本 1 - 金色主题
│   ├── Dockerfile       # 优化的 Docker 配置
│   └── index.html       # 现代化的 HTML 页面
├── V2-Release/          # 版本 2 - 蓝色主题
│   ├── Dockerfile
│   └── index.html
├── V3-Release/          # 版本 3 - 紫色主题
│   ├── Dockerfile
│   └── index.html
├── V4-Release/          # 版本 4 - 粉色主题
│   ├── Dockerfile
│   └── index.html
├── build.sh             # 自动化构建脚本
└── README.md            # 项目说明文档
```

## 2. 🎨 版本特色

每个版本都有独特的视觉设计：

- **V1**: 🚀 金色渐变主题，温暖活力
- **V2**: 🌊 蓝色渐变主题，清新专业
- **V3**: 🔮 紫色渐变主题，神秘优雅
- **V4**: 🌸 粉色渐变主题，温柔浪漫

## 3. 🚀 快速开始

### 3.1 使用构建脚本（推荐）

```bash
# 构建所有版本
./build.sh -a

# 构建特定版本
./build.sh V1

# 构建并推送到 Registry
./build.sh -a -p

# 查看帮助
./build.sh -h

# 列出本地镜像
./build.sh -l

# 清理本地镜像
./build.sh -c
```

### 3.2 手动构建

```bash
# 构建 V1 版本
cd V1-Release
docker build -t grissomsh/kubenginx:v1 .

# 构建 V2 版本
cd ../V2-Release
docker build -t grissomsh/kubenginx:v2 .

# 以此类推...
```

### 3.3 运行容器

```bash
# 运行 V1 版本
docker run -d -p 8080:80 grissomsh/kubenginx:v1

# 访问应用
open http://localhost:8080
```

## 4. 🎯 Kubernetes 部署

### 4.1 创建 Deployment

```bash
# 部署 V1 版本
kubectl create deployment kubenginx --image=grissomsh/kubenginx:v1

# 扩展到 3 个副本
kubectl scale deployment kubenginx --replicas=3

# 暴露服务
kubectl expose deployment kubenginx --port=80 --type=NodePort
```

### 4.2 滚动更新

```bash
# 更新到 V2 版本
kubectl set image deployment/kubenginx nginx=grissomsh/kubenginx:v2

# 查看更新状态
kubectl rollout status deployment/kubenginx

# 回滚到上一版本
kubectl rollout undo deployment/kubenginx
```

## 5. 🔍 镜像信息

| 版本 | 标签 | 大小 | 主题色 | 特色 |
|------|------|------|--------|------|
| V1 | `grissomsh/kubenginx:v1` | ~15MB | 金色 | 🚀 活力启航 |
| V2 | `grissomsh/kubenginx:v2` | ~15MB | 蓝色 | 🌊 专业稳重 |
| V3 | `grissomsh/kubenginx:v3` | ~15MB | 紫色 | 🔮 神秘优雅 |
| V4 | `grissomsh/kubenginx:v4` | ~15MB | 粉色 | 🌸 温柔浪漫 |

## 6. 🛡️ 安全特性

- **非特权用户**: 容器以 nginx 用户身份运行，而非 root
- **最小权限**: 只暴露必要的端口和文件
- **Alpine 基础**: 使用安全的 Alpine Linux 减少攻击面
- **健康检查**: 内置健康检查确保服务可用性
