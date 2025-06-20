# 使用官方nginx镜像的alpine版本，更小更安全
FROM nginx:1.25-alpine

# 设置维护者信息
LABEL maintainer="Grissom <wang.tianqing.cn@outlook.com>"
LABEL version="3.0"
LABEL description="Grissom's Kubernetes Demo Application V3"

# 重用现有 nginx 用户
RUN id nginx

# 复制自定义HTML文件
COPY index.html /usr/share/nginx/html/

# 复制自定义nginx配置（可选）
# COPY nginx.conf /etc/nginx/nginx.conf

# 设置正确的文件权限
RUN chown -R nginx:nginx /usr/share/nginx/html && \
    chown -R nginx:nginx /var/cache/nginx && \
    chown -R nginx:nginx /var/log/nginx && \
    chown -R nginx:nginx /etc/nginx/conf.d

# 修改nginx配置以使用非root用户运行
RUN touch /var/run/nginx.pid && \
    chown -R nginx:nginx /var/run/nginx.pid

# 暴露端口
EXPOSE 80

# 切换到非root用户
USER nginx

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost/ || exit 1

# 启动nginx
CMD ["nginx", "-g", "daemon off;"]