FROM golang:1.21.3 AS builder

WORKDIR /app

COPY go.mod ./
COPY cmd/balancer/* ./

RUN go mod download

ARG LDFLAGS
RUN CGO_ENABLED=0 go build \
        -ldflags "$LDFLAGS" \
        -o lb .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root
COPY --from=builder /app/lb .
ENTRYPOINT [ "/root/lb" ]