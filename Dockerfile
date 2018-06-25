FROM golang:1.10 AS builder

RUN curl -fsSL -o /usr/local/bin/dep https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64 && chmod +x /usr/local/bin/dep
RUN mkdir -p /go/src/github.com/emojitracker/emojitrack-fakefeeder
WORKDIR /go/src/github.com/emojitracker/emojitrack-fakefeeder

COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure -vendor-only

COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
	go build -ldflags "-s" -a -installsuffix cgo -o emojitrack-fakefeeder

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=0 /go/src/github.com/emojitracker/emojitrack-fakefeeder/emojitrack-fakefeeder .
ENTRYPOINT ["./emojitrack-fakefeeder"]
