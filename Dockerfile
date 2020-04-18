FROM golang:1.14 as builder

# this module knows itself as internal.com/chat-client, so switch to that directory
WORKDIR /go/src/internal.com/chat-client/

# copy all of the current files/directories into the builder
COPY . .

# build/install the binary
RUN export GO111MODULE=auto && \
    go mod init && \
    go mod tidy && \
    CGO_ENABLED=0 GOOS=linux go install -a -tags netgo -ldflags="-s -w" ./...

# copy the binary into its own image
FROM scratch
COPY --from=builder /go/bin/chat-client .
COPY ./web /web
CMD ["/chat-client"]
