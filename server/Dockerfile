FROM golang:1.21.1-alpine3.18 as builder

RUN apk add git

COPY . /home/src

WORKDIR /home/src

RUN GOOS=linux GOARCH=amd64 go build -o server server.go
RUN chmod +x server

FROM alpine:3.14.2
WORKDIR /app
COPY --from=builder /home/src/server /app/

CMD [ "/app/server" ]