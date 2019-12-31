# build dashboard
FROM golang:1.13

ENV GO111MODULE=on

COPY ./ /go/src/github.com/huajiao-tv/dashboard

WORKDIR /go/src/github.com/huajiao-tv/dashboard

RUN go build

# build ui
FROM node:11

COPY ./views /data/views

WORKDIR /data/views

RUN npm install -g cnpm --registry=https://registry.npm.taobao.org
RUN rm -rf node_modules && cnpm install && cnpm run build

# image
FROM alpine:3.9

# replace update source
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

# install nginx
RUN apk update && apk add nginx && mkdir -p /run/nginx/

# dashboard api
COPY --from=0 /go/src/github.com/huajiao-tv/dashboard/dashboard /data/dashboard/bin/dashboard

# dashboard ui
COPY --from=1 /data/views/dist /var/lib/nginx/html

# nginx conf
COPY deploy/nginx/default.conf /etc/nginx/conf.d/default.conf

# dashboard conf
COPY ./config.yaml /etcd/dashboard.yaml

COPY ./deploy/run.sh /run.sh

EXPOSE 80

CMD ["sh", "/run.sh"]
