# 使用官方nginx镜像的特定版本，提高安全性和稳定性
FROM nginx:1.25-alpine

# 设置维护者信息
LABEL maintainer="Grissom <wang.tianqing.cn@outlook.com>" \
      description="Frontend Nginx proxy for Kubernetes demo" \
      version="1.0"

# 创建非root用户提高安全性
RUN addgroup -g 1001 -S nginx-user && \
    adduser -S -D -H -u 1001 -h /var/cache/nginx -s /sbin/nologin -G nginx-user -g nginx-user nginx-user

# 复制nginx配置文件
COPY default.conf /etc/nginx/conf.d/default.conf

# 移除默认的nginx配置
RUN rm -f /etc/nginx/conf.d/default.conf.bak

# 创建必要的目录并设置权限
RUN mkdir -p /var/cache/nginx/client_temp \
             /var/cache/nginx/proxy_temp \
             /var/cache/nginx/fastcgi_temp \
             /var/cache/nginx/uwsgi_temp \
             /var/cache/nginx/scgi_temp && \
    chown -R nginx-user:nginx-user /var/cache/nginx && \
    chown -R nginx-user:nginx-user /var/log/nginx && \
    chown -R nginx-user:nginx-user /etc/nginx/conf.d && \
    touch /var/run/nginx.pid && \
    chown nginx-user:nginx-user /var/run/nginx.pid

# 暴露端口
EXPOSE 80

# 切换到非root用户
USER nginx-user

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost/ || exit 1

# 启动nginx
CMD ["nginx", "-g", "daemon off;"]