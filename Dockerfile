FROM golang AS builder
ENV GO111MODULE=on
ENV GOOS=linux
ENV GOARCH=$GOARCH
ENV CGO_ENABLED=0
WORKDIR /build
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build -o s3web .

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /build/s3web /
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
ENTRYPOINT ["/s3web"]