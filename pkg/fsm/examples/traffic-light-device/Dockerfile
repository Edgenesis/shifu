# syntax=docker/dockerfile:1

FROM golang:1.20.2-alpine
WORKDIR /app
ENV GO111MODULE=on
ENV GOPRIVATE=github.com/Edgenesis
ENV GOPROXY=https://goproxy.cn,direct
COPY go.mod ./
COPY go.sum ./
COPY pkg/deviceshifu pkg/deviceshifu
COPY pkg/logger pkg/logger
RUN go mod download
COPY pkg/fsm/examples/traffic-light-device/*.go ./
RUN go build -o /trafficlight
EXPOSE 11111
CMD [ "/trafficlight" ]