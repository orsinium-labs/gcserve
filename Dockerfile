FROM golang:1.16
WORKDIR /opt/
COPY main.go .
COPY go.mod .
COPY go.sum .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o gcserve.bin .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=0 /opt/gcserve.bin .
ENTRYPOINT ["./gcserve.bin", "--cred", "/mnt/cred.json"]
