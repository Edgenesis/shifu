# syntax=docker/dockerfile:1

FROM golang:1.19.2-alpine
WORKDIR /app
COPY *.go ./
RUN go mod init humidity-detector
COPY * ./
RUN go build -o /humidity-detector
EXPOSE 11111
CMD [ "/humidity-detector" ]