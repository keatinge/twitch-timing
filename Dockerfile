FROM golang:1.13 as builder
COPY . /app

WORKDIR /app
RUN go get -d . && CGO_ENABLED=0 GOOS=linux go build -o /server twitch_api_calcs.go server.go
FROM alpine:3
RUN apk add --no-cache tzdata ca-certificates

COPY --from=builder /server /server
COPY --from=builder /app/srv /srv

CMD ["/server"]
