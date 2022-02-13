FROM golang:1.17 AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go install -a -installsuffix cgo

FROM alpine:3.15
COPY --from=builder /go/bin/ /bin/
CMD ["hot-potato-discord"]