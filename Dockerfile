# Build the manager binary
FROM golang:1.17 as builder

WORKDIR /workspace

COPY go.mod go.mod
COPY go.sum go.sum
COPY main.go main.go

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o cluster-wh main.go

FROM gcr.io/distroless/static:nonroot 
WORKDIR /
COPY --from=builder /workspace/cluster-wh .
USER nonroot:nonroot

ENTRYPOINT ["/cluster-wh"]
