# syntax=docker/dockerfile:1

FROM --platform=${BUILDPLATFORM} golang:1.19-alpine as builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

RUN apk add build-base

WORKDIR /src

COPY . /src

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /launcher 

FROM --platform=${TARGETPLATFORM} alpine:latest as runner

COPY --from=builder /launcher ./

CMD ["./launcher"]