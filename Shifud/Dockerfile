# syntax=docker/dockerfile:1

FROM golang:1.19.2-alpine

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

# Install required packages
RUN apk add --no-cache gcc musl-dev libc-dev

COPY . .

RUN for f in plugins/*.go; do go build -buildmode=plugin -o "${f%.go}.so" "$f"; done

RUN go build -o main main.go

CMD ["./main"]
