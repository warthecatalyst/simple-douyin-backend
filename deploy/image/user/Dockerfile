FROM golang:1.18-alpine as builder

ENV GO111MODULE=on \
GOPROXY=https://goproxy.cn,direct

WORKDIR /simple-douyin-backend

COPY . .
RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -o gateway_service ./gateway/cmd

FROM alpine as prod

WORKDIR /simple-douyin-backend

COPY --from=builder /simple-douyin-backend/configs/config.ini ./conf/config.ini
COPY --from=builder /simple-douyin-backend/gateway_service .

EXPOSE 8060

ENTRYPOINT ["./gateway_service"]