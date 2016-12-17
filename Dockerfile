FROM alpine:3.4

COPY toctoc /toctoc

ENTRYPOINT ["/toctoc"]