FROM golang:1.25-alpine3.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mygodhcpd

FROM alpine:3.22

RUN addgroup -g 1001 -S dhcp && \
    adduser -S -D -H -u 1001 -h /app -s /sbin/nologin -G dhcp -g dhcp dhcp

WORKDIR /app

COPY --from=builder /app/mygodhcpd .

RUN chown dhcp:dhcp /app/mygodhcpd && \
    chown dhcp:dhcp /app && \
    mkdir -p /var/lib/golang-dhcpd && \
    chown dhcp:dhcp /var/lib/golang-dhcpd

EXPOSE 67/udp

CMD ["./mygodhcpd", "-conf", "conf.yaml"]
