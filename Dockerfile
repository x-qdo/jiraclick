FROM golang:1.21 AS builder

ARG GOPROXY=https://proxy.golang.org,direct
ENV GOPROXY ${GOPROXY}

ARG GOPRIVATE
ENV GOPRIVATE ${GOPRIVATE}

# dependencies caching
COPY go.mod .
COPY go.sum .

RUN GOPATH=/tmp GOPROXY=${GOPROXY} GOPRIVATE=${GOPRIVATE} go mod download

COPY . .

RUN GOPATH=/tmp CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app
RUN adduser --disabled-password --gecos '' appuser

FROM scratch

COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /app /
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY config.example.yaml /config.yaml

USER appuser

CMD ["/app"]
