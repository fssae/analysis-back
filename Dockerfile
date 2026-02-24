# # 阶段 1: 编译环境
# FROM golang:1.23 AS builder

# # 配置环境变量
# ENV GO111MODULE=on \
#     GOPROXY=https://goproxy.cn,direct \
#     CGO_ENABLED=0 \
#     GOOS=linux \
#     GOARCH=amd64

# WORKDIR /build

# # 预先下载依赖，利用 Docker 缓存层
# COPY go.mod go.sum ./
# RUN go mod download

# # 复制源码并编译
# COPY . .
# RUN go build -ldflags="-s -w" -o classroom-analysis .

# # 阶段 2: 运行环境
# # 使用 alpine (约 5MB) 或 distroless 镜像
# FROM alpine:latest

# # 安装基础证书（如果涉及 HTTPS 请求）和时区数据
# RUN apk --no-cache add ca-certificates tzdata

# WORKDIR /app

# # 从 builder 阶段复制二进制文件
# COPY --from=builder /build/classroom-analysis .
# # 复制配置文件 (根据实际目录结构调整)
# COPY --from=builder /build/config ./config

# # 暴露端口
# EXPOSE 8080

# # 运行
# CMD ["./classroom-analysis"]