#ARG GO_VERSION=1.14.3
FROM golang:alpine as StoreX
# Set necessary environmet variables needed for our image
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64


WORKDIR /server
COPY go.mod go.sum ./
RUN go mod download

ADD . .
RUN go build -o bin/StoreX cmd/main.go


FROM alpine:latest

WORKDIR /

COPY --from=StoreX /server/bin .
COPY --from=StoreX /server/database/migrations ./database/migrations

EXPOSE 8080
ENTRYPOINT ["./StoreX"]