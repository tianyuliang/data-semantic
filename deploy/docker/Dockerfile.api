# 构建阶段
FROM golang:1.24-alpine AS builder

WORKDIR /build

# 安装依赖
RUN apk add --no-cache git ca-certificates tzdata

# 复制依赖文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 编译 API 服务
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /app/api ./api/*.go

# 运行阶段
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# 从构建阶段复制
COPY --from=builder /app/api .
COPY --from=builder /build/api/etc ./etc

# 设置时区
ENV TZ=Asia/Shanghai

EXPOSE 8888

CMD ["./api", "-f", "etc/api.yaml"]
