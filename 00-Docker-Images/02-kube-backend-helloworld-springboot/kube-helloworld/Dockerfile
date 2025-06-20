# =============================================================================
# 多阶段构建 Dockerfile
# 阶段1: 构建阶段 - 使用 Maven 镜像编译源码
# 阶段2: 运行阶段 - 使用轻量级 JRE 镜像运行应用
# =============================================================================

# ===== 构建阶段 =====
# 使用官方 Maven 镜像，包含 JDK 8 和 Maven，避免额外安装
FROM maven:3.8.7-eclipse-temurin-8 AS builder

# 设置构建阶段的维护者信息
LABEL stage=builder

# 设置工作目录
WORKDIR /build

# 首先复制 pom.xml 以利用 Docker 缓存层
# 这样当源码改变但依赖不变时，可以重用依赖下载的缓存层
COPY pom.xml .

# 下载依赖（利用缓存层优化）
RUN mvn dependency:go-offline -B

# 复制源码
COPY src ./src

# 编译应用并跳过测试（生产环境建议启用测试）
RUN mvn clean package -DskipTests -B

# 验证构建产物
RUN ls -la /build/target/ && \
    test -f /build/target/hello-world-rest-api.jar

# ===== 运行阶段 =====
# 使用 JRE 8 镜像，与构建阶段的 JDK 版本保持一致
FROM eclipse-temurin:8u372-b07-jdk-jammy AS runtime

# 设置维护者信息和标签
LABEL maintainer="Grissom <wang.tianqing.cn@outlook.com>" \
      description="Spring Boot Hello World Application - Multi-stage Build" \
      version="1.0.0" \
      build-stage="multi-stage" \
      base-image="eclipse-temurin:8u372-b07-jdk-jammy"

# 安装运行时需要的工具（用于健康检查）
RUN apt update && \
    apt install -y wget

# 创建非root用户以提高安全性（Ubuntu/Debian 语法）
RUN groupadd -g 1001 appgroup && \
    useradd -u 1001 -g appgroup -m -s /bin/bash appuser

# 设置工作目录
WORKDIR /app

# 从构建阶段复制编译好的 JAR 文件
COPY --from=builder /build/target/hello-world-rest-api.jar app.jar

# 创建日志目录
RUN mkdir -p /app/logs && \
    chown -R appuser:appgroup /app

# 切换到非root用户
USER appuser

# 暴露端口
EXPOSE 8080

# 添加健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/actuator/health || exit 1

# 优化JVM参数并启动应用
ENTRYPOINT ["java", \
    "-XX:+UseContainerSupport", \
    "-XX:MaxRAMPercentage=75.0", \
    "-XX:+UseG1GC", \
    "-XX:+UseStringDeduplication", \
    "-XX:+PrintGCDetails", \
    "-XX:+PrintGCTimeStamps", \
    "-Xloggc:/app/logs/gc.log", \
    "-Djava.security.egd=file:/dev/./urandom", \
    "-Dspring.profiles.active=${SPRING_PROFILES_ACTIVE:-default}", \
    "-jar", "app.jar"]