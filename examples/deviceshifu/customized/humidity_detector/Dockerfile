FROM golang:1.19.2-alpine AS builder
WORKDIR /humidity-detector
COPY *.go ./
RUN go mod init humidity-detector
COPY * ./
RUN CGO_ENABLED=0 go build -o main

FROM gcr.io/distroless/static-debian11:latest
WORKDIR /humidity-detector
COPY --from=builder /humidity-detector/main main
CMD [ "./main" ]