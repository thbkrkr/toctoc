FROM golang:1.9.1-alpine
ENV GOBIN=/go/bin/ GOPATH=/go
WORKDIR /go/src/github.com/thbkrkr/toctoc
COPY . /go/src/github.com/thbkrkr/toctoc
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo .

FROM alpine:3.6
RUN apk --no-cache add ca-certificates
COPY _static /_static
COPY --from=0 /go/src/github.com/thbkrkr/toctoc/toctoc /toctoc
ENTRYPOINT ["/toctoc"]