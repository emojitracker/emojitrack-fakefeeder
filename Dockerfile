FROM golang:1.14-alpine AS builder
WORKDIR /src/emojitrack-fakefeeder
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
	go build -ldflags "-s -w" -o fakefeeder ./cmd/fakefeeder

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /src/emojitrack-fakefeeder/fakefeeder .
ENTRYPOINT ["/app/fakefeeder"]
