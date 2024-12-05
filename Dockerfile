# 使用 Alpine 3.16 作为基础镜像
FROM alpine:3.16 as root-certs
# 替换 apk 源
RUN echo "http://dl-cdn.alpinelinux.org/alpine/v3.16/main" > /etc/apk/repositories && \
    echo "http://dl-cdn.alpinelinux.org/alpine/v3.16/community" >> /etc/apk/repositories && \
    apk update
# 安装 CA 证书
RUN apk add -U --no-cache ca-certificates
# 创建应用用户和组
RUN addgroup -g 1001 app && \
    adduser app -u 1001 -D -G app -h /home/app app

# 使用 Go 1.22 作为构建环境
FROM golang:1.22 as builder
# 设置工作目录
WORKDIR /youtube-api-files
# 复制 CA 证书
COPY --from=root-certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
# 复制应用代码
COPY . .
# 构建应用
# RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -o ./youtube-stats ./app/./...
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./youtube-stats ./app/./...

# 使用 scratch 作为最终镜像
FROM scratch as final
# 复制用户和组信息
COPY --from=root-certs /etc/passwd /etc/passwd
COPY --from=root-certs /etc/group /etc/group
# 复制 CA 证书
COPY --chown=1001:1001 --from=root-certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/


# 复制构建好的应用
COPY --chown=1001:1001 --from=builder /youtube-api-files/youtube-stats /youtube-stats
# 切换到应用用户
USER app
# 设置入口点
ENTRYPOINT [ "/youtube-stats" ]
# 暴露端口
EXPOSE 8080