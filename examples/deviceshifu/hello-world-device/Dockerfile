# syntax=docker/dockerfile:1

FROM golang:1.22.0-alpine
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY *.go ./
RUN go build -o /helloworld
EXPOSE 11111
CMD [ "/helloworld" ]