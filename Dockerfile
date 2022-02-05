FROM golang:1.16.3-buster as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download
# Copy the go source
COPY cmd/ cmd/
COPY pkg/ pkg/
COPY internal/ internal/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o dns-proxy cmd/main.go

FROM gcr.io/distroless/base-debian10
COPY --from=builder /workspace/dns-proxy .
CMD ["/dns-proxy"]
