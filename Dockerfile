FROM golang:alpine AS build

WORKDIR /simyo
ENV GO111MODULE=on
#ENV GOPROXY=https://goproxy.cn,direct

# 添加项目路径
ADD ./ /simyo


RUN go mod download &&  \
    GOOS=linux CGO_ENABLED=0 GOARCH=amd64 go build -ldflags="-s -w" -installsuffix cgo -v  -o app /simyo

# 第二阶段,小镜像
FROM alpine AS prod

WORKDIR /simyo

# 复制html到镜像
COPY --from=build /simyo/index.html /simyo/

# 复制二进制到镜像
COPY --from=build /simyo/app /simyo/

# 声明端口
EXPOSE 80

# 入口命令
ENTRYPOINT ["/simyo/app"]



