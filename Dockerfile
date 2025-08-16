FROM golang:1.24-alpine AS builder
RUN apk add --no-cache git ca-certificates tzdata && update-ca-certificates
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG TARGETOS TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} \
    go build -trimpath -buildvcs=false \
    -ldflags="-s -w -extldflags '-static'" \
    -o /out/app ./cmd/api

FROM scratch
ENV TZ=UTC \
    PROCESSOR_DEFAULT_URL=http://payment-processor-default:8080 \
    PROCESSOR_FALLBACK_URL=http://payment-processor-fallback:8080
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /usr/share/zoneinfo/UTC /etc/localtime
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /out/app /app
EXPOSE 8080
ENTRYPOINT ["/app"]
