FROM golang:1.12-alpine as builder

WORKDIR /src/app
COPY  . .
#COPY .git ./

RUN apk add --no-cache \
        git \
#        tzdata \
#        ca-certificates \
        upx

#RUN go get -u github.com/semrush/zenrpc/zenrpc \
#    && go generate .

RUN GIT_COMMIT=$(git rev-list -1 HEAD --) && \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
        go build -ldflags="-X main.GitCommit=${GIT_COMMIT} -w -s" -mod vendor -o /app ./cmd

RUN upx -q /app && \
    upx -t /app

# ---

FROM alpine:3.9

WORKDIR /

RUN adduser -S -D -H -h /srv appuser
USER appuser

#COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
#COPY --from=builder /usr/share/zoneinfo/Europe/Moscow /etc/localtime
#COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

COPY --from=builder /app /app

EXPOSE 8080

CMD ["/app"]
