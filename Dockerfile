# syntax-docker/dockerfile:1

FROM golang:1.22-alpine

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN go build -o book-tickets ./pkg
CMD ["/build/book-tickets"]
