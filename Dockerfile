FROM golang:1.13.0-alpine AS builder

#ENV GOFLAGS="-mod=readonly"

RUN apk add --update --no-cache ca-certificates make
RUN mkdir -p /workspace

WORKDIR /workspace

ARG GOPROXY

COPY go.* /workspace/
RUN go mod download

COPY . /workspace
ARG BUILD_TARGET

RUN mkdir -p build; \
    go build -o build/kube-service-annotate; \
    mv build/ /build; \
    ls /build;

FROM alpine:3.10.1

RUN apk add --update --no-cache ca-certificates tzdata bash curl
COPY --from=builder /build/* /usr/local/bin/

EXPOSE 8080
CMD ["kube-service-annotate"]