FROM golang:1.23.4-alpine3.21

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY *.go ./

RUN go build -v

EXPOSE 67/udp

CMD ["/app/mygodhcpd", "-conf", "conf.yaml"]
