FROM hub.hzbxhd.com/library/alpine:latest

# 安装 tzdata 并设置时区
RUN apk add --no-cache tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

WORKDIR /go/src/anesthesia
COPY cmdb . 
CMD ["./cmdb"]

